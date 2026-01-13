// Package auth provides Google OAuth2 authentication for Google Sheets access.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Authenticator handles Google OAuth2 authentication.
type Authenticator struct {
	credentialsFile string
	tokenFile       string
	scopes          []string
	mu              sync.Mutex
}

// NewAuthenticator creates a new Authenticator with the specified credentials and token file paths.
func NewAuthenticator(credentialsFile, tokenFile string, scopes []string) (*Authenticator, error) {
	if credentialsFile == "" {
		return nil, fmt.Errorf("credentials file path cannot be empty")
	}
	if tokenFile == "" {
		return nil, fmt.Errorf("token file path cannot be empty")
	}
	if len(scopes) == 0 {
		scopes = DefaultScopes()
	}

	return &Authenticator{
		credentialsFile: credentialsFile,
		tokenFile:       tokenFile,
		scopes:          scopes,
	}, nil
}

// GetClient returns an authenticated HTTP client for making Google API requests.
func (a *Authenticator) GetClient(ctx context.Context) (*http.Client, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	config, err := a.loadCredentials()
	if err != nil {
		return nil, err
	}

	extToken, err := a.loadExtendedToken()
	if err != nil {
		extToken, err = a.getTokenFromWeb(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("get token from web: %w", err)
		}

		if err := a.saveExtendedToken(extToken); err != nil {
			return nil, fmt.Errorf("save token: %w", err)
		}
	} else {
		missing := MissingScopes(&extToken.Metadata, a.scopes)
		if len(missing) > 0 {
			return nil, &ScopeUpgradeRequired{
				MissingScopes: missing,
				CurrentScopes: extToken.Metadata.GrantedScopes,
			}
		}
	}

	return config.Client(ctx, extToken.Token), nil
}

// GetClientWithScopeUpgrade returns an authenticated HTTP client, automatically
// re-authenticating if additional scopes are needed.
func (a *Authenticator) GetClientWithScopeUpgrade(ctx context.Context) (*http.Client, error) {
	client, err := a.GetClient(ctx)
	if err != nil {
		if scopeErr, ok := err.(*ScopeUpgradeRequired); ok {
			fmt.Printf("Additional permissions required: %v\n", ScopesToFeatures(scopeErr.MissingScopes))
			fmt.Println("Re-authentication needed to grant new permissions.")

			if clearErr := a.ClearToken(); clearErr != nil {
				return nil, fmt.Errorf("clear token for re-auth: %w", clearErr)
			}

			return a.GetClient(ctx)
		}
		return nil, err
	}
	return client, nil
}

// ClearToken deletes the cached token file.
func (a *Authenticator) ClearToken() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	err := os.Remove(a.tokenFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove token file: %w", err)
	}
	return nil
}

// NeedsAuth returns true if authentication is required.
func (a *Authenticator) NeedsAuth() bool {
	_, err := a.loadExtendedToken()
	return err != nil
}

// GetGrantedScopes returns the scopes that the current token was granted.
func (a *Authenticator) GetGrantedScopes() ([]string, error) {
	extToken, err := a.loadExtendedToken()
	if err != nil {
		return nil, err
	}
	return extToken.Metadata.GrantedScopes, nil
}

func (a *Authenticator) loadCredentials() (*oauth2.Config, error) {
	b, err := os.ReadFile(a.credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("read credentials file: %w", err)
	}

	config, err := google.ConfigFromJSON(b, a.scopes...)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	return config, nil
}

func (a *Authenticator) loadExtendedToken() (*ExtendedToken, error) {
	f, err := os.Open(a.tokenFile)
	if err != nil {
		return nil, fmt.Errorf("open token file: %w", err)
	}
	defer f.Close()

	extToken := &ExtendedToken{}
	if err := json.NewDecoder(f).Decode(extToken); err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}

	if extToken.Token == nil {
		f.Seek(0, 0)
		token := &oauth2.Token{}
		if err := json.NewDecoder(f).Decode(token); err != nil {
			return nil, fmt.Errorf("decode legacy token: %w", err)
		}
		extToken = &ExtendedToken{
			Token: token,
			Metadata: TokenMetadata{
				GrantedScopes: []string{ScopeSheetsReadonly},
			},
		}
	}

	return extToken, nil
}

func (a *Authenticator) saveExtendedToken(extToken *ExtendedToken) error {
	f, err := os.OpenFile(a.tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create token file: %w", err)
	}

	if err := json.NewEncoder(f).Encode(extToken); err != nil {
		f.Close()
		return fmt.Errorf("encode token: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("close token file: %w", err)
	}

	return nil
}

func (a *Authenticator) getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*ExtendedToken, error) {
	redirectURL := "http://localhost:8080"
	config.RedirectURL = redirectURL

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n", authURL)

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)
	serverDone := make(chan struct{})

	mux := http.NewServeMux()
	server := &http.Server{Addr: ":8080", Handler: mux}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Authorization code not found", http.StatusBadRequest)
			select {
			case errChan <- fmt.Errorf("authorization code not found in callback"):
			default:
			}
			return
		}
		fmt.Fprintf(w, "Authorization successful! You can close this window.")
		select {
		case codeChan <- code:
		default:
		}
	})

	go func() {
		defer close(serverDone)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			select {
			case errChan <- fmt.Errorf("callback server error: %w", err):
			default:
			}
		}
	}()

	shutdownServer := func(shutdownCtx context.Context) {
		server.Shutdown(shutdownCtx)
		<-serverDone
	}

	var authCode string
	select {
	case authCode = <-codeChan:
		shutdownServer(ctx)
	case err := <-errChan:
		shutdownServer(ctx)
		return nil, err
	case <-ctx.Done():
		shutdownServer(context.Background())
		return nil, ctx.Err()
	}

	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("exchange auth code: %w", err)
	}

	return &ExtendedToken{
		Token: token,
		Metadata: TokenMetadata{
			GrantedScopes: a.scopes,
		},
	}, nil
}

// ScopeUpgradeRequired is returned when the cached token doesn't have all required scopes.
type ScopeUpgradeRequired struct {
	MissingScopes []string
	CurrentScopes []string
}

func (e *ScopeUpgradeRequired) Error() string {
	return fmt.Sprintf("token missing required scopes: %v (have: %v)", e.MissingScopes, e.CurrentScopes)
}
