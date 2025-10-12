package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIdempotency_DuplicateTransferSingleEffect tests that duplicate transfer requests
// with the same idempotency key result in only one transfer being processed
func TestIdempotency_DuplicateTransferSingleEffect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Setup test environment
	t.Log("Setting up test environment...")
	env, err := setupTestEnvironment(ctx, t)
	require.NoError(t, err, "Failed to setup test environment")
	defer func() {
		t.Log("Tearing down test environment...")
		env.Teardown(ctx)
	}()

	// For now, assume services are running manually
	gatewayURL := "http://localhost:4000"

	// Wait for Gateway to be ready
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = WaitForService(ctx, gatewayURL+"/healthz", 30*time.Second)
	if err != nil {
		t.Skipf("Gateway not available, skipping test: %v", err)
		return
	}

	client := NewAPIClient(gatewayURL)

	// Step 1: Create two accounts
	t.Log("Creating accounts...")
	accountA, err := client.CreateAccount("USD")
	require.NoError(t, err, "Failed to create account A")
	t.Logf("Account A: %s", accountA)

	accountB, err := client.CreateAccount("USD")
	require.NoError(t, err, "Failed to create account B")
	t.Logf("Account B: %s", accountB)

	// Step 2: Execute transfer with a specific idempotency key
	t.Log("Executing first transfer request...")
	idempotencyKey := "test-idem-" + uuid.New().String()
	transferID1, err := client.CreateTransfer(accountA, accountB, 5000, "USD", idempotencyKey)
	require.NoError(t, err, "Failed to create first transfer")
	require.NotEmpty(t, transferID1, "Transfer ID should not be empty")
	t.Logf("First transfer ID: %s", transferID1)

	// Step 3: Execute SAME transfer again (duplicate request)
	t.Log("Executing duplicate transfer request (same idempotency key)...")
	transferID2, err := client.CreateTransfer(accountA, accountB, 5000, "USD", idempotencyKey)
	require.NoError(t, err, "Duplicate request should not fail")
	t.Logf("Second transfer ID: %s", transferID2)

	// Step 4: Execute SAME transfer a third time
	t.Log("Executing third duplicate request...")
	transferID3, err := client.CreateTransfer(accountA, accountB, 5000, "USD", idempotencyKey)
	require.NoError(t, err, "Third duplicate request should not fail")
	t.Logf("Third transfer ID: %s", transferID3)

	// ASSERTION 1: All requests should return the SAME transfer ID
	t.Log("Verifying all requests returned same transfer ID...")
	assert.Equal(t, transferID1, transferID2, "Second request should return same transfer ID")
	assert.Equal(t, transferID1, transferID3, "Third request should return same transfer ID")
	t.Log("✓ All requests returned identical transfer ID")

	// Step 5: Wait for event processing
	t.Log("Waiting for event processing...")
	time.Sleep(3 * time.Second)

	// ASSERTION 2: Balance should have changed only ONCE (not three times)
	t.Log("Verifying balances changed only once...")
	balanceA, err := client.GetBalance(accountA)
	require.NoError(t, err, "Failed to get account A balance")
	assert.Equal(t, int64(-5000), balanceA, "Account A should be debited once, not three times")
	t.Logf("Account A balance: %d (expected: -5000, NOT -15000)", balanceA)

	balanceB, err := client.GetBalance(accountB)
	require.NoError(t, err, "Failed to get account B balance")
	assert.Equal(t, int64(5000), balanceB, "Account B should be credited once, not three times")
	t.Logf("Account B balance: %d (expected: 5000, NOT 15000)", balanceB)
	t.Log("✓ Balances changed only once")

	// ASSERTION 3: Only ONE journal entry should exist
	t.Log("Verifying only one journal entry was created...")
	statementsB, err := client.GetStatements(accountB)
	require.NoError(t, err, "Failed to get statements for account B")
	assert.Len(t, statementsB, 1, "Should have exactly 1 statement entry, not 3")
	t.Logf("Statement count for Account B: %d (expected: 1)", len(statementsB))
	t.Log("✓ Only one journal entry created")

	// ASSERTION 4: Verify transfer status
	t.Log("Verifying transfer status...")
	transfer, err := client.GetTransfer(transferID1)
	require.NoError(t, err, "Failed to get transfer details")
	assert.Equal(t, "COMPLETED", transfer.Status, "Transfer should be completed")
	t.Log("✓ Transfer status is COMPLETED")

	t.Log("✅ Idempotency test completed successfully!")
	t.Logf("Summary:")
	t.Logf("  - 3 duplicate requests made")
	t.Logf("  - All returned same transfer ID: %s", transferID1)
	t.Logf("  - Balance changed only once: -5000 / +5000")
	t.Logf("  - Only 1 journal entry created")
}

