package domain

import (
	"time"

	"github.com/google/uuid"
)

type OutboxStatus string

const (
	OutboxPending   OutboxStatus = "PENDING"
	OutboxPublished OutboxStatus = "PUBLISHED"
	OutboxFailed    OutboxStatus = "FAILED"
)

type OutboxItem struct {
	ID            uuid.UUID
	AggregateType string
	AggregateID   uuid.UUID
	EventType     string
	Payload       []byte // JSON payload
	Status        OutboxStatus
	Attempts      int
	LastError     string
	CreatedAt     time.Time
	PublishedAt   *time.Time
}

func NewOutboxItem(aggregateType string, aggregateID uuid.UUID, eventType string, payload []byte) *OutboxItem {
	return &OutboxItem{
		ID:            uuid.New(),
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Payload:       payload,
		Status:        OutboxPending,
		Attempts:      0,
		CreatedAt:     time.Now(),
	}
}
