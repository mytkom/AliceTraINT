package migrate

import (
	"log"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"gorm.io/gorm"
)

func MigrateDB(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.User{},
		&models.TrainingDataset{},
		&models.TrainingTask{},
		&models.TrainingMachine{},
		&models.TrainingTaskResult{},
		&models.File{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
}
