package migrate

import (
	"log"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"gorm.io/gorm"
)

func MigrateDB(db *gorm.DB) {
	err := db.AutoMigrate(&models.User{}, &models.TrainingDataset{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
}
