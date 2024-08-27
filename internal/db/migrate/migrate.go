package migrate

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"gorm.io/gorm"
	"log"
)

func MigrateDB(db *gorm.DB) {
	err := db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
}
