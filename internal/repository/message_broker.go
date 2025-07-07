package repository

import (
	"context"
	"encoding/json"
)


type MessageBroker interface {

	Publish(ctx context.Context, stream string, message interface{}) error
}


type Message struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

		
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
