package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"codebasego/internal/platform/config"
)

// NewDB initializes GORM database connection based on config.
// Supports "mysql", "postgres", and "sqlite".
func NewDB(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.DB.Driver {
	case "mysql":
		if cfg.DB.DSN == "" {
			return nil, fmt.Errorf("mysql DSN cannot be empty")
		}
		dialector = mysql.Open(cfg.DB.DSN)

	case "postgres":
		if cfg.DB.DSN == "" {
			return nil, fmt.Errorf("postgres DSN cannot be empty")
		}
		dialector = postgres.Open(cfg.DB.DSN)

	case "sqlite", "":
		dsn := cfg.DB.DSN
		if dsn == "" {
			dsn = "app.db"
		}
		if !strings.Contains(dsn, "_pragma") && !strings.Contains(dsn, "?") {
			dsn += "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"
		}
		dialector = sqlite.Open(dsn)

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.DB.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database (%s): %w", cfg.DB.Driver, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	maxOpen := cfg.DB.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 25
	}
	maxIdle := cfg.DB.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 10
	}
	lifetime := time.Duration(cfg.DB.ConnMaxLifetimeMinute) * time.Minute
	if lifetime <= 0 {
		lifetime = 15 * time.Minute
	}

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(lifetime)

	return db, nil
}

