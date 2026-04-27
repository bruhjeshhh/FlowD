package worker

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bruhjeshhh/flowD/internal/database"
)

type WebhookPayload struct {
	Event    string          `json:"event"`
	JobID    uuid.UUID       `json:"job_id"`
	JobType  string         `json:"job_type"`
	Status  string          `json:"status"`
	Payload json.RawMessage `json:"payload"`
	Time    time.Time      `json:"time"`
}

type WebhookClient struct {
	httpClient *http.Client
	db        *database.Queries
	log       *slog.Logger
}

func NewWebhookClient(db *database.Queries, log *slog.Logger) *WebhookClient {
	return &WebhookClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		db:  db,
		log: log,
	}
}

func (c *WebhookClient) DeliverJobEvent(ctx context.Context, jobType, event string, job *database.Job) error {
	webhooks, err := c.db.GetWebhooksByJobTypeAndEvent(ctx, database.GetWebhooksByJobTypeAndEventParams{
		JobType: jobType,
		Event:  event,
	})
	if err != nil {
		return fmt.Errorf("failed to get webhooks: %w", err)
	}

	if len(webhooks) == 0 {
		return nil
	}

	payload := WebhookPayload{
		Event:    event,
		JobID:    job.ID,
		JobType:  job.Type,
		Status:  job.Status.String,
		Payload: job.Payload,
		Time:    time.Now().UTC(),
	}

	for _, webhook := range webhooks {
		if err := c.deliver(ctx, webhook, payload); err != nil {
			c.log.Error("webhook delivery failed",
				"webhook_id", webhook.ID,
				"url", webhook.URL,
				"error", err)
		}
	}

	return nil
}

func (c *WebhookClient) deliver(ctx context.Context, webhook database.Webhook, payload WebhookPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", payload.Event)

	if webhook.Secret.Valid {
		signature := c.computeSignature(body, webhook.Secret.String)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(bodyBytes))
}

func (c *WebhookClient) computeSignature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func StartWebhookDelivery(db *pgxpool.Pool, log *slog.Logger) {
	q := database.New(db)
	client := NewWebhookClient(q, log)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if ctx.Err() != nil {
			break
		}
	}
}