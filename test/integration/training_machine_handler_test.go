package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/handler"
	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestTrainingMachineHandler_List(t *testing.T) {
	router, env, cleanup := setupIntegrationTest(t)
	defer cleanup()

	hasher := service.NewArgon2Hasher()
	tmService := service.NewTrainingMachineService(env.RepositoryContext, hasher)
	handler := handler.NewTrainingMachineHandler(env, tmService)

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, handler.User.Create(user))

	trainingMachine := &models.TrainingMachine{
		Name:   "Machine 1",
		UserId: user.ID,
	}
	assert.NoError(t, handler.TrainingMachine.Create(trainingMachine))

	req, err := http.NewRequest("GET", "/training-machines/list", nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, env, req, user.ID)

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Machine 1")
}

func TestTrainingMachineHandler_Create(t *testing.T) {
	router, env, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, env.User.Create(user))

	trainingMachine := models.TrainingMachine{
		Name:   "New Machine",
		UserId: user.ID,
	}
	body, err := json.Marshal(trainingMachine)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/training-machines", bytes.NewReader(body))
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, env, req, user.ID)

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	responseBody := rr.Body.String()
	assert.Contains(t, responseBody, "SecretKey")
}

func TestTrainingMachineHandler_Show(t *testing.T) {
	router, env, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, env.User.Create(user))

	trainingMachine := &models.TrainingMachine{
		Name:            "Machine 1",
		UserId:          user.ID,
		SecretKeyHashed: "secret",
		LastActivityAt:  time.Now(),
	}
	assert.NoError(t, env.TrainingMachine.Create(trainingMachine))

	req, err := http.NewRequest("GET", fmt.Sprintf("/training-machines/%d", trainingMachine.ID), nil)
	assert.NoError(t, err)
	rr := addSessionCookie(t, env, req, user.ID)

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Machine 1")
}

func TestTrainingMachineHandler_Index(t *testing.T) {
	router, env, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, env.User.Create(user))

	req, err := http.NewRequest("GET", "/training-machines", nil)
	assert.NoError(t, err)
	rr := addSessionCookie(t, env, req, user.ID)

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Training Machines")
}

func TestTrainingMachineHandler_New(t *testing.T) {
	router, env, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, env.User.Create(user))

	req, err := http.NewRequest("GET", "/training-machines/new", nil)
	assert.NoError(t, err)
	rr := addSessionCookie(t, env, req, user.ID)

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Register New Training Machine!")
}

func TestTrainingMachineHandler_Delete(t *testing.T) {
	router, env, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, env.User.Create(user))

	trainingMachine := &models.TrainingMachine{
		Name:   "Machine to Delete",
		UserId: user.ID,
	}
	assert.NoError(t, env.TrainingMachine.Create(trainingMachine))

	req, err := http.NewRequest("DELETE", fmt.Sprintf("/training-machines/%d", trainingMachine.ID), nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, env, req, user.ID)

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	deletedMachine, err := env.TrainingMachine.GetByID(trainingMachine.ID)
	assert.Error(t, err)
	assert.Nil(t, deletedMachine)
}
