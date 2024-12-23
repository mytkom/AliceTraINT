package repository_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/stretchr/testify/assert"
)

func TestTrainingTaskResultRepository_Create(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingTaskResultRepo := repository.NewTrainingTaskResultRepository(db)

	ttr := &models.TrainingTaskResult{
		Name:           "train log",
		Type:           models.Log,
		Description:    "log of training",
		FileId:         1,
		TrainingTaskId: 1,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "training_task_results" (.+) RETURNING "id"`).
		WithArgs(AnyTime(), AnyTime(), AnyTime(), ttr.Name, ttr.Type, ttr.Description, ttr.FileId, ttr.TrainingTaskId).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := trainingTaskResultRepo.Create(ttr)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingTaskResultRepository_GetAll(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingTaskRepo := repository.NewTrainingTaskResultRepository(db)

	ttrs := []models.TrainingTaskResult{
		{
			Name:           "train log",
			Type:           models.Log,
			Description:    "log of training",
			FileId:         1,
			TrainingTaskId: 1,
		},
		{
			Name:           "benchmark log",
			Type:           models.Log,
			Description:    "log of benchmarking",
			FileId:         2,
			TrainingTaskId: 1,
		},
	}

	taskRows := sqlmock.NewRows([]string{"id", "name", "type", "description", "file_id", "training_task_id"})

	for i, ttr := range ttrs {
		taskRows = taskRows.AddRow(i+1, ttr.Name, ttr.Type, ttr.Description, ttr.FileId, ttr.TrainingTaskId)
	}

	mock.ExpectQuery(`SELECT .* FROM "training_task_results" LEFT JOIN "files" .* WHERE "training_task_id" = (.+) AND .* ORDER BY "training_task_results"."created_at" desc`).
		WillReturnRows(taskRows)

	results, err := trainingTaskRepo.GetAll(1)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, results[0].Name, "train log")
	assert.Equal(t, results[1].Name, "benchmark log")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingTaskResultRepository_GetByType(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingTaskRepo := repository.NewTrainingTaskResultRepository(db)

	ttrs := []models.TrainingTaskResult{
		{
			Name:           "train log",
			Type:           models.Log,
			Description:    "log of training",
			FileId:         1,
			TrainingTaskId: 1,
		},
		{
			Name:           "benchmark log",
			Type:           models.Log,
			Description:    "log of benchmarking",
			FileId:         2,
			TrainingTaskId: 1,
		},
	}

	taskRows := sqlmock.NewRows([]string{"id", "name", "type", "description", "file_id", "training_task_id"})

	for i, ttr := range ttrs {
		taskRows = taskRows.AddRow(i+1, ttr.Name, ttr.Type, ttr.Description, ttr.FileId, ttr.TrainingTaskId)
	}

	mock.ExpectQuery(`SELECT .* FROM "training_task_results" LEFT JOIN "files" .* WHERE "training_task_id" = (.+) AND "type" = (.*) AND .* ORDER BY "training_task_results"."created_at" desc`).
		WillReturnRows(taskRows)

	results, err := trainingTaskRepo.GetByType(1, models.Log)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, results[0].Name, "train log")
	assert.Equal(t, results[1].Name, "benchmark log")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingTaskResultRepository_GetById(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingTaskResultRepo := repository.NewTrainingTaskResultRepository(db)

	ttr := &models.TrainingTaskResult{
		Name:           "train log",
		Type:           models.Log,
		Description:    "log of training",
		FileId:         1,
		TrainingTaskId: 1,
	}

	rows := sqlmock.NewRows([]string{"id", "name", "type", "description", "file_id", "training_task_id"})
	rows = rows.AddRow(1, ttr.Name, ttr.Type, ttr.Description, ttr.FileId, ttr.TrainingTaskId)
	mock.ExpectQuery("SELECT (.*) FROM \"training_task_results\" WHERE \"training_task_results\".\"id\" = (.+) LIMIT (.+)").
		WillReturnRows(rows)

	result, err := trainingTaskResultRepo.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, "train log", result.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingTaskResultRepository_Delete(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingTaskResultRepo := repository.NewTrainingTaskResultRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE \"training_task_results\" SET \"deleted_at\"=(.+) WHERE \"training_task_results\".\"id\" = (.+)").WithArgs(AnyTime(), 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := trainingTaskResultRepo.Delete(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
