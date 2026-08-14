package eventbus

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"codebasego/internal/common"
	"codebasego/internal/platform/database"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(1)
	}
	if err := db.AutoMigrate(&OutboxEvent{}); err != nil {
		t.Fatalf("failed to auto migrate outbox: %v", err)
	}
	return db
}

func TestTransactionalOutbox(t *testing.T) {
	db := setupTestDB(t)
	bus := NewMemoryEventBus(zerolog.Nop())
	processor := NewOutboxProcessor(db, bus, zerolog.Nop())

	t.Run("SaveOutboxEvent within transaction and process", func(t *testing.T) {
		ctx := common.WithTraceID(context.Background(), "trace-outbox-123")

		// 1. Save inside a transaction
		err := database.WithTransaction(ctx, db, func(txCtx context.Context) error {
			return SaveOutboxEvent(txCtx, db, common.Event{
				Name:    "order.created",
				Payload: map[string]string{"order_id": "ord-999"},
			})
		})
		if err != nil {
			t.Fatalf("failed to save outbox event: %v", err)
		}

		// Verify event was saved with status PENDING and trace ID
		var saved OutboxEvent
		if err := db.First(&saved, "event_name = ?", "order.created").Error; err != nil {
			t.Fatalf("failed to find outbox event: %v", err)
		}
		if saved.Status != OutboxStatusPending {
			t.Fatalf("expected status PENDING, got %s", saved.Status)
		}
		if saved.TraceID != "trace-outbox-123" {
			t.Fatalf("expected trace_id trace-outbox-123, got %s", saved.TraceID)
		}

		// 2. Subscribe to eventbus
		var receivedEvent common.Event
		var wg sync.WaitGroup
		wg.Add(1)
		bus.Subscribe("order.created", func(ctx context.Context, event common.Event) error {
			receivedEvent = event
			wg.Done()
			return nil
		})

		// 3. Process pending outbox events
		if err := processor.ProcessPending(ctx); err != nil {
			t.Fatalf("ProcessPending error: %v", err)
		}

		// Wait for event to arrive
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			if receivedEvent.TraceID != "trace-outbox-123" {
				t.Fatalf("expected received trace_id trace-outbox-123, got %s", receivedEvent.TraceID)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for outbox event delivery")
		}

		// 4. Verify status updated to PUBLISHED
		var updated OutboxEvent
		if err := db.First(&updated, "id = ?", saved.ID).Error; err != nil {
			t.Fatalf("failed to find updated outbox event: %v", err)
		}
		if updated.Status != OutboxStatusPublished {
			t.Fatalf("expected status PUBLISHED, got %s", updated.Status)
		}
		if updated.PublishedAt == nil {
			t.Fatal("expected published_at timestamp to be set")
		}
	})

	t.Run("Transaction Rollback discards Outbox Event", func(t *testing.T) {
		ctx := context.Background()

		// Attempt transaction that fails
		_ = database.WithTransaction(ctx, db, func(txCtx context.Context) error {
			_ = SaveOutboxEvent(txCtx, db, common.Event{
				Name:    "payment.failed",
				Payload: "rollback-me",
			})
			return gorm.ErrInvalidTransaction
		})

		var count int64
		db.Model(&OutboxEvent{}).Where("event_name = ?", "payment.failed").Count(&count)
		if count != 0 {
			t.Fatalf("expected 0 events after rollback, got %d", count)
		}
	})

	t.Run("Concurrent processors do not duplicate dispatch", func(t *testing.T) {
		ctx := context.Background()

		// Save 5 events
		for i := 0; i < 5; i++ {
			_ = SaveOutboxEvent(ctx, db, common.Event{
				Name:    "bulk.event",
				Payload: map[string]int{"index": i},
			})
		}

		var countMu sync.Mutex
		receivedCount := 0
		bus.Subscribe("bulk.event", func(ctx context.Context, event common.Event) error {
			countMu.Lock()
			receivedCount++
			countMu.Unlock()
			return nil
		})

		// Run 3 concurrent processors
		var procWg sync.WaitGroup
		proc1 := NewOutboxProcessor(db, bus, zerolog.Nop())
		proc2 := NewOutboxProcessor(db, bus, zerolog.Nop())
		proc3 := NewOutboxProcessor(db, bus, zerolog.Nop())

		procWg.Add(3)
		go func() { defer procWg.Done(); _ = proc1.ProcessPending(ctx) }()
		go func() { defer procWg.Done(); _ = proc2.ProcessPending(ctx) }()
		go func() { defer procWg.Done(); _ = proc3.ProcessPending(ctx) }()
		procWg.Wait()

		// Give async handlers a moment to complete
		time.Sleep(100 * time.Millisecond)

		countMu.Lock()
		finalCount := receivedCount
		countMu.Unlock()

		if finalCount != 5 {
			t.Fatalf("expected exactly 5 events received without duplicates, got %d", finalCount)
		}
	})
}
