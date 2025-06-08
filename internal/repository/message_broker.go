package repository

import (
	"context"
	"encoding/json"
)

// MessageBroker defines the interface for message publishing operations
type MessageBroker interface {
	// Publish sends a message to a stream
	Publish(ctx context.Context, stream string, message interface{}) error
}

// Message represents a message to be published
type Message struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// NewMessage creates a new message
func NewMessage(id, msgType string, payload interface{}) (*Message, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:      id,
		Type:    msgType,
		Payload: payloadBytes,
	}, nil
}
