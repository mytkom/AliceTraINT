package migrate

import (
	"log"
	"time"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/jalien"
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

func SeedDB(db *gorm.DB) error {
	users := []models.User{
		{FirstName: "Albert", FamilyName: "Einstein", Username: "aeinstein", Email: "aeinstein@cern.ch", CernPersonId: "aeinsteinPersonId"},
		{FirstName: "Niels", FamilyName: "Bohr", Username: "nbohr", Email: "nbohr@cern.ch", CernPersonId: "nbohrPersonId"},
	}

	if err := db.Save(users).Error; err != nil {
		return err
	}

	aods := []jalien.AODFile{{
		Name:      "AO2D.root",
		Path:      "/alice/sim/2024/LHC24b1b/0/567454/AOD/002/AO2D.root",
		Size:      2312421213,
		LHCPeriod: "LHC24b1b",
		RunNumber: 567454,
		AODNumber: 2,
	}}

	trainingDatasets := []models.TrainingDataset{
		{Name: "LHC24b1b", AODFiles: aods, UserId: users[0].ID},
		{Name: "LHC24b1b2", AODFiles: aods, UserId: users[1].ID},
	}

	if err := db.Save(trainingDatasets).Error; err != nil {
		return err
	}

	trainingMachines := []models.TrainingMachine{
		{Name: "tm1", LastActivityAt: time.Now(), SecretKeyHashed: "salt:secret", UserId: users[0].ID},
		{Name: "tm2", LastActivityAt: time.Now(), SecretKeyHashed: "salt:secret2", UserId: users[1].ID},
	}

	if err := db.Save(trainingMachines).Error; err != nil {
		return err
	}

	trainingTasks := []models.TrainingTask{
		{
			Name:              "LHC24b1b Task",
			Status:            models.Queued,
			UserId:            users[0].ID,
			TrainingDatasetId: trainingDatasets[0].ID,
			TrainingMachineId: nil,
			Configuration:     "",
		},
		{
			Name:              "LHC24b1b Task 2",
			Status:            models.Completed,
			UserId:            users[0].ID,
			TrainingDatasetId: trainingDatasets[0].ID,
			TrainingMachineId: &trainingMachines[0].ID,
			Configuration:     "",
		},
		{
			Name:              "LHC24b1b Other Task",
			Status:            models.Queued,
			UserId:            users[1].ID,
			TrainingDatasetId: trainingDatasets[1].ID,
			TrainingMachineId: nil,
			Configuration:     "",
		},
	}

	if err := db.Save(trainingTasks).Error; err != nil {
		return err
	}

	return nil
}
