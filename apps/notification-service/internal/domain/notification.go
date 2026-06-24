package domain

import (
	"time"
)

type NotificationStatus string

const (
	Pending   NotificationStatus = "PENDING"
	Sent      NotificationStatus = "SENT"
	Failed    NotificationStatus = "FAILED"
	Cancelled NotificationStatus = "CANCELLED"
)

type Notification struct {
	ID        string
	Type      string // ex: "WEBHOOK", "EMAIL"
	Target    string // ex: email@example.com ou webhook url
	Payload   []byte
	Status    NotificationStatus
	Attempts  int
	MaxRetries int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (n *Notification) IncrementAttempt() {
	n.Attempts++
	n.UpdatedAt = time.Now()
}

func (n *Notification) ShouldRetry() bool {
	return n.Attempts < n.MaxRetries
}
