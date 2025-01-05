package migrate

import (
	"log"
	"time"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/hash"
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

	c1Aods := []jalien.AODFile{
		{
			Name:      "AO2D.root",
			Path:      "/alice/sim/2023/LHC23c1/302004/AOD/010/AO2D.root",
			Size:      3266476446,
			LHCPeriod: "LHC23c1",
			RunNumber: 302004,
			AODNumber: 10,
		},
		{
			Name:      "AO2D.root",
			Path:      "/alice/sim/2023/LHC23c1/302004/AOD/011/AO2D.root",
			Size:      3239114872,
			LHCPeriod: "LHC23c1",
			RunNumber: 302004,
			AODNumber: 11,
		},
		{
			Name:      "AO2D.root",
			Path:      "/alice/sim/2023/LHC23c1/302004/AOD/013/AO2D.root",
			Size:      3260265579,
			LHCPeriod: "LHC23c1",
			RunNumber: 302004,
			AODNumber: 13,
		},
	}

	mixedPeriodsAods := []jalien.AODFile{
		{
			Name:      "AO2D.root",
			Path:      "/alice/sim/2023/LHC23c1/302004/AOD/010/AO2D.root",
			Size:      3266476446,
			LHCPeriod: "LHC23c1",
			RunNumber: 302004,
			AODNumber: 10,
		},
		{
			Name:      "AO2D.root",
			Path:      "/alice/sim/2023/LHC23c1/302004/AOD/011/AO2D.root",
			Size:      3239114872,
			LHCPeriod: "LHC23c1",
			RunNumber: 302004,
			AODNumber: 11,
		},
		{
			Name:      "AO2D.root",
			Path:      "/alice/sim/2023/LHC23c1/302004/AOD/013/AO2D.root",
			Size:      3260265579,
			LHCPeriod: "LHC23c1",
			RunNumber: 302004,
			AODNumber: 13,
		},
		{
			Name:      "AO2D.root",
			Path:      "/alice/sim/2023/LHC23d4/302005/AOD/013/AO2D.root",
			Size:      35403114,
			LHCPeriod: "LHC23d4",
			RunNumber: 302002,
			AODNumber: 13,
		},
		{
			Name:      "AO2D.root",
			Path:      "/alice/sim/2023/LHC23d4/302005/AOD/024/AO2D.root",
			Size:      97906832,
			LHCPeriod: "LHC23d4",
			RunNumber: 302002,
			AODNumber: 24,
		},
		{
			Name:      "AO2D.root",
			Path:      "/alice/sim/2023/LHC23d4/302005/AOD/030/AO2D.root",
			Size:      175726295,
			LHCPeriod: "LHC23d4",
			RunNumber: 302002,
			AODNumber: 30,
		},
	}

	trainingDatasets := []models.TrainingDataset{
		{Name: "Mixed periods 2023", AODFiles: mixedPeriodsAods, UserId: users[0].ID},
		{Name: "LHC23c1", AODFiles: c1Aods, UserId: users[1].ID},
	}

	if err := db.Save(trainingDatasets).Error; err != nil {
		return err
	}

	secretHashed, err := hash.HashKey("secretkey_secretkey_secretkey_sk")
	if err != nil {
		return err
	}
	trainingMachines := []models.TrainingMachine{
		{Name: "tm1", LastActivityAt: time.Now(), SecretKeyHashed: secretHashed, UserId: users[0].ID},
		{Name: "tm2", LastActivityAt: time.Now(), SecretKeyHashed: secretHashed, UserId: users[1].ID},
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
