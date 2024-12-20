package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/assert"
)

func TestTrainingMachineHandler_Index(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	req, err := http.NewRequest("GET", "/training-machines", nil)
	assert.NoError(t, err)
	rr := addSessionCookie(t, ut.Env, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Training Machines")
}

func TestTrainingMachineHandler_List(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	trainingMachine := &models.TrainingMachine{
		Name:   "Machine 1",
		UserId: user.ID,
	}
	assert.NoError(t, ut.TrainingMachine.Create(trainingMachine))

	req, err := http.NewRequest("GET", "/training-machines/list", nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Env, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Machine 1")
}

func TestTrainingMachineHandler_Show(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	trainingMachine := &models.TrainingMachine{
		Name:            "Machine 1",
		UserId:          user.ID,
		SecretKeyHashed: "secret",
		LastActivityAt:  time.Now(),
	}
	assert.NoError(t, ut.TrainingMachine.Create(trainingMachine))

	req, err := http.NewRequest("GET", fmt.Sprintf("/training-machines/%d", trainingMachine.ID), nil)
	assert.NoError(t, err)
	rr := addSessionCookie(t, ut.Env, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Machine 1")
}

func TestTrainingMachineHandler_New(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	req, err := http.NewRequest("GET", "/training-machines/new", nil)
	assert.NoError(t, err)
	rr := addSessionCookie(t, ut.Env, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Register New Training Machine!")
}

func TestTrainingMachineHandler_Create(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	trainingMachine := models.TrainingMachine{
		Name:   "New Machine",
		UserId: user.ID,
	}
	body, err := json.Marshal(trainingMachine)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/training-machines", bytes.NewReader(body))
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Env, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	responseBody := rr.Body.String()
	assert.Contains(t, responseBody, "SecretKey")
}

func TestTrainingMachineHandler_Delete(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	trainingMachine := &models.TrainingMachine{
		Name:   "Machine to Delete",
		UserId: user.ID,
	}
	assert.NoError(t, ut.TrainingMachine.Create(trainingMachine))

	req, err := http.NewRequest("DELETE", fmt.Sprintf("/training-machines/%d", trainingMachine.ID), nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Env, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	deletedMachine, err := ut.TrainingMachine.GetByID(trainingMachine.ID)
	assert.Error(t, err)
	assert.Nil(t, deletedMachine)
}

func TestTrainingMachineHandler_Index_Unauthorized(t *testing.T) {
	testUnauthorized(t, "GET", "/training-machines", nil)
}
func TestTrainingMachineHandler_List_Unauthorized(t *testing.T) {
	testUnauthorized(t, "GET", "/training-machines/list", nil)
}

func TestTrainingMachineHandler_Show_Unauthorized(t *testing.T) {
	testUnauthorized(t, "GET", "/training-machines/1", nil)
}

func TestTrainingMachineHandler_New_Unauthorized(t *testing.T) {
	testUnauthorized(t, "GET", "/training-machines/new", nil)
}

func TestTrainingMachineHandler_Create_Unauthorized(t *testing.T) {
	trainingMachine := models.TrainingMachine{
		Name:   "New Machine",
		UserId: 1, // Arbitrary user ID
	}
	body, err := json.Marshal(trainingMachine)
	assert.NoError(t, err)

	testUnauthorized(t, "POST", "/training-machines", body)
}
func TestTrainingMachineHandler_Delete_Unauthorized(t *testing.T) {
	testUnauthorized(t, "DELETE", "/training-machines/1", nil)
}
