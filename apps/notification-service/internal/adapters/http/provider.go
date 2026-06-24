package http

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

type NotificationProvider struct {
	client *http.Client
}

func NewNotificationProvider() *NotificationProvider {
	return &NotificationProvider{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (p *NotificationProvider) Send(ctx context.Context, target string, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, "POST", target, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao enviar requisição HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("provedor externo retornou erro: status %d", resp.StatusCode)
	}

	return nil
}
