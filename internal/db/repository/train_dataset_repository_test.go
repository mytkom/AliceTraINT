package repository

import (
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/stretchr/testify/assert"
)

func marshalAODFiles(t *testing.T, dataset *models.TrainDataset) string {
	bytes, err := json.Marshal(dataset.AODFiles)
	assert.NoError(t, err)
	return string(bytes)
}

func TestTrainDatasetRepository_Create(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainDatasetRepo := NewTrainDatasetRepository(db)

	trainDataset := &models.TrainDataset{
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
	mock.ExpectQuery(`INSERT INTO "train_datasets" (.+) RETURNING "id"`).
		WithArgs(AnyTime(), AnyTime(), AnyTime(), trainDataset.Name, marshalAODFiles(t, trainDataset), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := trainDatasetRepo.Create(trainDataset)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainDatasetRepository_GetAll(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainDatasetRepo := NewTrainDatasetRepository(db)

	trainDatasets := []models.TrainDataset{
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

	for i, dataset := range trainDatasets {
		rows = rows.AddRow(i+1, dataset.Name, marshalAODFiles(t, &dataset))
	}

	mock.ExpectQuery("SELECT \\* FROM \"train_datasets\"").
		WillReturnRows(rows)

	trainDatasets, err := trainDatasetRepo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, trainDatasets, 2)
	assert.Equal(t, "fbw2", trainDatasets[0].Name)
	assert.Equal(t, uint64(12), trainDatasets[0].AODFiles[0].AODNumber)
	assert.Equal(t, "fbw", trainDatasets[1].Name)
	assert.Equal(t, uint64(13), trainDatasets[1].AODFiles[0].AODNumber)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainDatasetRepository_GetById(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainDatasetRepo := NewTrainDatasetRepository(db)

	mockTrainDataset := &models.TrainDataset{
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
	rows = rows.AddRow(1, mockTrainDataset.Name, marshalAODFiles(t, mockTrainDataset))

	mock.ExpectQuery("SELECT \\* FROM \"train_datasets\" WHERE \"train_datasets\".\"id\" = (.+) ORDER BY \"train_datasets\".\"id\" LIMIT (.+)").
		WillReturnRows(rows)

	trainDataset, err := trainDatasetRepo.GetByID(3)
	assert.NoError(t, err)
	assert.Equal(t, "fbw2", trainDataset.Name)
	assert.Equal(t, uint64(12), trainDataset.AODFiles[0].AODNumber)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainDatasetRepository_Delete(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainDatasetRepo := NewTrainDatasetRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE \"train_datasets\" SET \"deleted_at\"=(.+) WHERE \"user_id\" = (.+) AND \"train_datasets\".\"id\" = (.+) AND \"train_datasets\".\"deleted_at\" IS NULL").WithArgs(AnyTime(), 1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := trainDatasetRepo.Delete(1, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainDatasetRepository_Update(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainDatasetRepo := NewTrainDatasetRepository(db)
	mockTrainDataset := &models.TrainDataset{
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
		UserId: 1,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "train_datasets" (.+) RETURNING "id"`).
		WithArgs(AnyTime(), AnyTime(), AnyTime(), mockTrainDataset.Name, marshalAODFiles(t, mockTrainDataset), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := trainDatasetRepo.Update(1, mockTrainDataset)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
