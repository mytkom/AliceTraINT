package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/assert"
)

func TestTrainingTaskResultRepository_Create(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingTaskResultRepo := NewTrainingTaskResultRepository(db)

	ttr := &models.TrainingTaskResult{
		Name:           "train log",
		Type:           models.Log,
		Description:    "log of training",
		File:           []byte("file"),
		TrainingTaskId: 1,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "training_task_results" (.+) RETURNING "id"`).
		WithArgs(AnyTime(), AnyTime(), AnyTime(), ttr.Name, ttr.Type, ttr.Description, ttr.File, ttr.TrainingTaskId).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := trainingTaskResultRepo.Create(ttr)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingTaskResultRepository_GetAll(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingTaskRepo := NewTrainingTaskResultRepository(db)

	ttrs := []models.TrainingTaskResult{
		{
			Name:           "train log",
			Type:           models.Log,
			Description:    "log of training",
			File:           []byte("file"),
			TrainingTaskId: 1,
		},
		{
			Name:           "benchmark log",
			Type:           models.Log,
			Description:    "log of benchmarking",
			File:           []byte("file"),
			TrainingTaskId: 1,
		},
	}

	taskRows := sqlmock.NewRows([]string{"id", "name", "type", "description", "file", "training_task_id"})

	for i, ttr := range ttrs {
		taskRows = taskRows.AddRow(i+1, ttr.Name, ttr.Type, ttr.Description, ttr.File, ttr.TrainingTaskId)
	}

	mock.ExpectQuery("SELECT (.*) FROM \"training_task_results\" WHERE \"training_task_id\" = (.*) ORDER BY \"created_at\" desc").
		WillReturnRows(taskRows)

	ttrs, err := trainingTaskRepo.GetAll(1)
	assert.NoError(t, err)
	assert.Len(t, ttrs, 2)
	assert.Equal(t, ttrs[0].Name, "train log")
	assert.Equal(t, ttrs[1].Name, "benchmark log")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingTaskResultRepository_GetById(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingTaskResultRepo := NewTrainingTaskResultRepository(db)

	ttr := &models.TrainingTaskResult{
		Name:           "train log",
		Type:           models.Log,
		Description:    "log of training",
		File:           []byte("file"),
		TrainingTaskId: 1,
	}

	rows := sqlmock.NewRows([]string{"id", "name", "type", "description", "file", "training_task_id"})
	rows = rows.AddRow(1, ttr.Name, ttr.Type, ttr.Description, ttr.File, ttr.TrainingTaskId)
	mock.ExpectQuery("SELECT (.*) FROM \"training_task_results\" WHERE \"training_task_results\".\"id\" = (.+) LIMIT (.+)").
		WillReturnRows(rows)

	ttr, err := trainingTaskResultRepo.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, "train log", ttr.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingTaskResultRepository_Delete(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingTaskResultRepo := NewTrainingTaskResultRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE \"training_task_results\" SET \"deleted_at\"=(.+) WHERE \"training_task_results\".\"id\" = (.+)").WithArgs(AnyTime(), 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := trainingTaskResultRepo.Delete(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingTaskResultRepository_Update(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()
	trainingTaskResultRepo := NewTrainingTaskResultRepository(db)
	ttr := &models.TrainingTaskResult{
		Name:           "train log",
		Type:           models.Log,
		Description:    "log of training",
		File:           []byte("file"),
		TrainingTaskId: 1,
	}
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "training_task_results" (.+) RETURNING "id"`).
		WithArgs(AnyTime(), AnyTime(), AnyTime(), ttr.Name, ttr.Type, ttr.Description, ttr.File, ttr.TrainingTaskId).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	err := trainingTaskResultRepo.Update(ttr)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
