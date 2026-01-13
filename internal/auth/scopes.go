package auth

import "golang.org/x/oauth2"

// Google API OAuth2 scopes.
const (
	ScopeSheetsReadonly = "https://www.googleapis.com/auth/spreadsheets.readonly"
	ScopeSpreadsheets   = "https://www.googleapis.com/auth/spreadsheets"
)

// Feature names for RequiredScopes.
const (
	FeatureSheetsReadonly = "sheets-readonly"
	FeatureSheets         = "sheets"
)

// DefaultScopes returns scopes needed for reading Google Sheets.
func DefaultScopes() []string {
	return []string{ScopeSheetsReadonly}
}

// AllScopes returns all supported scopes.
func AllScopes() []string {
	return []string{ScopeSpreadsheets}
}

// RequiredScopes returns the scopes needed for the specified features.
func RequiredScopes(features ...string) []string {
	scopeSet := make(map[string]bool)

	for _, feature := range features {
		switch feature {
		case FeatureSheetsReadonly:
			scopeSet[ScopeSheetsReadonly] = true
		case FeatureSheets:
			scopeSet[ScopeSpreadsheets] = true
		}
	}

	scopes := make([]string, 0, len(scopeSet))
	for scope := range scopeSet {
		scopes = append(scopes, scope)
	}
	return scopes
}

// TokenHasScope checks if a token was granted with a specific scope.
func TokenHasScope(meta *TokenMetadata, scope string) bool {
	if meta == nil {
		return false
	}
	for _, s := range meta.GrantedScopes {
		if s == scope {
			return true
		}
	}
	return false
}

// TokenHasAllScopes checks if a token has all the required scopes.
func TokenHasAllScopes(meta *TokenMetadata, required []string) bool {
	if meta == nil {
		return false
	}

	grantedSet := make(map[string]bool)
	for _, s := range meta.GrantedScopes {
		grantedSet[s] = true
	}

	for _, req := range required {
		if !grantedSet[req] {
			return false
		}
	}
	return true
}

// MissingScopes returns which scopes from required are not in the token.
func MissingScopes(meta *TokenMetadata, required []string) []string {
	if meta == nil {
		return required
	}

	grantedSet := make(map[string]bool)
	for _, s := range meta.GrantedScopes {
		grantedSet[s] = true
	}

	var missing []string
	for _, req := range required {
		if !grantedSet[req] {
			missing = append(missing, req)
		}
	}
	return missing
}

// ScopesToFeatures converts scope URLs to human-readable feature names.
func ScopesToFeatures(scopes []string) []string {
	var features []string
	for _, scope := range scopes {
		switch scope {
		case ScopeSheetsReadonly:
			features = append(features, "Google Sheets (read-only)")
		case ScopeSpreadsheets:
			features = append(features, "Google Sheets (read/write)")
		default:
			features = append(features, scope)
		}
	}
	return features
}

// TokenMetadata stores additional information about the OAuth token.
type TokenMetadata struct {
	GrantedScopes []string `json:"granted_scopes"`
}

// ExtendedToken wraps oauth2.Token with additional metadata.
type ExtendedToken struct {
	*oauth2.Token
	Metadata TokenMetadata `json:"metadata"`
}
