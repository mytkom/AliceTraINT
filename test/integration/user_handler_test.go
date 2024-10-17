package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/handler"
	"github.com/mytkom/AliceTraINT/internal/utils"
	"github.com/stretchr/testify/assert"
	_ "github.com/thomasdarimont/go-kc-example/session_memory"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Make paths relative to root dir when running tests
func init() {
	if err := os.Chdir("../.."); err != nil {
		panic(err)
	}
}

func setupIntegrationTest(t *testing.T) (*handler.UserHandler, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	userRepo := repository.NewUserRepository(db)

	tmpl := utils.BaseTemplate()

	auth := auth.MockAuth()

	handler := handler.NewUserHandler(tmpl, userRepo, auth)

	cleanup := func() {
		dbSQL, err := db.DB()
		if err == nil {
			dbSQL.Close()
		}
	}

	return handler, cleanup
}

func TestUserHandler_Integration_Index(t *testing.T) {
	handler, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Seed the database with a user
	err := handler.UserRepo.Create(&models.User{
		CernPersonId: "12345",
		Username:     "johndoe",
		FirstName:    "John",
		FamilyName:   "Doe",
		Email:        "johndoe@example.com",
	})
	assert.NoError(t, err)

	// Create an HTTP request to the index route
	req, err := http.NewRequest("GET", "/", strings.NewReader(""))
	assert.NoError(t, err)

	// Record the response
	rr := httptest.NewRecorder()

	// Mock the session to simulate a logged-in user
	sess := handler.Auth.GlobalSessions.SessionStart(rr, req)
	err = sess.Set("loggedUserId", uint(1))
	assert.NoError(t, err)

	cookie := &http.Cookie{
		Name:  "gosessionid",
		Value: strings.Split(strings.Split(rr.Header().Get("Set-Cookie"), ";")[0], "=")[1],
	}
	req.AddCookie(cookie)

	handler.Index(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response body
	assert.Contains(t, rr.Body.String(), "John (johndoe@example.com)")
}

func TestUserHandler_Integration_CreateUser(t *testing.T) {
	handler, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Create a POST request to create a new user
	form := strings.NewReader("cern-person-id=67890&username=janedoe&first-name=Jane&family-name=Doe&email=janedoe@example.com")
	req, err := http.NewRequest("POST", "/users", form)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Record the response
	rr := httptest.NewRecorder()
	handler.CreateUser(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify that the user was created in the database
	users, err := handler.UserRepo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "Jane Doe", users[0].FirstName+" "+users[0].FamilyName)
	assert.Equal(t, "janedoe@example.com", users[0].Email)
	assert.Equal(t, "janedoe", users[0].Username)
}
