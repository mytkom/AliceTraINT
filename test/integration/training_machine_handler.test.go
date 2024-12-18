package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/handler"
	"github.com/mytkom/AliceTraINT/internal/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTrainingMachineIntegrationTest(t *testing.T) (*handler.TrainingMachineHandler, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.TrainingMachine{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	trainingMachineRepo := repository.NewTrainingMachineRepository(db)
	userRepo := repository.NewUserRepository(db)
	tmpl := utils.BaseTemplate()
	auth := auth.MockAuth()

	handler := &handler.TrainingMachineHandler{
		TrainingMachineRepo: trainingMachineRepo,
		UserRepo:            userRepo,
		Auth:                auth,
		Template:            tmpl,
	}

	cleanup := func() {
		dbSQL, err := db.DB()
		if err == nil {
			dbSQL.Close()
		}
	}

	return handler, cleanup
}

func TestTrainingMachineHandler_List(t *testing.T) {
	handler, cleanup := setupTrainingMachineIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, handler.UserRepo.Create(user))

	trainingMachine := &models.TrainingMachine{
		Name:   "Machine 1",
		UserId: user.ID,
	}
	assert.NoError(t, handler.TrainingMachineRepo.Create(trainingMachine))

	req, err := http.NewRequest("GET", "/training-machines/list", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	sess := handler.Auth.GlobalSessions.SessionStart(rr, req)
	assert.NoError(t, sess.Set("loggedUserId", user.ID))

	cookie := &http.Cookie{
		Name:  "gosessionid",
		Value: sess.SessionID(),
	}
	req.AddCookie(cookie)

	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Machine 1")
}

func TestTrainingMachineHandler_Create(t *testing.T) {
	handler, cleanup := setupTrainingMachineIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, handler.UserRepo.Create(user))

	trainingMachine := models.TrainingMachine{
		Name: "New Machine",
	}
	body, err := json.Marshal(trainingMachine)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/training-machines", bytes.NewReader(body))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	sess := handler.Auth.GlobalSessions.SessionStart(rr, req)
	assert.NoError(t, sess.Set("loggedUserId", user.ID))

	cookie := &http.Cookie{
		Name:  "gosessionid",
		Value: sess.SessionID(),
	}
	req.AddCookie(cookie)

	handler.Create(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	responseBody := rr.Body.String()
	assert.Contains(t, responseBody, "Secret Key")

	machines, err := handler.TrainingMachineRepo.GetAllUser(user.ID)
	assert.NoError(t, err)
	assert.Len(t, machines, 1)
	assert.Equal(t, "New Machine", machines[0].Name)
}

func TestTrainingMachineHandler_Show(t *testing.T) {
	handler, cleanup := setupTrainingMachineIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, handler.UserRepo.Create(user))

	trainingMachine := &models.TrainingMachine{
		Name:   "Machine 1",
		UserId: user.ID,
	}
	assert.NoError(t, handler.TrainingMachineRepo.Create(trainingMachine))

	req, err := http.NewRequest("GET", "/training-machines/"+strconv.Itoa(int(trainingMachine.ID)), nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	sess := handler.Auth.GlobalSessions.SessionStart(rr, req)
	assert.NoError(t, sess.Set("loggedUserId", user.ID))

	cookie := &http.Cookie{
		Name:  "gosessionid",
		Value: sess.SessionID(),
	}
	req.AddCookie(cookie)

	handler.Show(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Machine 1")
}
