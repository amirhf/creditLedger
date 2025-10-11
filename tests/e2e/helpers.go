package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// APIClient is an HTTP client for the Gateway REST API
type APIClient struct {
	BaseURL string
	Client  *http.Client
}

// NewAPIClient creates a new API client
func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// AccountResponse represents the response from creating an account
type AccountResponse struct {
	AccountID string `json:"account_id"`
	Currency  string `json:"currency"`
	CreatedAt string `json:"created_at"`
}

// TransferResponse represents the response from creating a transfer
type TransferResponse struct {
	TransferID string `json:"transfer_id"`
	Status     string `json:"status"`
	EntryID    string `json:"entry_id"`
	CreatedAt  string `json:"created_at"`
}

// BalanceResponse represents the response from querying balance
type BalanceResponse struct {
	AccountID    string `json:"account_id"`
	BalanceMinor int64  `json:"balance_minor"`
	Currency     string `json:"currency"`
	UpdatedAt    string `json:"updated_at"`
}

// StatementEntry represents a single statement entry
type StatementEntry struct {
	EntryID     string `json:"entry_id"`
	AmountMinor int64  `json:"amount_minor"`
	Side        string `json:"side"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

// StatementsResponse represents the response from querying statements
type StatementsResponse struct {
	Statements []StatementEntry `json:"statements"`
}

// CreateAccount creates a new account
func (c *APIClient) CreateAccount(currency string) (string, error) {
	body := map[string]string{
		"currency": currency,
	}

	respBody, err := c.doRequest("POST", "/accounts", body)
	if err != nil {
		return "", err
	}

	var resp AccountResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp.AccountID, nil
}

// CreateTransfer creates a new transfer
func (c *APIClient) CreateTransfer(fromAccountID, toAccountID string, amountMinor int64, currency, idempotencyKey string) (string, error) {
	body := map[string]interface{}{
		"from_account_id": fromAccountID,
		"to_account_id":   toAccountID,
		"amount_minor":    amountMinor,
		"currency":        currency,
		"idempotency_key": idempotencyKey,
	}

	respBody, err := c.doRequest("POST", "/transfers", body)
	if err != nil {
		return "", err
	}

	var resp TransferResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp.TransferID, nil
}

// GetBalance retrieves the balance for an account
func (c *APIClient) GetBalance(accountID string) (int64, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/accounts/%s/balance", accountID), nil)
	if err != nil {
		return 0, err
	}

	var resp BalanceResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp.BalanceMinor, nil
}

// GetTransfer retrieves transfer details
func (c *APIClient) GetTransfer(transferID string) (*TransferResponse, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/transfers/%s", transferID), nil)
	if err != nil {
		return nil, err
	}

	var resp TransferResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// GetStatements retrieves statements for an account
func (c *APIClient) GetStatements(accountID string) ([]StatementEntry, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/accounts/%s/statements?limit=100", accountID), nil)
	if err != nil {
		return nil, err
	}

	var resp StatementsResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resp.Statements, nil
}

// doRequest executes an HTTP request
func (c *APIClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// WaitForBalance waits for an account balance to reach the expected value
func WaitForBalance(ctx context.Context, client *APIClient, accountID string, expected int64, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				actual, _ := client.GetBalance(accountID)
				return fmt.Errorf("timeout waiting for balance: expected %d, got %d", expected, actual)
			}

			balance, err := client.GetBalance(accountID)
			if err == nil && balance == expected {
				return nil
			}
		}
	}
}

// WaitForService waits for a service to be ready
func WaitForService(ctx context.Context, url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	httpClient := &http.Client{Timeout: 2 * time.Second}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for service: %s", url)
			}

			resp, err := httpClient.Get(url)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode < 500 {
					return nil
				}
			}
		}
	}
}

// AssertBalance asserts that an account has the expected balance
func AssertBalance(t *testing.T, client *APIClient, accountID string, expected int64) {
	t.Helper()
	balance, err := client.GetBalance(accountID)
	require.NoError(t, err, "Failed to get balance for account %s", accountID)
	assert.Equal(t, expected, balance, "Balance mismatch for account %s", accountID)
}

// AssertTransferStatus asserts that a transfer has the expected status
func AssertTransferStatus(t *testing.T, client *APIClient, transferID string, expected string) {
	t.Helper()
	transfer, err := client.GetTransfer(transferID)
	require.NoError(t, err, "Failed to get transfer %s", transferID)
	assert.Equal(t, expected, transfer.Status, "Transfer status mismatch")
}

// AssertStatementCount asserts that an account has the expected number of statement entries
func AssertStatementCount(t *testing.T, client *APIClient, accountID string, expected int) {
	t.Helper()
	statements, err := client.GetStatements(accountID)
	require.NoError(t, err, "Failed to get statements for account %s", accountID)
	assert.Len(t, statements, expected, "Statement count mismatch for account %s", accountID)
}
