package websocket

import (
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"
)

func TestHubRegistration(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	hub := NewHub(logger)

	// Start hub in background
	go hub.Run()

	// Create mock client
	client := &Client{
		ID:            "test-client-1",
		TenantID:      "tenant-1",
		Hub:           hub,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	// Register client
	hub.Register <- client

	// Wait a bit for processing
	time.Sleep(10 * time.Millisecond)

	// Check client is registered
	hub.mu.RLock()
	_, exists := hub.clients[client.ID]
	hub.mu.RUnlock()

	if !exists {
		t.Errorf("Client should be registered")
	}
}

func TestHubUnregistration(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	hub := NewHub(logger)

	go hub.Run()

	client := &Client{
		ID:            "test-client-1",
		TenantID:      "tenant-1",
		Hub:           hub,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	// Register then unregister
	hub.Register <- client
	time.Sleep(10 * time.Millisecond)

	hub.Unregister <- client
	time.Sleep(10 * time.Millisecond)

	// Check client is unregistered
	hub.mu.RLock()
	_, exists := hub.clients[client.ID]
	hub.mu.RUnlock()

	if exists {
		t.Errorf("Client should be unregistered")
	}
}

func TestHubSubscription(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	hub := NewHub(logger)

	go hub.Run()

	client := &Client{
		ID:            "test-client-1",
		TenantID:      "tenant-1",
		Hub:           hub,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	hub.Register <- client
	time.Sleep(10 * time.Millisecond)

	// Subscribe to room
	room := "execution:test-123"
	hub.SubscribeClient(client, room)

	// Check subscription
	hub.mu.RLock()
	roomClients, roomExists := hub.rooms[room]
	hub.mu.RUnlock()

	if !roomExists {
		t.Errorf("Room should exist")
	}

	if _, clientInRoom := roomClients[client.ID]; !clientInRoom {
		t.Errorf("Client should be in room")
	}

	client.mu.RLock()
	subscribed := client.Subscriptions[room]
	client.mu.RUnlock()

	if !subscribed {
		t.Errorf("Client should be subscribed to room")
	}
}

func TestHubBroadcast(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	hub := NewHub(logger)

	go hub.Run()

	// Create two clients
	client1 := &Client{
		ID:            "client-1",
		TenantID:      "tenant-1",
		Hub:           hub,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	client2 := &Client{
		ID:            "client-2",
		TenantID:      "tenant-1",
		Hub:           hub,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	// Register both clients
	hub.Register <- client1
	hub.Register <- client2
	time.Sleep(10 * time.Millisecond)

	// Subscribe both to same room
	room := "execution:test-123"
	hub.SubscribeClient(client1, room)
	hub.SubscribeClient(client2, room)

	// Broadcast message
	message := []byte(`{"type":"test","data":"hello"}`)
	hub.BroadcastToRoom(room, message)

	// Wait for broadcast to process
	time.Sleep(50 * time.Millisecond)

	// Check both clients received message
	var wg sync.WaitGroup
	wg.Add(2)

	checkClient := func(client *Client, name string) {
		defer wg.Done()
		select {
		case msg := <-client.Send:
			if string(msg) != string(message) {
				t.Errorf("%s received wrong message: %s", name, string(msg))
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("%s did not receive message", name)
		}
	}

	go checkClient(client1, "client1")
	go checkClient(client2, "client2")

	wg.Wait()
}

func TestHubUnsubscribe(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	hub := NewHub(logger)

	go hub.Run()

	client := &Client{
		ID:            "test-client-1",
		TenantID:      "tenant-1",
		Hub:           hub,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	hub.Register <- client
	time.Sleep(10 * time.Millisecond)

	// Subscribe then unsubscribe
	room := "execution:test-123"
	hub.SubscribeClient(client, room)
	hub.UnsubscribeClient(client, room)

	// Check unsubscription
	hub.mu.RLock()
	_, roomExists := hub.rooms[room]
	hub.mu.RUnlock()

	if roomExists {
		t.Errorf("Room should be removed when no clients")
	}

	client.mu.RLock()
	subscribed := client.Subscriptions[room]
	client.mu.RUnlock()

	if subscribed {
		t.Errorf("Client should not be subscribed to room")
	}
}

func TestHubMultipleRooms(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	hub := NewHub(logger)

	go hub.Run()

	client := &Client{
		ID:            "test-client-1",
		TenantID:      "tenant-1",
		Hub:           hub,
		Send:          make(chan []byte, 256),
		Subscriptions: make(map[string]bool),
	}

	hub.Register <- client
	time.Sleep(10 * time.Millisecond)

	// Subscribe to multiple rooms
	room1 := "execution:test-123"
	room2 := "workflow:wf-456"
	room3 := "tenant:tenant-1"

	hub.SubscribeClient(client, room1)
	hub.SubscribeClient(client, room2)
	hub.SubscribeClient(client, room3)

	// Check all subscriptions
	client.mu.RLock()
	subCount := len(client.Subscriptions)
	client.mu.RUnlock()

	if subCount != 3 {
		t.Errorf("Client should have 3 subscriptions, got %d", subCount)
	}

	// Broadcast to each room
	hub.BroadcastToRoom(room1, []byte("msg1"))
	hub.BroadcastToRoom(room2, []byte("msg2"))
	hub.BroadcastToRoom(room3, []byte("msg3"))

	time.Sleep(50 * time.Millisecond)

	// Should receive all 3 messages
	receivedCount := 0
	for i := 0; i < 3; i++ {
		select {
		case <-client.Send:
			receivedCount++
		case <-time.After(100 * time.Millisecond):
			break
		}
	}

	if receivedCount != 3 {
		t.Errorf("Client should receive 3 messages, got %d", receivedCount)
	}
}
