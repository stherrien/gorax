package websocket

import (
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestHubBroadcasterExecutionStarted(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	hub := NewHub(logger)
	go hub.Run()

	broadcaster := NewHubBroadcaster(hub)

	// Create test client subscribed to execution room
	client := &Client{
		ID:            "test-client",
		TenantID:      "tenant-1",
		Hub:           hub,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	hub.Register <- client
	time.Sleep(10 * time.Millisecond)

	executionID := "exec-123"
	hub.SubscribeClient(client, "execution:"+executionID)

	// Broadcast execution started
	broadcaster.BroadcastExecutionStarted("tenant-1", "workflow-1", executionID, 5)

	// Wait and receive message
	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-client.Send:
		var event ExecutionEvent
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Fatalf("Failed to unmarshal event: %v", err)
		}

		if event.Type != EventTypeExecutionStarted {
			t.Errorf("Expected type %s, got %s", EventTypeExecutionStarted, event.Type)
		}

		if event.ExecutionID != executionID {
			t.Errorf("Expected execution_id %s, got %s", executionID, event.ExecutionID)
		}

		if event.Status != "running" {
			t.Errorf("Expected status 'running', got %s", event.Status)
		}

		if event.Progress == nil {
			t.Fatal("Progress should not be nil")
		}

		if event.Progress.TotalSteps != 5 {
			t.Errorf("Expected total_steps 5, got %d", event.Progress.TotalSteps)
		}

		if event.Progress.CompletedSteps != 0 {
			t.Errorf("Expected completed_steps 0, got %d", event.Progress.CompletedSteps)
		}

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Did not receive message")
	}
}

func TestHubBroadcasterExecutionCompleted(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	hub := NewHub(logger)
	go hub.Run()

	broadcaster := NewHubBroadcaster(hub)

	client := &Client{
		ID:            "test-client",
		TenantID:      "tenant-1",
		Hub:           hub,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	hub.Register <- client
	time.Sleep(10 * time.Millisecond)

	executionID := "exec-123"
	hub.SubscribeClient(client, "execution:"+executionID)

	// Broadcast execution completed
	output := json.RawMessage(`{"result":"success","data":{"count":42}}`)
	broadcaster.BroadcastExecutionCompleted("tenant-1", "workflow-1", executionID, output)

	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-client.Send:
		var event ExecutionEvent
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Fatalf("Failed to unmarshal event: %v", err)
		}

		if event.Type != EventTypeExecutionCompleted {
			t.Errorf("Expected type %s, got %s", EventTypeExecutionCompleted, event.Type)
		}

		if event.Status != "completed" {
			t.Errorf("Expected status 'completed', got %s", event.Status)
		}

		if event.Output == nil {
			t.Fatal("Output should not be nil")
		}

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Did not receive message")
	}
}

func TestHubBroadcasterExecutionFailed(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	hub := NewHub(logger)
	go hub.Run()

	broadcaster := NewHubBroadcaster(hub)

	client := &Client{
		ID:            "test-client",
		TenantID:      "tenant-1",
		Hub:           hub,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	hub.Register <- client
	time.Sleep(10 * time.Millisecond)

	executionID := "exec-123"
	hub.SubscribeClient(client, "execution:"+executionID)

	// Broadcast execution failed
	errorMsg := "Node http-request failed: connection timeout"
	broadcaster.BroadcastExecutionFailed("tenant-1", "workflow-1", executionID, errorMsg)

	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-client.Send:
		var event ExecutionEvent
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Fatalf("Failed to unmarshal event: %v", err)
		}

		if event.Type != EventTypeExecutionFailed {
			t.Errorf("Expected type %s, got %s", EventTypeExecutionFailed, event.Type)
		}

		if event.Status != "failed" {
			t.Errorf("Expected status 'failed', got %s", event.Status)
		}

		if event.Error == nil || *event.Error != errorMsg {
			t.Errorf("Expected error '%s', got '%v'", errorMsg, event.Error)
		}

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Did not receive message")
	}
}

func TestHubBroadcasterStepStarted(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	hub := NewHub(logger)
	go hub.Run()

	broadcaster := NewHubBroadcaster(hub)

	client := &Client{
		ID:            "test-client",
		TenantID:      "tenant-1",
		Hub:           hub,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	hub.Register <- client
	time.Sleep(10 * time.Millisecond)

	executionID := "exec-123"
	hub.SubscribeClient(client, "execution:"+executionID)

	// Broadcast step started
	broadcaster.BroadcastStepStarted("tenant-1", "workflow-1", executionID, "node-1", "action:http")

	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-client.Send:
		var event ExecutionEvent
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Fatalf("Failed to unmarshal event: %v", err)
		}

		if event.Type != EventTypeStepStarted {
			t.Errorf("Expected type %s, got %s", EventTypeStepStarted, event.Type)
		}

		if event.Step == nil {
			t.Fatal("Step should not be nil")
		}

		if event.Step.NodeID != "node-1" {
			t.Errorf("Expected node_id 'node-1', got %s", event.Step.NodeID)
		}

		if event.Step.NodeType != "action:http" {
			t.Errorf("Expected node_type 'action:http', got %s", event.Step.NodeType)
		}

		if event.Step.Status != "running" {
			t.Errorf("Expected status 'running', got %s", event.Step.Status)
		}

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Did not receive message")
	}
}

