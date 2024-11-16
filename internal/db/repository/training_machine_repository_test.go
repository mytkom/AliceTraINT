package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/assert"
)

func TestTrainingMachineRepository_Create(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingMachineRepo := NewTrainingMachineRepository(db)

	trainingMachine := &models.TrainingMachine{
		Name:            "m1",
		LastActivityAt:  time.Now(),
		SecretKeyHashed: "salt:secret",
		UserId:          1,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "training_machines" (.+) RETURNING "id"`).
		WithArgs(AnyTime(), AnyTime(), AnyTime(), trainingMachine.Name, trainingMachine.LastActivityAt, trainingMachine.SecretKeyHashed, trainingMachine.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := trainingMachineRepo.Create(trainingMachine)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingMachineRepository_GetAll(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingMachineRepo := NewTrainingMachineRepository(db)

	trainingMachines := []models.TrainingMachine{
		{
			Name:            "m1",
			LastActivityAt:  time.Now(),
			SecretKeyHashed: "salt:secret1",
			UserId:          1,
		},
		{
			Name:            "m2",
			LastActivityAt:  time.Now(),
			SecretKeyHashed: "salt:secret2",
			UserId:          1,
		},
	}

	rows := sqlmock.NewRows([]string{"id", "name", "last_activity_at", "secret_key_hashed", "user_id"})

	for i, machine := range trainingMachines {
		rows = rows.AddRow(i+1, machine.Name, machine.LastActivityAt, machine.SecretKeyHashed, machine.UserId)
	}

	mock.ExpectQuery("SELECT (.*) FROM \"training_machines\" LEFT JOIN \"users\" (.*) ORDER BY \"last_activity_at\" desc").
		WillReturnRows(rows)

	trainingMachines, err := trainingMachineRepo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, trainingMachines, 2)
	assert.Equal(t, "m1", trainingMachines[0].Name)
	assert.Equal(t, "m2", trainingMachines[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingMachineRepository_GetById(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingMachineRepo := NewTrainingMachineRepository(db)

	trainingMachine := &models.TrainingMachine{
		Name:            "m1",
		LastActivityAt:  time.Now(),
		SecretKeyHashed: "salt:secret",
		UserId:          1,
	}

	rows := sqlmock.NewRows([]string{"id", "name", "last_activity_at", "secret_key_hashed", "user_id"})
	rows = rows.AddRow(1, trainingMachine.Name, trainingMachine.LastActivityAt, trainingMachine.SecretKeyHashed, trainingMachine.UserId)

	mock.ExpectQuery("SELECT (.+) FROM \"training_machines\" LEFT JOIN \"users\" (.+) WHERE \"training_machines\".\"id\" = (.+) LIMIT (.+)").
		WillReturnRows(rows)

	trainingDataset, err := trainingMachineRepo.GetByID(3)
	assert.NoError(t, err)
	assert.Equal(t, "m1", trainingDataset.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTrainingMachineRepository_Delete(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	trainingMachineRepo := NewTrainingMachineRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE \"training_machines\" SET \"deleted_at\"=(.+) WHERE \"user_id\" = (.+) AND \"training_machines\".\"id\" = (.+) AND \"training_machines\".\"deleted_at\" IS NULL").WithArgs(AnyTime(), 1, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := trainingMachineRepo.Delete(1, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
