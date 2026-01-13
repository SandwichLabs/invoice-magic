// Package sheets provides a Google Sheets API client with rate limiting.
package sheets

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"golang.org/x/time/rate"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// sheetNamePattern validates sheet names to prevent injection.
var sheetNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_\- ]{1,100}$`)

// ValidateSheetName checks if a sheet name is safe for use in API calls.
func ValidateSheetName(name string) error {
	if name == "" {
		return fmt.Errorf("sheet name cannot be empty")
	}
	if !sheetNamePattern.MatchString(name) {
		return fmt.Errorf("sheet name contains invalid characters or is too long: %s", name)
	}
	return nil
}

// Client wraps the Google Sheets API with rate limiting.
type Client struct {
	service *sheets.Service
	limiter *rate.Limiter
}

// NewClient creates a new Sheets client using the provided authenticated HTTP client.
func NewClient(ctx context.Context, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		return nil, fmt.Errorf("http client cannot be nil")
	}

	service, err := sheets.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create sheets service: %w", err)
	}

	// Google Sheets API limit: 300 requests per minute per project
	// Use 250/min (4.16/sec) to leave headroom
	limiter := rate.NewLimiter(rate.Every(time.Minute/250), 10)

	return &Client{
		service: service,
		limiter: limiter,
	}, nil
}

// Spreadsheet represents basic spreadsheet metadata.
type Spreadsheet struct {
	ID     string
	Title  string
	Sheets []Sheet
}

// Sheet represents a single sheet (tab) within a spreadsheet.
type Sheet struct {
	ID    int64
	Title string
}

// GetSpreadsheet retrieves spreadsheet metadata including sheet names.
func (c *Client) GetSpreadsheet(ctx context.Context, spreadsheetID string) (*Spreadsheet, error) {
	if err := c.wait(ctx); err != nil {
		return nil, err
	}

	resp, err := c.service.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get spreadsheet: %w", err)
	}

	ss := &Spreadsheet{
		ID:     resp.SpreadsheetId,
		Title:  resp.Properties.Title,
		Sheets: make([]Sheet, len(resp.Sheets)),
	}

	for i, sheet := range resp.Sheets {
		ss.Sheets[i] = Sheet{
			ID:    sheet.Properties.SheetId,
			Title: sheet.Properties.Title,
		}
	}

	return ss, nil
}

// GetSheetData retrieves all data from a named sheet.
func (c *Client) GetSheetData(ctx context.Context, spreadsheetID, sheetName string) ([][]interface{}, error) {
	if err := ValidateSheetName(sheetName); err != nil {
		return nil, err
	}

	if err := c.wait(ctx); err != nil {
		return nil, err
	}

	resp, err := c.service.Spreadsheets.Values.Get(spreadsheetID, sheetName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get sheet data: %w", err)
	}

	return resp.Values, nil
}

// GetRange retrieves data from a specific range (e.g., "Sheet1!A1:D10").
func (c *Client) GetRange(ctx context.Context, spreadsheetID, rangeA1 string) ([][]interface{}, error) {
	if err := c.wait(ctx); err != nil {
		return nil, err
	}

	resp, err := c.service.Spreadsheets.Values.Get(spreadsheetID, rangeA1).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get range: %w", err)
	}

	return resp.Values, nil
}

// FindSheetByName finds a sheet by name within a spreadsheet.
func (c *Client) FindSheetByName(ctx context.Context, spreadsheetID, sheetName string) (*Sheet, error) {
	if err := ValidateSheetName(sheetName); err != nil {
		return nil, err
	}

	ss, err := c.GetSpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return nil, err
	}

	for _, sheet := range ss.Sheets {
		if sheet.Title == sheetName {
			return &sheet, nil
		}
	}

	return nil, nil
}

// wait blocks until rate limit allows the request.
func (c *Client) wait(ctx context.Context) error {
	return c.limiter.Wait(ctx)
}