// TestIdempotency_DifferentAmountsSameKey tests that requests with the same
// idempotency key but different amounts are handled correctly
func TestIdempotency_DifferentAmountsSameKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Setup test environment
	t.Log("Setting up test environment...")
	env, err := setupTestEnvironment(ctx, t)
	require.NoError(t, err, "Failed to setup test environment")
	defer env.Teardown(ctx)

	gatewayURL := "http://localhost:4000"

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = WaitForService(ctx, gatewayURL+"/healthz", 30*time.Second)
	if err != nil {
		t.Skipf("Gateway not available, skipping test: %v", err)
		return
	}

	client := NewAPIClient(gatewayURL)

	// Create accounts
	t.Log("Creating accounts...")
	accountA, _ := client.CreateAccount("USD")
	accountB, _ := client.CreateAccount("USD")

	// Execute transfer with amount 1000
	idempotencyKey := "test-idem-conflict-" + uuid.New().String()
	t.Log("Executing first transfer: 1000 units...")
	transferID1, err := client.CreateTransfer(accountA, accountB, 1000, "USD", idempotencyKey)
	require.NoError(t, err, "Failed to create first transfer")
	t.Logf("First transfer: %s (amount: 1000)", transferID1)

	// Try to execute transfer with DIFFERENT amount but SAME idempotency key
	// This should either:
	// 1. Return the same transfer ID (ignore the new amount)
	// 2. Return an error (conflict detected)
	t.Log("Executing second transfer with DIFFERENT amount (5000) but SAME idempotency key...")
	transferID2, err := client.CreateTransfer(accountA, accountB, 5000, "USD", idempotencyKey)
	
	// The system should either reject this or return the original transfer
	if err == nil {
		// If no error, it should return the same transfer ID
		assert.Equal(t, transferID1, transferID2, "Should return original transfer ID when idempotency key conflicts")
		t.Logf("✓ System returned original transfer ID: %s", transferID2)
		
		// Verify the amount is still the original (1000, not 5000)
		time.Sleep(2 * time.Second)
		balanceB, _ := client.GetBalance(accountB)
		assert.Equal(t, int64(1000), balanceB, "Balance should reflect original amount (1000), not new amount (5000)")
		t.Logf("✓ Balance is correct: %d (original amount)", balanceB)
	} else {
		// If error, that's also acceptable behavior (rejecting conflicting request)
		t.Logf("✓ System rejected conflicting request: %v", err)
	}

	t.Log("✅ Idempotency conflict test completed successfully!")
}

// TestIdempotency_ConcurrentRequests tests that concurrent duplicate requests
// are handled correctly (stress test)
func TestIdempotency_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Setup test environment
	t.Log("Setting up test environment...")
	env, err := setupTestEnvironment(ctx, t)
	require.NoError(t, err, "Failed to setup test environment")
	defer env.Teardown(ctx)

	gatewayURL := "http://localhost:4000"

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = WaitForService(ctx, gatewayURL+"/healthz", 30*time.Second)
	if err != nil {
		t.Skipf("Gateway not available, skipping test: %v", err)
		return
	}

	client := NewAPIClient(gatewayURL)

	// Create accounts
	t.Log("Creating accounts...")
	accountA, _ := client.CreateAccount("USD")
	accountB, _ := client.CreateAccount("USD")

	// Execute 10 CONCURRENT requests with the SAME idempotency key
	idempotencyKey := "test-idem-concurrent-" + uuid.New().String()
	t.Log("Executing 10 concurrent duplicate requests...")

	type result struct {
		transferID string
		err        error
	}
	results := make(chan result, 10)

	// Launch 10 goroutines making the same request
	for i := 0; i < 10; i++ {
		go func(idx int) {
			transferID, err := client.CreateTransfer(accountA, accountB, 3000, "USD", idempotencyKey)
			results <- result{transferID: transferID, err: err}
		}(i)
	}

	// Collect all results
	var transferIDs []string
	var errors []error
	for i := 0; i < 10; i++ {
		res := <-results
		if res.err == nil {
			transferIDs = append(transferIDs, res.transferID)
		} else {
			errors = append(errors, res.err)
		}
	}

	t.Logf("Received %d successful responses, %d errors", len(transferIDs), len(errors))

	// All successful responses should have the SAME transfer ID
	if len(transferIDs) > 0 {
		firstID := transferIDs[0]
		for i, id := range transferIDs {
			assert.Equal(t, firstID, id, "Request %d should return same transfer ID", i)
		}
		t.Logf("✓ All %d successful requests returned same transfer ID: %s", len(transferIDs), firstID)
	}

	// Wait for processing
	time.Sleep(3 * time.Second)

	// Balance should change only ONCE despite 10 concurrent requests
	balanceA, _ := client.GetBalance(accountA)
	balanceB, _ := client.GetBalance(accountB)

	assert.Equal(t, int64(-3000), balanceA, "Account A should be debited only once")
	assert.Equal(t, int64(3000), balanceB, "Account B should be credited only once")
	t.Logf("✓ Balances correct: A=%d, B=%d (changed only once despite 10 requests)", balanceA, balanceB)

	// Only one statement entry should exist
	statements, _ := client.GetStatements(accountB)
	assert.Len(t, statements, 1, "Should have exactly 1 statement entry despite 10 concurrent requests")
	t.Log("✓ Only 1 journal entry created despite concurrent requests")

	t.Log("✅ Concurrent idempotency test completed successfully!")
}
