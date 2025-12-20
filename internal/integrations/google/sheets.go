package google

import (
	"context"
	"fmt"

	"github.com/gorax/gorax/internal/credential"
	"github.com/gorax/gorax/internal/executor/actions"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	sheetsScope = "https://www.googleapis.com/auth/spreadsheets"
)

// SheetsReadAction implements the Google Sheets Read action
type SheetsReadAction struct {
	credentialService credential.Service
	baseURL           string // For testing
}

// SheetsReadConfig defines the configuration for reading a range
type SheetsReadConfig struct {
	SpreadsheetID string `json:"spreadsheet_id"`
	Range         string `json:"range"`
}

// SheetsReadResult represents the result of reading a range
type SheetsReadResult struct {
	Values [][]interface{} `json:"values"`
	Range  string          `json:"range"`
}

// Validate validates the Sheets read configuration
func (c *SheetsReadConfig) Validate() error {
	if c.SpreadsheetID == "" {
		return fmt.Errorf("spreadsheet_id is required")
	}
	if c.Range == "" {
		return fmt.Errorf("range is required")
	}
	return nil
}

// NewSheetsReadAction creates a new Sheets read action
func NewSheetsReadAction(credentialService credential.Service) *SheetsReadAction {
	return &SheetsReadAction{
		credentialService: credentialService,
	}
}

// Execute implements the Action interface
func (a *SheetsReadAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	config, ok := input.Config.(SheetsReadConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected SheetsReadConfig")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	tenantID, err := extractString(input.Context, "env.tenant_id")
	if err != nil {
		return nil, fmt.Errorf("tenant_id is required in context: %w", err)
	}

	credentialID, err := extractString(input.Context, "credential_id")
	if err != nil {
		return nil, fmt.Errorf("credential_id is required in context: %w", err)
	}

	decryptedCred, err := a.credentialService.GetValue(ctx, tenantID, credentialID, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}

	token, err := createOAuth2Token(decryptedCred.Value)
	if err != nil {
		return nil, err
	}

	var sheetsService *sheets.Service
	if a.baseURL != "" {
		sheetsService, err = sheets.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(a.baseURL))
	} else {
		sheetsService, err = sheets.NewService(ctx, createOAuth2Client(ctx, token))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Sheets service: %w", err)
	}

	resp, err := sheetsService.Spreadsheets.Values.Get(config.SpreadsheetID, config.Range).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read range: %w", err)
	}

	result := &SheetsReadResult{
		Values: resp.Values,
		Range:  resp.Range,
	}

	output := actions.NewActionOutput(result)
	output.WithMetadata("range", resp.Range)
	output.WithMetadata("row_count", len(resp.Values))

	return output, nil
}

// SheetsWriteAction implements the Google Sheets Write action
type SheetsWriteAction struct {
	credentialService credential.Service
	baseURL           string
}

// SheetsWriteConfig defines the configuration for writing to a range
type SheetsWriteConfig struct {
	SpreadsheetID string          `json:"spreadsheet_id"`
	Range         string          `json:"range"`
	Values        [][]interface{} `json:"values"`
}

// SheetsWriteResult represents the result of writing to a range
type SheetsWriteResult struct {
	UpdatedRange   string `json:"updated_range"`
	UpdatedRows    int64  `json:"updated_rows"`
	UpdatedColumns int64  `json:"updated_columns"`
	UpdatedCells   int64  `json:"updated_cells"`
}

// Validate validates the Sheets write configuration
func (c *SheetsWriteConfig) Validate() error {
	if c.SpreadsheetID == "" {
		return fmt.Errorf("spreadsheet_id is required")
	}
	if c.Range == "" {
		return fmt.Errorf("range is required")
	}
	if len(c.Values) == 0 {
		return fmt.Errorf("values are required")
	}
	return nil
}

// NewSheetsWriteAction creates a new Sheets write action
func NewSheetsWriteAction(credentialService credential.Service) *SheetsWriteAction {
	return &SheetsWriteAction{
		credentialService: credentialService,
	}
}

