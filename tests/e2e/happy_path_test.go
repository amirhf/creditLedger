package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHappyPath_CreateAccountsTransferCheckBalance tests the basic workflow:
// 1. Create two accounts
// 2. Execute transfer from A to B
// 3. Wait for event processing
// 4. Verify balances are correct
// 5. Verify transfer status is COMPLETED
// 6. Verify statements have correct entries
func TestHappyPath_CreateAccountsTransferCheckBalance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Setup test environment (Testcontainers)
	t.Log("Setting up test environment...")
	env, err := setupTestEnvironment(ctx, t)
	require.NoError(t, err, "Failed to setup test environment")
	defer func() {
		t.Log("Tearing down test environment...")
		env.Teardown(ctx)
	}()

	// TODO: Start all microservices
	// services := startAllServices(ctx, env, t)
	// defer services.StopAll()

	// For now, assume services are running manually
	// Gateway should be available at http://localhost:4000
	gatewayURL := "http://localhost:4000"
	
	// Wait for Gateway to be ready
	t.Log("Waiting for Gateway to be ready...")
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	err = WaitForService(ctx, gatewayURL+"/healthz", 30*time.Second)
	if err != nil {
		t.Skipf("Gateway not available, skipping test: %v", err)
		return
	}

	client := NewAPIClient(gatewayURL)

	// Step 1: Create Account A
	t.Log("Creating Account A...")
	accountA, err := client.CreateAccount("USD")
	require.NoError(t, err, "Failed to create account A")
	require.NotEmpty(t, accountA, "Account A ID should not be empty")
	t.Logf("Account A created: %s", accountA)

	// Verify it's a valid UUID
	_, err = uuid.Parse(accountA)
	require.NoError(t, err, "Account A ID should be a valid UUID")

	// Step 2: Create Account B
	t.Log("Creating Account B...")
	accountB, err := client.CreateAccount("USD")
	require.NoError(t, err, "Failed to create account B")
	require.NotEmpty(t, accountB, "Account B ID should not be empty")
	t.Logf("Account B created: %s", accountB)

	// Verify it's a valid UUID
	_, err = uuid.Parse(accountB)
	require.NoError(t, err, "Account B ID should be a valid UUID")

	// Step 3: Execute transfer: A → B ($50.00 = 5000 minor units)
	t.Log("Executing transfer from A to B...")
	idempotencyKey := uuid.New().String()
	transferID, err := client.CreateTransfer(accountA, accountB, 5000, "USD", idempotencyKey)
	require.NoError(t, err, "Failed to create transfer")
	require.NotEmpty(t, transferID, "Transfer ID should not be empty")
	t.Logf("Transfer created: %s (idempotency: %s)", transferID, idempotencyKey)

	// Verify it's a valid UUID
	_, err = uuid.Parse(transferID)
	require.NoError(t, err, "Transfer ID should be a valid UUID")

	// Step 4: Wait for event processing (balance projection)
	t.Log("Waiting for event processing...")
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = WaitForBalance(ctx, client, accountB, 5000, 10*time.Second)
	require.NoError(t, err, "Account B balance not updated within timeout")

	// Step 5: Verify balances
	t.Log("Verifying balances...")
	AssertBalance(t, client, accountA, -5000)
	AssertBalance(t, client, accountB, 5000)
	t.Log("✓ Balances are correct")

	// Step 6: Verify transfer status
	t.Log("Verifying transfer status...")
	AssertTransferStatus(t, client, transferID, "COMPLETED")
	t.Log("✓ Transfer status is COMPLETED")

	// Step 7: Verify statements
	t.Log("Verifying statements...")
	
	statementsA, err := client.GetStatements(accountA)
	require.NoError(t, err, "Failed to get statements for account A")
	require.Len(t, statementsA, 1, "Account A should have exactly 1 statement entry")
	assert.Equal(t, "DEBIT", statementsA[0].Side, "Account A should have a DEBIT entry")
	assert.Equal(t, int64(5000), statementsA[0].AmountMinor, "Statement amount should be 5000")
	
	statementsB, err := client.GetStatements(accountB)
	require.NoError(t, err, "Failed to get statements for account B")
	require.Len(t, statementsB, 1, "Account B should have exactly 1 statement entry")
	assert.Equal(t, "CREDIT", statementsB[0].Side, "Account B should have a CREDIT entry")
	assert.Equal(t, int64(5000), statementsB[0].AmountMinor, "Statement amount should be 5000")
	
	t.Log("✓ Statements are correct")

	t.Log("✅ Happy path test completed successfully!")
}

// TestHappyPath_MultipleTransfers tests multiple transfers between accounts
func TestHappyPath_MultipleTransfers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Setup test environment
	t.Log("Setting up test environment...")
	env, err := setupTestEnvironment(ctx, t)
	require.NoError(t, err, "Failed to setup test environment")
	defer env.Teardown(ctx)

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

	// Create 3 accounts
	t.Log("Creating accounts...")
	accountA, _ := client.CreateAccount("USD")
	accountB, _ := client.CreateAccount("USD")
	accountC, _ := client.CreateAccount("USD")

	// Execute multiple transfers
	t.Log("Executing multiple transfers...")
	
	// Transfer 1: A → B (1000)
	_, err = client.CreateTransfer(accountA, accountB, 1000, "USD", uuid.New().String())
	require.NoError(t, err)
	
	// Transfer 2: B → C (500)
	_, err = client.CreateTransfer(accountB, accountC, 500, "USD", uuid.New().String())
	require.NoError(t, err)
	
	// Transfer 3: A → C (300)
	_, err = client.CreateTransfer(accountA, accountC, 300, "USD", uuid.New().String())
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(3 * time.Second)

	// Verify final balances
	t.Log("Verifying final balances...")
	AssertBalance(t, client, accountA, -1300) // -1000 - 300
	AssertBalance(t, client, accountB, 500)    // +1000 - 500
	AssertBalance(t, client, accountC, 800)    // +500 + 300

	// Verify statement counts
	AssertStatementCount(t, client, accountA, 2) // 2 debits
	AssertStatementCount(t, client, accountB, 2) // 1 credit, 1 debit
	AssertStatementCount(t, client, accountC, 2) // 2 credits

	t.Log("✅ Multiple transfers test completed successfully!")
}
