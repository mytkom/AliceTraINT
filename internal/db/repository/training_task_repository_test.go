package repository

import (
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/stretchr/testify/assert"
)

func marshalTrainingTaskConfig(t *testing.T, task *models.TrainingTask) string {
	bytes, err := json.Marshal(task.Configuration)
	assert.NoError(t, err)
	return string(bytes)
}

func TestTrainingTaskRepository_Create(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingTaskRepo := NewTrainingTaskRepository(db)

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

	trainingTask := &models.TrainingTask{
		Name:            "LHC24b1b undersampling",
		Status:          models.Queued,
		TrainingDataset: *trainingDataset,
		UserId:          1,
		Configuration:   struct{ bs uint }{bs: 216},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "training_datasets" (.+) RETURNING "id"`).
		WithArgs(AnyTime(), AnyTime(), AnyTime(), trainingDataset.Name, marshalAODFiles(t, trainingDataset), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`INSERT INTO "training_tasks" (.+) RETURNING "id"`).
		WithArgs(AnyTime(), AnyTime(), AnyTime(), trainingTask.Name, trainingTask.Status, 1, 1, marshalTrainingTaskConfig(t, trainingTask)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := trainingTaskRepo.Create(trainingTask)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingTaskRepository_GetAll(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingTaskRepo := NewTrainingTaskRepository(db)

	trainingTasks := []models.TrainingTask{
		{
			Name:              "LHC24b1b undersampling",
			Status:            models.Queued,
			TrainingDatasetId: 1,
			UserId:            1,
			Configuration:     struct{ bs uint }{bs: 216},
		},
		{
			Name:              "LHC24b1b",
			Status:            models.Completed,
			TrainingDatasetId: 1,
			UserId:            1,
			Configuration:     struct{ bs uint }{bs: 155},
		},
	}

	rows := sqlmock.NewRows([]string{"id", "name", "status", "training_dataset_id", "user_id", "configuration"})

	for i, task := range trainingTasks {
		rows = rows.AddRow(i+1, task.Name, task.Status, 1, 1, marshalTrainingTaskConfig(t, &task))
	}

	mock.ExpectQuery("SELECT (.*) FROM \"training_tasks\" LEFT JOIN \"training_datasets\" (.*) LEFT JOIN \"users\" (.*) ORDER BY \"created_at\" desc").
		WillReturnRows(rows)

	trainingTasks, err := trainingTaskRepo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, trainingTasks, 2)
	assert.Equal(t, trainingTasks[0].Name, "LHC24b1b undersampling")
	assert.Equal(t, trainingTasks[1].Name, "LHC24b1b")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingTaskRepository_GetById(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingTaskRepo := NewTrainingTaskRepository(db)

	trainingTask := &models.TrainingTask{
		Name:              "LHC24b1b undersampling",
		Status:            models.Queued,
		TrainingDatasetId: 1,
		UserId:            1,
		Configuration:     struct{ bs uint }{bs: 155},
	}

	rows := sqlmock.NewRows([]string{"id", "name", "status", "train_dataset_id", "user_id", "configuration"})
	rows = rows.AddRow(1, trainingTask.Name, trainingTask.Status, 1, 1, marshalTrainingTaskConfig(t, trainingTask))
	mock.ExpectQuery("SELECT (.*) FROM \"training_tasks\" LEFT JOIN \"training_datasets\" (.*) LEFT JOIN \"users\" (.*) WHERE \"training_tasks\".\"id\" = (.+) ORDER BY \"training_tasks\".\"id\" LIMIT (.+)").
		WillReturnRows(rows)

	trainingTask, err := trainingTaskRepo.GetByID(3)
	assert.NoError(t, err)
	assert.Equal(t, "LHC24b1b undersampling", trainingTask.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingTaskRepository_Delete(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingTaskRepo := NewTrainingTaskRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE \"training_tasks\" SET \"deleted_at\"=(.+) WHERE \"user_id\" = (.+) AND \"training_tasks\".\"id\" = (.+) AND \"training_tasks\".\"deleted_at\" IS NULL").WithArgs(AnyTime(), 1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := trainingTaskRepo.Delete(1, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingTaskRepository_Update(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingTaskRepo := NewTrainingTaskRepository(db)

	trainingTask := &models.TrainingTask{
		Name:              "LHC24b1b undersampling",
		Status:            models.Queued,
		TrainingDatasetId: 1,
		UserId:            1,
		Configuration:     struct{ bs uint }{bs: 155},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "training_tasks" (.+) RETURNING "id"`).
		WithArgs(AnyTime(), AnyTime(), AnyTime(), trainingTask.Name, trainingTask.Status, 1, 1, marshalTrainingTaskConfig(t, trainingTask)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := trainingTaskRepo.Update(1, trainingTask)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
