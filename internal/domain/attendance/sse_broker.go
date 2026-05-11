package attendance

import (
	"fmt"
	"sync"
)

// SSEBroker manages Server-Sent Events connections for NFC registration.
// Admin clients subscribe via GET /iot/listen, and ESP32 devices publish
// NFC UIDs via POST /iot/assign. The broker fans out events to all listeners.
type SSEBroker struct {
	mu          sync.RWMutex
	subscribers map[string]chan SSEEvent // key: subscriber ID
}

// SSEEvent represents an event pushed to SSE clients.
type SSEEvent struct {
	Event string `json:"event"` // "nfc_detected"
	Data  string `json:"data"`  // JSON payload
}

// NewSSEBroker creates a new SSE broker.
func NewSSEBroker() *SSEBroker {
	return &SSEBroker{
		subscribers: make(map[string]chan SSEEvent),
	}
}

// Subscribe registers a new SSE client and returns its event channel.
func (b *SSEBroker) Subscribe(id string) chan SSEEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan SSEEvent, 10) // buffered to prevent blocking
	b.subscribers[id] = ch
	return ch
}

// Unsubscribe removes a client from the broker.
func (b *SSEBroker) Unsubscribe(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ch, ok := b.subscribers[id]; ok {
		close(ch)
		delete(b.subscribers, id)
	}
}

// Publish sends an event to all connected SSE clients.
func (b *SSEBroker) Publish(event SSEEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			// Skip slow consumers
		}
	}
}

// PublishNFCDetected broadcasts a detected NFC UID to all admin listeners.
func (b *SSEBroker) PublishNFCDetected(nfcUID string, deviceID uint, deviceName string) {
	data := fmt.Sprintf(`{"nfc_uid":"%s","device_id":%d,"device_name":"%s"}`, nfcUID, deviceID, deviceName)
	b.Publish(SSEEvent{
		Event: "nfc_detected",
		Data:  data,
	})
}

// ActiveCount returns the number of active SSE subscribers.
func (b *SSEBroker) ActiveCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers)
}
