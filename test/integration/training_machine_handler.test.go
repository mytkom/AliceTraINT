package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/handler"
	"github.com/stretchr/testify/assert"
)

func TestTrainingMachineHandler_List(t *testing.T) {
	env, cleanup := setupIntegrationTest(t)
	defer cleanup()

	handler := handler.NewTrainingMachineHandler(env)

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, handler.User.Create(user))

	trainingMachine := &models.TrainingMachine{
		Name:   "Machine 1",
		UserId: user.ID,
	}
	assert.NoError(t, handler.TrainingMachine.Create(trainingMachine))

	req, err := http.NewRequest("GET", "/training-machines/list", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	sess := handler.GlobalSessions.SessionStart(rr, req)
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
	env, cleanup := setupIntegrationTest(t)
	defer cleanup()

	handler := handler.NewTrainingMachineHandler(env)

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, handler.User.Create(user))

	trainingMachine := models.TrainingMachine{
		Name: "New Machine",
	}
	body, err := json.Marshal(trainingMachine)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/training-machines", bytes.NewReader(body))
	assert.NoError(t, err)
	rr := addSessionCookie(t, env, req, user.ID)

	handler.Create(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	responseBody := rr.Body.String()
	assert.Contains(t, responseBody, "Secret Key")

	machines, err := handler.TrainingMachine.GetAllUser(user.ID)
	assert.NoError(t, err)
	assert.Len(t, machines, 1)
	assert.Equal(t, "New Machine", machines[0].Name)
}

func TestTrainingMachineHandler_Show(t *testing.T) {
	env, cleanup := setupIntegrationTest(t)
	defer cleanup()

	handler := handler.NewTrainingMachineHandler(env)

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, handler.User.Create(user))

	trainingMachine := &models.TrainingMachine{
		Name:   "Machine 1",
		UserId: user.ID,
	}
	assert.NoError(t, handler.TrainingMachine.Create(trainingMachine))

	req, err := http.NewRequest("GET", "/training-machines/"+strconv.Itoa(int(trainingMachine.ID)), nil)
	assert.NoError(t, err)
	rr := addSessionCookie(t, env, req, user.ID)

	handler.Show(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Machine 1")
}
