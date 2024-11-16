package repository

import (
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/stretchr/testify/assert"
)

func marshalAODFiles(t *testing.T, dataset *models.TrainingDataset) string {
	bytes, err := json.Marshal(dataset.AODFiles)
	assert.NoError(t, err)
	return string(bytes)
}

func TestTrainingDatasetRepository_Create(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingDatasetRepo := NewTrainingDatasetRepository(db)

	trainingDataset := &models.TrainingDataset{
		Name: "fbw",
		AODFiles: []jalien.AODFile{
			{
				Name:      "AO2D.root",
				Path:      "/alice/sim/2024/LHC24f3/0/654324/AOD/013",
				Size:      3000000000,
				LHCPeriod: "LHC24f3",
				RunNumber: 654324,
				AODNumber: 13,
			},
		},
		UserId: 1,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "training_datasets" (.+) RETURNING "id"`).
		WithArgs(AnyTime(), AnyTime(), AnyTime(), trainingDataset.Name, marshalAODFiles(t, trainingDataset), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := trainingDatasetRepo.Create(trainingDataset)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingDatasetRepository_GetAll(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingDatasetRepo := NewTrainingDatasetRepository(db)

	trainingDatasets := []models.TrainingDataset{
		{
			Name: "fbw2",
			AODFiles: []jalien.AODFile{
				{
					Name:      "AO2D.root",
					Path:      "/alice/sim/2024/LHC24f3/0/654324/AOD/013",
					Size:      3000000000,
					LHCPeriod: "LHC24f3",
					RunNumber: 654324,
					AODNumber: 12,
				},
			},
		},
		{
			Name: "fbw",
			AODFiles: []jalien.AODFile{
				{
					Name:      "AO2D.root",
					Path:      "/alice/sim/2024/LHC24f3/0/654324/AOD/013",
					Size:      3000000000,
					LHCPeriod: "LHC24f3",
					RunNumber: 654324,
					AODNumber: 13,
				},
			},
		},
	}

	rows := sqlmock.NewRows([]string{"id", "name", "aod_files"})

	for i, dataset := range trainingDatasets {
		rows = rows.AddRow(i+1, dataset.Name, marshalAODFiles(t, &dataset))
	}

	mock.ExpectQuery("SELECT (.*) FROM \"training_datasets\" LEFT JOIN \"users\" (.*) ORDER BY \"created_at\" desc").
		WillReturnRows(rows)

	trainingDatasets, err := trainingDatasetRepo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, trainingDatasets, 2)
	assert.Equal(t, "fbw2", trainingDatasets[0].Name)
	assert.Equal(t, uint64(12), trainingDatasets[0].AODFiles[0].AODNumber)
	assert.Equal(t, "fbw", trainingDatasets[1].Name)
	assert.Equal(t, uint64(13), trainingDatasets[1].AODFiles[0].AODNumber)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingDatasetRepository_GetById(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingDatasetRepo := NewTrainingDatasetRepository(db)

	mockTrainingDataset := &models.TrainingDataset{
		Name: "fbw2",
		AODFiles: []jalien.AODFile{
			{
				Name:      "AO2D.root",
				Path:      "/alice/sim/2024/LHC24f3/0/654324/AOD/013",
				Size:      3000000000,
				LHCPeriod: "LHC24f3",
				RunNumber: 654324,
				AODNumber: 12,
			},
		},
	}

	rows := sqlmock.NewRows([]string{"id", "name", "aod_files"})
	rows = rows.AddRow(1, mockTrainingDataset.Name, marshalAODFiles(t, mockTrainingDataset))

	mock.ExpectQuery("SELECT (.+) FROM \"training_datasets\" LEFT JOIN \"users\" (.+) WHERE \"training_datasets\".\"id\" = (.+) ORDER BY \"created_at\" desc(.+)LIMIT (.+)").
		WillReturnRows(rows)

	trainingDataset, err := trainingDatasetRepo.GetByID(3)
	assert.NoError(t, err)
	assert.Equal(t, "fbw2", trainingDataset.Name)
	assert.Equal(t, uint64(12), trainingDataset.AODFiles[0].AODNumber)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingDatasetRepository_Delete(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingDatasetRepo := NewTrainingDatasetRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE \"training_datasets\" SET \"deleted_at\"=(.+) WHERE \"user_id\" = (.+) AND \"training_datasets\".\"id\" = (.+) AND \"training_datasets\".\"deleted_at\" IS NULL").WithArgs(AnyTime(), 1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := trainingDatasetRepo.Delete(1, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
