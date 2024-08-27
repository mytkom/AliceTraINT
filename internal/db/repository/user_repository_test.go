package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	dialector := postgres.New(postgres.Config{
		DSN:                  "sqlmock_db_0",
		Conn:                 db,
		PreferSimpleProtocol: true,
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open gorm DB: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return gormDB, mock, cleanup
}

func TestUserRepository_CreateUser(t *testing.T) {
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
		WithArgs(user.CernPersonId, user.Username, user.FirstName, user.FamilyName, user.Email).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := userRepo.CreateUser(user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetAllUsers(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "cern_person_id", "username", "first_name", "family_name", "email"}).
		AddRow(1, "1", "aeinstein", "Albert", "Einstein", "aeinstein@example.com").
		AddRow(2, "2", "bohrn", "Niels", "Bohr", "bohrn@example.com")

	mock.ExpectQuery("SELECT \\* FROM \"users\"").
		WillReturnRows(rows)

	users, err := userRepo.GetAllUsers()
	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, "Albert", users[0].FirstName)
	assert.Equal(t, "Niels", users[1].FirstName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserById(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "cern_person_id", "username", "first_name", "family_name", "email"}).
		AddRow(1, "1", "aeinstein", "Albert", "Einstein", "aeinstein@example.com").
		AddRow(2, "2", "bohrn", "Niels", "Bohr", "bohrn@example.com")

	mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE \"users\".\"id\" = (.+) ORDER BY \"users\".\"id\" LIMIT (.+)").
		WillReturnRows(rows)

	user, err := userRepo.GetUserByID(2)
	assert.NoError(t, err)
	assert.Equal(t, "Albert", user.FirstName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByCernPersonId(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "cern_person_id", "username", "first_name", "family_name", "email"}).
		AddRow(1, "1", "aeinstein", "Albert", "Einstein", "aeinstein@example.com").
		AddRow(2, "2", "bohrn", "Niels", "Bohr", "bohrn@example.com")

	mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE \"cern_person_id\" = (.+) ORDER BY \"users\".\"id\" LIMIT (.+)").
		WithArgs("2", 1).
		WillReturnRows(rows)

	user, err := userRepo.GetUserByCernPersonId("2")
	assert.NoError(t, err)
	assert.Equal(t, "Albert", user.FirstName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_DeleteUser(t *testing.T) {
	db, mock, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := NewUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM \"users\" WHERE \"users\".\"id\" = (.+)").WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := userRepo.DeleteUser(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateUser(t *testing.T) {
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
		WithArgs(user.CernPersonId, user.Username, user.FirstName, user.FamilyName, user.Email).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := userRepo.UpdateUser(user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
