package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"codebasego/internal/common"
	"codebasego/internal/platform/database"
)

type OutboxStatus string

const (
	OutboxStatusPending    OutboxStatus = "PENDING"
	OutboxStatusProcessing OutboxStatus = "PROCESSING"
	OutboxStatusPublished  OutboxStatus = "PUBLISHED"
	OutboxStatusFailed     OutboxStatus = "FAILED"
)

// OutboxEvent represents an event persisted in the DB as part of a local transaction (Transactional Outbox Pattern).
type OutboxEvent struct {
	ID          string       `gorm:"primaryKey;size:36" json:"id"`
	EventName   string       `gorm:"size:255;not null;index" json:"event_name"`
	Payload     string       `gorm:"type:text;not null" json:"payload"`
	TraceID     string       `gorm:"size:64;index" json:"trace_id,omitempty"`
	Status      OutboxStatus `gorm:"size:32;not null;default:'PENDING';index" json:"status"`
	RetryCount  int          `gorm:"not null;default:0" json:"retry_count"`
	MaxRetries  int          `gorm:"not null;default:3" json:"max_retries"`
	LastError   string       `gorm:"type:text" json:"last_error,omitempty"`
	CreatedAt   time.Time    `gorm:"not null;index" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"not null;index" json:"updated_at"`
	PublishedAt *time.Time   `json:"published_at,omitempty"`
}

func (OutboxEvent) TableName() string {
	return "outbox_events"
}

// SaveOutboxEvent persists an event into the outbox table within the current database transaction in ctx.
func SaveOutboxEvent(ctx context.Context, defaultDB *gorm.DB, event common.Event) error {
	db := database.GetDB(ctx, defaultDB)
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	var payloadStr string
	switch p := event.Payload.(type) {
	case string:
		payloadStr = p
	case []byte:
		payloadStr = string(p)
	default:
		data, err := json.Marshal(p)
		if err != nil {
			return fmt.Errorf("failed to marshal outbox event payload: %w", err)
		}
		payloadStr = string(data)
	}

	traceID := event.TraceID
	if traceID == "" {
		traceID = common.GetTraceID(ctx)
	}

	now := time.Now()
	outbox := &OutboxEvent{
		ID:         uuid.NewString(),
		EventName:  event.Name,
		Payload:    payloadStr,
		TraceID:    traceID,
		Status:     OutboxStatusPending,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := db.Create(outbox).Error; err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}
	return nil
}

// OutboxProcessor polls pending outbox events and publishes them to the EventBus reliably.
type OutboxProcessor struct {
	db        *gorm.DB
	bus       common.EventBus
	logger    zerolog.Logger
	interval  time.Duration
	batchSize int
}

func NewOutboxProcessor(db *gorm.DB, bus common.EventBus, log zerolog.Logger) *OutboxProcessor {
	return &OutboxProcessor{
		db:        db,
		bus:       bus,
		logger:    log.With().Str("component", "outbox_processor").Logger(),
		interval:  1 * time.Second,
		batchSize: 50,
	}
}

// AutoMigrate creates the outbox_events table if it does not exist.
func (p *OutboxProcessor) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&OutboxEvent{})
}

// StartBackground starts the periodic outbox polling worker with graceful shutdown support.
func (p *OutboxProcessor) StartBackground(ctx context.Context, wg *sync.WaitGroup) {
	if p.db == nil || p.bus == nil {
		p.logger.Warn().Msg("database or eventbus is nil; outbox processor skipped")
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		p.logger.Info().Msg("outbox processor background worker started")
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				p.logger.Info().Msg("outbox processor worker stopping gracefully")
				return
			case <-ticker.C:
				if err := p.ProcessPending(ctx); err != nil {
					p.logger.Error().Err(err).Msg("error processing pending outbox events")
				}
			}
		}
	}()
}

// ProcessPending reads and dispatches pending events from the outbox table.
// It uses atomic batch claiming with timeout recovery to prevent race conditions across multiple instances.
func (p *OutboxProcessor) ProcessPending(ctx context.Context) error {
	now := time.Now()
	lockTimeout := now.Add(-2 * time.Minute)

	// Step 1: Find candidate event IDs (PENDING or stalled in PROCESSING for > 2m)
	var candidateIDs []string
	err := p.db.WithContext(ctx).Model(&OutboxEvent{}).
		Where("retry_count < max_retries AND (status = ? OR (status = ? AND updated_at < ?))",
			OutboxStatusPending, OutboxStatusProcessing, lockTimeout).
		Order("created_at ASC").
		Limit(p.batchSize).
		Pluck("id", &candidateIDs).Error
	if err != nil {
		return err
	}

	if len(candidateIDs) == 0 {
		return nil
	}

	// Step 2: Atomically claim the candidate batch to avoid duplicate dispatching across instances
	claimRes := p.db.WithContext(ctx).Model(&OutboxEvent{}).
		Where("id IN (?) AND (status = ? OR (status = ? AND updated_at < ?))",
			candidateIDs, OutboxStatusPending, OutboxStatusProcessing, lockTimeout).
		Updates(map[string]any{
			"status":     OutboxStatusProcessing,
			"updated_at": now,
		})
	if claimRes.Error != nil {
		return claimRes.Error
	}
	if claimRes.RowsAffected == 0 {
		return nil
	}

	// Step 3: Fetch the claimed events
	var events []OutboxEvent
	if err := p.db.WithContext(ctx).
		Where("id IN (?) AND status = ?", candidateIDs, OutboxStatusProcessing).
		Find(&events).Error; err != nil {
		return err
	}

	// Step 4: Dispatch events safely
	for _, evt := range events {
		p.dispatchEvent(ctx, evt)
	}
	return nil
}

func (p *OutboxProcessor) dispatchEvent(ctx context.Context, evt OutboxEvent) {
	var payload any
	if err := json.Unmarshal([]byte(evt.Payload), &payload); err != nil {
		payload = evt.Payload
	}

	eventCtx := ctx
	if evt.TraceID != "" {
		eventCtx = common.WithTraceID(ctx, evt.TraceID)
	}

	var dispatchErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				dispatchErr = fmt.Errorf("panic during publish: %v", r)
			}
		}()
		p.bus.Publish(eventCtx, common.Event{
			Name:    evt.EventName,
			Payload: payload,
			TraceID: evt.TraceID,
		})
	}()

	now := time.Now()
	if dispatchErr != nil {
		p.logger.Error().Err(dispatchErr).Str("outbox_id", evt.ID).Msg("failed to dispatch outbox event")
		newRetry := evt.RetryCount + 1
		newStatus := OutboxStatusPending
		if newRetry >= evt.MaxRetries {
			newStatus = OutboxStatusFailed
		}
		_ = p.db.WithContext(ctx).Model(&OutboxEvent{}).
			Where("id = ?", evt.ID).
			Updates(map[string]any{
				"status":      newStatus,
				"retry_count": newRetry,
				"last_error":  dispatchErr.Error(),
				"updated_at":  now,
			}).Error
		return
	}

	updateErr := p.db.WithContext(ctx).Model(&OutboxEvent{}).
		Where("id = ?", evt.ID).
		Updates(map[string]any{
			"status":       OutboxStatusPublished,
			"published_at": &now,
			"updated_at":   now,
		}).Error
	if updateErr != nil {
		p.logger.Error().Err(updateErr).Str("outbox_id", evt.ID).Msg("failed to update outbox event to published")
	}
}
