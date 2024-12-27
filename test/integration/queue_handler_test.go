package integration_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestUserAndTask(t *testing.T, ut *IntegrationTestUtils) (*models.User, *models.TrainingTask, *models.TrainingMachine) {
	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	td := &models.TrainingDataset{
		Name: "Unique Dataset Name",
		AODFiles: []jalien.AODFile{
			{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24b1b/0/567454/AOD/002/AO2D.root", Size: 2312421213, LHCPeriod: "LHC24b1b", RunNumber: 567454, AODNumber: 2},
			{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24b1b/0/567458/AOD/002/AO2D.root", Size: 2312421213, LHCPeriod: "LHC24b1b", RunNumber: 567458, AODNumber: 2},
		},
		UserId: user.ID,
	}
	assert.NoError(t, ut.TrainingDataset.Create(td))

	tm := &models.TrainingMachine{
		Name:            "Unique Training Machine",
		UserId:          user.ID,
		SecretKeyHashed: "secret",
	}
	assert.NoError(t, ut.TrainingMachine.Create(tm))

	ut.Hasher.On("VerifyKey", tm.SecretKeyHashed, tm.SecretKeyHashed).Return(true, nil)
	ut.Hasher.On("VerifyKey", mock.Anything, tm.SecretKeyHashed).Return(false, errors.New("wrong secret key"))

	trainingTask := &models.TrainingTask{
		Name:              "TrainingTaskcl",
		UserId:            user.ID,
		TrainingDataset:   *td,
		TrainingDatasetId: td.ID,
		TrainingMachineId: &tm.ID,
		Status:            models.Queued,
		Configuration:     "",
	}
	assert.NoError(t, ut.TrainingTask.Create(trainingTask))

	return user, trainingTask, tm
}

func TestQueueHandler_UpdateStatus_Success(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	_, tt, tm := setupTestUserAndTask(t, ut)

	body, err := json.Marshal(map[string]uint{"Status": uint(models.Benchmarking)})
	assert.NoError(t, err)

	req := newRequest(t, "POST", fmt.Sprintf("/training-tasks/%d/status", tt.ID), body, tm.SecretKeyHashed)

	rr := httptest.NewRecorder()
	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, rr.Body.String())
}

type queryTaskResponse struct {
	ID       uint
	AODFiles []jalien.AODFile
}

func TestQueueHandler_QueryTask_Success(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	_, tt, tm := setupTestUserAndTask(t, ut)

	queuedTrainingTask := &models.TrainingTask{
		Name:              "Queued training task",
		TrainingDatasetId: tt.TrainingDatasetId,
		Status:            models.Queued,
		UserId:            tt.UserId,
	}
	assert.NoError(t, ut.TrainingTask.Create(queuedTrainingTask))

	req := newRequest(t, "GET", fmt.Sprintf("/training-machines/%d/training-task", tm.ID), nil, tm.SecretKeyHashed)

	rr := httptest.NewRecorder()
	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp queryTaskResponse
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, tt.ID, resp.ID)
	assert.True(t, reflect.DeepEqual(resp.AODFiles, tt.TrainingDataset.AODFiles))
}

func TestQueueHandler_CreateTrainingTaskResult_Success(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	_, tt, tm := setupTestUserAndTask(t, ut)

	ttr := models.TrainingTaskResult{
		Name:        "Local file",
		Type:        models.Onnx,
		Description: "some local file",
		File:        models.File{Name: "file.txt", Path: "./file.txt", Size: 432},
	}
	ut.FileService.On("SaveFile", mock.Anything, mock.Anything).Return(&ttr.File, nil)

	buf, mw := prepareMultipartData(t, map[string]string{
		"name":        ttr.Name,
		"description": ttr.Description,
		"file-type":   fmt.Sprintf("%d", uint(ttr.Type)),
	}, "file", "file.txt", []byte("Test file"))

	req := newRequestWithMultipart(t, "POST", fmt.Sprintf("/training-tasks/%d/training-task-results", tt.ID), &buf, mw.FormDataContentType(), tm.SecretKeyHashed)

	rr := httptest.NewRecorder()
	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resTtr models.TrainingTaskResult
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resTtr))
	assert.Equal(t, ttr.Name, resTtr.Name)
	assert.Equal(t, ttr.Description, resTtr.Description)
	assert.Equal(t, ttr.Type, resTtr.Type)
	assert.Equal(t, ttr.File.Path, resTtr.File.Path)
}

func TestQueueHandler_UpdateStatus_Unauthorized(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	_, tt, _ := setupTestUserAndTask(t, ut)

	body, err := json.Marshal(map[string]uint{"Status": uint(models.Benchmarking)})
	assert.NoError(t, err)

	req := newRequest(t, "POST", fmt.Sprintf("/training-tasks/%d/status", tt.ID), body, "bad-secret")

	rr := httptest.NewRecorder()
	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "unauthorized machine")
}

func TestQueueHandler_QueryTask_Unauthorized(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	_, _, tm := setupTestUserAndTask(t, ut)

	req := newRequest(t, "GET", fmt.Sprintf("/training-machines/%d/training-task", tm.ID), nil, "bad-secret")

	rr := httptest.NewRecorder()
	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "unauthorized machine")
}

func TestQueueHandler_CreateTrainingTaskResult_Unauthorized(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	_, tt, _ := setupTestUserAndTask(t, ut)

	buf, mw := prepareMultipartData(t, map[string]string{
		"name":        "Local file",
		"description": "some local file",
		"file-type":   fmt.Sprintf("%d", uint(models.Onnx)),
	}, "file", "file.txt", []byte("Test file"))

	req := newRequestWithMultipart(t, "POST", fmt.Sprintf("/training-tasks/%d/training-task-results", tt.ID), &buf, mw.FormDataContentType(), "bad-secret")

	rr := httptest.NewRecorder()
	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "unauthorized machine")
}

func newRequest(t *testing.T, method, url string, body []byte, secret string) *http.Request {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	assert.NoError(t, err)
	req.Header.Add("Secret-Id", secret)
	return req
}

func newRequestWithMultipart(t *testing.T, method, url string, body *bytes.Buffer, contentType, secret string) *http.Request {
	req, err := http.NewRequest(method, url, body)
	assert.NoError(t, err)
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Secret-Id", secret)
	return req
}

func prepareMultipartData(t *testing.T, fields map[string]string, fileField, fileName string, fileData []byte) (bytes.Buffer, *multipart.Writer) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	for key, value := range fields {
		fieldWriter, err := mw.CreateFormField(key)
		assert.NoError(t, err)
		_, err = fieldWriter.Write([]byte(value))
		assert.NoError(t, err)
	}

	fileWriter, err := mw.CreateFormFile(fileField, fileName)
	assert.NoError(t, err)
	_, err = fileWriter.Write(fileData)
	assert.NoError(t, err)

	assert.NoError(t, mw.Close())
	return buf, mw
}
