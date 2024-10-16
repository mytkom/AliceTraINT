package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_Create(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)

	user := &models.User{
		CernPersonId: "1",
		Username:     "aeinstein",
		FirstName:    "Albert",
		FamilyName:   "Einstein",
		Email:        "aeinstein@example.com",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users" (.+) RETURNING "id"`).
		WithArgs(AnyTime(), AnyTime(), AnyTime(), user.CernPersonId, user.Username, user.FirstName, user.FamilyName, user.Email).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := userRepo.Create(user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetAll(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "cern_person_id", "username", "first_name", "family_name", "email"}).
		AddRow(1, "1", "aeinstein", "Albert", "Einstein", "aeinstein@example.com").
		AddRow(2, "2", "bohrn", "Niels", "Bohr", "bohrn@example.com")

	mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE \"users\".\"deleted_at\" IS NULL").
		WillReturnRows(rows)

	users, err := userRepo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "Albert", users[0].FirstName)
	assert.Equal(t, "Niels", users[1].FirstName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetById(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "cern_person_id", "username", "first_name", "family_name", "email"}).
		AddRow(1, "1", "aeinstein", "Albert", "Einstein", "aeinstein@example.com")

	mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE \"users\".\"id\" = (.+) ORDER BY \"users\".\"id\" LIMIT (.+)").
		WillReturnRows(rows)

	user, err := userRepo.GetByID(1)
	assert.NoError(t, err)
	assert.Equal(t, "Albert", user.FirstName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByCernPersonId(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "cern_person_id", "username", "first_name", "family_name", "email"}).
		AddRow(1, "1", "aeinstein", "Albert", "Einstein", "aeinstein@example.com").
		AddRow(2, "2", "bohrn", "Niels", "Bohr", "bohrn@example.com")

	mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE \"cern_person_id\" = (.+) ORDER BY \"users\".\"id\" LIMIT (.+)").
		WithArgs("2", 1).
		WillReturnRows(rows)

	user, err := userRepo.GetByCernPersonId("2")
	assert.NoError(t, err)
	assert.Equal(t, "Albert", user.FirstName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE \"users\" SET \"deleted_at\"=(.+) WHERE \"users\".\"id\" = (.+) AND \"users\".\"deleted_at\" IS NULL").WithArgs(AnyTime(), 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := userRepo.Delete(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)
	user := &models.User{
		CernPersonId: "1",
		Username:     "aeinstein",
		FirstName:    "Albert",
		FamilyName:   "Einstein",
		Email:        "aeinstein@example.com",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users" (.+) RETURNING "id"`).
		WithArgs(AnyTime(), AnyTime(), AnyTime(), user.CernPersonId, user.Username, user.FirstName, user.FamilyName, user.Email).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := userRepo.Update(user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