func TestHubBroadcasterStepCompleted(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	hub := NewHub(logger)
	go hub.Run()

	broadcaster := NewHubBroadcaster(hub)

	client := &Client{
		ID:            "test-client",
		TenantID:      "tenant-1",
		Hub:           hub,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	hub.Register <- client
	time.Sleep(10 * time.Millisecond)

	executionID := "exec-123"
	hub.SubscribeClient(client, "execution:"+executionID)

	// Broadcast step completed
	output := json.RawMessage(`{"statusCode":200,"body":"OK"}`)
	broadcaster.BroadcastStepCompleted("tenant-1", "workflow-1", executionID, "node-1", output, 150)

	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-client.Send:
		var event ExecutionEvent
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Fatalf("Failed to unmarshal event: %v", err)
		}

		if event.Type != EventTypeStepCompleted {
			t.Errorf("Expected type %s, got %s", EventTypeStepCompleted, event.Type)
		}

		if event.Step == nil {
			t.Fatal("Step should not be nil")
		}

		if event.Step.Status != "completed" {
			t.Errorf("Expected status 'completed', got %s", event.Step.Status)
		}

		if event.Step.DurationMs == nil || *event.Step.DurationMs != 150 {
			t.Errorf("Expected duration_ms 150, got %v", event.Step.DurationMs)
		}

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Did not receive message")
	}
}

func TestHubBroadcasterProgress(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	hub := NewHub(logger)
	go hub.Run()

	broadcaster := NewHubBroadcaster(hub)

	client := &Client{
		ID:            "test-client",
		TenantID:      "tenant-1",
		Hub:           hub,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	hub.Register <- client
	time.Sleep(10 * time.Millisecond)

	executionID := "exec-123"
	hub.SubscribeClient(client, "execution:"+executionID)

	// Broadcast progress
	broadcaster.BroadcastProgress("tenant-1", "workflow-1", executionID, 3, 5)

	time.Sleep(50 * time.Millisecond)

	select {
	case msg := <-client.Send:
		var event ExecutionEvent
		if err := json.Unmarshal(msg, &event); err != nil {
			t.Fatalf("Failed to unmarshal event: %v", err)
		}

		if event.Type != EventTypeExecutionProgress {
			t.Errorf("Expected type %s, got %s", EventTypeExecutionProgress, event.Type)
		}

		if event.Progress == nil {
			t.Fatal("Progress should not be nil")
		}

		if event.Progress.CompletedSteps != 3 {
			t.Errorf("Expected completed_steps 3, got %d", event.Progress.CompletedSteps)
		}

		if event.Progress.TotalSteps != 5 {
			t.Errorf("Expected total_steps 5, got %d", event.Progress.TotalSteps)
		}

		expectedPercentage := 60.0
		if event.Progress.Percentage != expectedPercentage {
			t.Errorf("Expected percentage %.1f, got %.1f", expectedPercentage, event.Progress.Percentage)
		}

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Did not receive message")
	}
}

func TestBroadcastToMultipleRooms(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	hub := NewHub(logger)
	go hub.Run()

	broadcaster := NewHubBroadcaster(hub)

	// Create client subscribed to execution, workflow, and tenant rooms
	client := &Client{
		ID:            "test-client",
		TenantID:      "tenant-1",
		Hub:           hub,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	hub.Register <- client
	time.Sleep(10 * time.Millisecond)

	executionID := "exec-123"
	workflowID := "workflow-1"
	tenantID := "tenant-1"

	hub.SubscribeClient(client, "execution:"+executionID)
	hub.SubscribeClient(client, "workflow:"+workflowID)
	hub.SubscribeClient(client, "tenant:"+tenantID)

	// Broadcast execution started (should go to all 3 rooms)
	broadcaster.BroadcastExecutionStarted(tenantID, workflowID, executionID, 5)

	time.Sleep(50 * time.Millisecond)

	// Should receive 3 copies of the message (one per room subscription)
	messageCount := 0
	for i := 0; i < 3; i++ {
		select {
		case <-client.Send:
			messageCount++
		case <-time.After(100 * time.Millisecond):
			break
		}
	}

	if messageCount != 3 {
		t.Errorf("Expected 3 messages (one per room), got %d", messageCount)
	}
}
