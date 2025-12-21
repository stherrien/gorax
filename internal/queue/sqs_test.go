package queue

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQSClient_HealthCheck(t *testing.T) {
	// This test requires AWS SQS or LocalStack - skip for unit tests
	if testing.Short() {
		t.Skip("skipping integration test - requires AWS SQS or LocalStack")
	}

	// This is an integration test that would require real SQS connection
	// The actual health check logic is tested in TestSQSClient_HealthCheck_WithNilClient
	t.Skip("integration test requires LocalStack or real AWS SQS")
}

func TestSQSClient_HealthCheck_WithNilClient(t *testing.T) {
	// Test that health check handles nil client gracefully
	client := &SQSClient{
		queueURL: "http://localhost:4566/000000000000/test-queue",
		client:   nil,
	}

	err := client.HealthCheck(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SQS client not initialized")
}