// Execute implements the Action interface
func (a *SheetsWriteAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	config, ok := input.Config.(SheetsWriteConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected SheetsWriteConfig")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	tenantID, err := extractString(input.Context, "env.tenant_id")
	if err != nil {
		return nil, fmt.Errorf("tenant_id is required in context: %w", err)
	}

	credentialID, err := extractString(input.Context, "credential_id")
	if err != nil {
		return nil, fmt.Errorf("credential_id is required in context: %w", err)
	}

	decryptedCred, err := a.credentialService.GetValue(ctx, tenantID, credentialID, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}

	token, err := createOAuth2Token(decryptedCred.Value)
	if err != nil {
		return nil, err
	}

	var sheetsService *sheets.Service
	if a.baseURL != "" {
		sheetsService, err = sheets.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(a.baseURL))
	} else {
		sheetsService, err = sheets.NewService(ctx, createOAuth2Client(ctx, token))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Sheets service: %w", err)
	}

	valueRange := &sheets.ValueRange{
		Values: config.Values,
	}

	resp, err := sheetsService.Spreadsheets.Values.Update(
		config.SpreadsheetID,
		config.Range,
		valueRange,
	).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to write range: %w", err)
	}

	result := &SheetsWriteResult{
		UpdatedRange:   resp.UpdatedRange,
		UpdatedRows:    resp.UpdatedRows,
		UpdatedColumns: resp.UpdatedColumns,
		UpdatedCells:   resp.UpdatedCells,
	}

	output := actions.NewActionOutput(result)
	output.WithMetadata("updated_range", resp.UpdatedRange)
	output.WithMetadata("updated_cells", resp.UpdatedCells)

	return output, nil
}

// SheetsAppendAction implements the Google Sheets Append action
type SheetsAppendAction struct {
	credentialService credential.Service
	baseURL           string
}

// SheetsAppendConfig defines the configuration for appending rows
type SheetsAppendConfig struct {
	SpreadsheetID string          `json:"spreadsheet_id"`
	Range         string          `json:"range"`
	Values        [][]interface{} `json:"values"`
}

// Validate validates the Sheets append configuration
func (c *SheetsAppendConfig) Validate() error {
	if c.SpreadsheetID == "" {
		return fmt.Errorf("spreadsheet_id is required")
	}
	if c.Range == "" {
		return fmt.Errorf("range is required")
	}
	if len(c.Values) == 0 {
		return fmt.Errorf("values are required")
	}
	return nil
}

// NewSheetsAppendAction creates a new Sheets append action
func NewSheetsAppendAction(credentialService credential.Service) *SheetsAppendAction {
	return &SheetsAppendAction{
		credentialService: credentialService,
	}
}

// Execute implements the Action interface
func (a *SheetsAppendAction) Execute(ctx context.Context, input *actions.ActionInput) (*actions.ActionOutput, error) {
	config, ok := input.Config.(SheetsAppendConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected SheetsAppendConfig")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	tenantID, err := extractString(input.Context, "env.tenant_id")
	if err != nil {
		return nil, fmt.Errorf("tenant_id is required in context: %w", err)
	}

	credentialID, err := extractString(input.Context, "credential_id")
	if err != nil {
		return nil, fmt.Errorf("credential_id is required in context: %w", err)
	}

	decryptedCred, err := a.credentialService.GetValue(ctx, tenantID, credentialID, "system")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credential: %w", err)
	}

	token, err := createOAuth2Token(decryptedCred.Value)
	if err != nil {
		return nil, err
	}

	var sheetsService *sheets.Service
	if a.baseURL != "" {
		sheetsService, err = sheets.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(a.baseURL))
	} else {
		sheetsService, err = sheets.NewService(ctx, createOAuth2Client(ctx, token))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Sheets service: %w", err)
	}

	valueRange := &sheets.ValueRange{
		Values: config.Values,
	}

	resp, err := sheetsService.Spreadsheets.Values.Append(
		config.SpreadsheetID,
		config.Range,
		valueRange,
	).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to append rows: %w", err)
	}

	result := &SheetsWriteResult{
		UpdatedRange:   resp.Updates.UpdatedRange,
		UpdatedRows:    resp.Updates.UpdatedRows,
		UpdatedColumns: resp.Updates.UpdatedColumns,
		UpdatedCells:   resp.Updates.UpdatedCells,
	}

	output := actions.NewActionOutput(result)
	output.WithMetadata("updated_range", resp.Updates.UpdatedRange)
	output.WithMetadata("updated_cells", resp.Updates.UpdatedCells)

	return output, nil
}
