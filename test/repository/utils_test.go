package repository_test

import (
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	dialector := postgres.New(postgres.Config{
		DSN:                  "sqlmock_db_0",
		Conn:                 db,
		PreferSimpleProtocol: true,
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open gorm DB: %v", err)
	}

	cleanup := func() {
		_ = db.Close()
	}

	return gormDB, mock, cleanup
}

type anyTime struct{}

func (a anyTime) Match(v driver.Value) bool {
	if v == nil {
		return true
	}
	_, ok := v.(time.Time)
	return ok
}

func AnyTime() anyTime {
	return anyTime{}
}
