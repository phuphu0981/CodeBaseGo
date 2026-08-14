package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"codebasego/internal/platform/database"
)

type DummyItem struct {
	ID   string `gorm:"primaryKey"`
	Name string
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite in memory: %v", err)
	}
	if err := db.AutoMigrate(&DummyItem{}); err != nil {
		t.Fatalf("failed to auto migrate: %v", err)
	}
	return db
}

func TestWithTransaction_Commit(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	err := database.WithTransaction(ctx, db, func(txCtx context.Context) error {
		txDB := database.GetDB(txCtx, db)
		return txDB.Create(&DummyItem{ID: "1", Name: "Item 1"}).Error
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	var count int64
	db.Model(&DummyItem{}).Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 item, got %d", count)
	}
}

func TestWithTransaction_Rollback(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	expectedErr := errors.New("simulated failure")
	err := database.WithTransaction(ctx, db, func(txCtx context.Context) error {
		txDB := database.GetDB(txCtx, db)
		if err := txDB.Create(&DummyItem{ID: "2", Name: "Item 2"}).Error; err != nil {
			return err
		}
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	var count int64
	db.Model(&DummyItem{}).Count(&count)
	if count != 0 {
		t.Fatalf("expected 0 items due to rollback, got %d", count)
	}
}

func TestGetDB_Fallback(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	retrieved := database.GetDB(ctx, db)
	if retrieved == nil {
		t.Fatalf("expected non-nil db")
	}
}
