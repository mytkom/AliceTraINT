package integration_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/mytkom/AliceTraINT/internal/ccdb"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTrainingTaskHandler_Index(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	req, err := http.NewRequest("GET", "/training-tasks", nil)
	assert.NoError(t, err)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Training Machines")
}

func TestTrainingTaskHandler_List(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	td := models.TrainingDataset{
		Name: "Unique Dataset Name",
		AODFiles: []jalien.AODFile{
			{
				Name:      "AO2D.root",
				Path:      "/alice/sim/2024/LHC24b1b/0/567454/AOD/002/AO2D.root",
				Size:      2312421213,
				LHCPeriod: "LHC24b1b",
				RunNumber: 567454,
				AODNumber: 2,
			},
			{
				Name:      "AO2D.root",
				Path:      "/alice/sim/2024/LHC24b1b/0/567458/AOD/002/AO2D.root",
				Size:      2312421213,
				LHCPeriod: "LHC24b1b",
				RunNumber: 567458,
				AODNumber: 2,
			},
		},
		UserId: 1,
	}
	assert.NoError(t, ut.TrainingDataset.Create(&td))

	trainingTask := &models.TrainingTask{
		Name:              "TrainingTaskcl",
		UserId:            user.ID,
		TrainingDatasetId: td.ID,
		TrainingMachineId: nil,
		Status:            models.Queued,
		Configuration:     "",
	}
	assert.NoError(t, ut.TrainingTask.Create(trainingTask))
	trainingTask2 := &models.TrainingTask{
		Name:              "TrainingTaskclOuher",
		UserId:            user.ID,
		TrainingDatasetId: td.ID,
		TrainingMachineId: nil,
		Status:            models.Queued,
		Configuration:     "",
	}
	assert.NoError(t, ut.TrainingTask.Create(trainingTask2))

	req, err := http.NewRequest("GET", "/training-tasks/list", nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	responseBody := rr.Body.String()
	assert.Contains(t, responseBody, trainingTask.Name)
	assert.Contains(t, responseBody, trainingTask2.Name)
}

func TestTrainingTaskHandler_Show(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	td := models.TrainingDataset{
		Name: "Unique Dataset Name",
		AODFiles: []jalien.AODFile{
			{
				Name:      "AO2D.root",
				Path:      "/alice/sim/2024/LHC24b1b/0/567454/AOD/002/AO2D.root",
				Size:      2312421213,
				LHCPeriod: "LHC24b1b",
				RunNumber: 567454,
				AODNumber: 2,
			},
			{
				Name:      "AO2D.root",
				Path:      "/alice/sim/2024/LHC24b1b/0/567458/AOD/002/AO2D.root",
				Size:      2312421213,
				LHCPeriod: "LHC24b1b",
				RunNumber: 567458,
				AODNumber: 2,
			},
		},
		UserId: 1,
	}
	assert.NoError(t, ut.TrainingDataset.Create(&td))

	trainingTask := &models.TrainingTask{
		Name:              "TrainingTaskcl",
		UserId:            user.ID,
		TrainingDatasetId: td.ID,
		TrainingMachineId: nil,
		Status:            models.Queued,
		Configuration:     "",
	}
	assert.NoError(t, ut.TrainingTask.Create(trainingTask))

	req, err := http.NewRequest("GET", fmt.Sprintf("/training-tasks/%d", trainingTask.ID), nil)
	assert.NoError(t, err)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), trainingTask.Name)
}

func TestTrainingTaskHandler_New(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	td := models.TrainingDataset{
		Name: "Unique Dataset Name",
		AODFiles: []jalien.AODFile{
			{
				Name:      "AO2D.root",
				Path:      "/alice/sim/2024/LHC24b1b/0/567454/AOD/002/AO2D.root",
				Size:      2312421213,
				LHCPeriod: "LHC24b1b",
				RunNumber: 567454,
				AODNumber: 2,
			},
			{
				Name:      "AO2D.root",
				Path:      "/alice/sim/2024/LHC24b1b/0/567458/AOD/002/AO2D.root",
				Size:      2312421213,
				LHCPeriod: "LHC24b1b",
				RunNumber: 567458,
				AODNumber: 2,
			},
		},
		UserId: 1,
	}
	assert.NoError(t, ut.TrainingDataset.Create(&td))

	req, err := http.NewRequest("GET", "/training-tasks/new", nil)
	assert.NoError(t, err)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	responseBody := rr.Body.String()
	assert.Contains(t, responseBody, "Unique Dataset Name")
	assert.Contains(t, responseBody, "fieldName")
	assert.Contains(t, responseBody, "512")
}

func TestTrainingTaskHandler_Create_Success(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	td := models.TrainingDataset{
		Name: "Unique Dataset Name",
		AODFiles: []jalien.AODFile{
			{
				Name:      "AO2D.root",
				Path:      "/alice/sim/2024/LHC24b1b/0/567454/AOD/002/AO2D.root",
				Size:      2312421213,
				LHCPeriod: "LHC24b1b",
				RunNumber: 567454,
				AODNumber: 2,
			},
			{
				Name:      "AO2D.root",
				Path:      "/alice/sim/2024/LHC24b1b/0/567458/AOD/002/AO2D.root",
				Size:      2312421213,
				LHCPeriod: "LHC24b1b",
				RunNumber: 567458,
				AODNumber: 2,
			},
		},
		UserId: 1,
	}
	assert.NoError(t, ut.TrainingDataset.Create(&td))

	trainingTask := &models.TrainingTask{
		Name:              "TrainingTask3",
		UserId:            user.ID,
		TrainingDatasetId: td.ID,
		TrainingMachineId: nil,
		Status:            models.Completed, // it should be overwritten
		Configuration:     "",
	}
	body, err := json.Marshal(trainingTask)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/training-tasks", bytes.NewReader(body))
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	responseBody := rr.Body.String()
	assert.Empty(t, responseBody)

	tts, err := ut.TrainingTask.GetAll()
	assert.NoError(t, err)
	assert.Len(t, tts, 1)
	assert.Equal(t, models.Queued, tts[0].Status)
}

func prepareUploadToCCDB(t *testing.T, ut *IntegrationTestUtils, user *models.User, withOnnxFile bool) *models.TrainingTask {
	td := models.TrainingDataset{
		Name: "Unique Dataset Name",
		AODFiles: []jalien.AODFile{
			{
				Name:      "AO2D.root",
				Path:      "/alice/sim/2024/LHC24b1b/0/567454/AOD/002/AO2D.root",
				Size:      2312421213,
				LHCPeriod: "LHC24b1b",
				RunNumber: 567454,
				AODNumber: 2,
			},
			{
				Name:      "AO2D.root",
				Path:      "/alice/sim/2024/LHC24b1b/0/567458/AOD/002/AO2D.root",
				Size:      2312421213,
				LHCPeriod: "LHC24b1b",
				RunNumber: 567458,
				AODNumber: 2,
			},
		},
		UserId: 1,
	}
	assert.NoError(t, ut.TrainingDataset.Create(&td))

	tm := models.TrainingMachine{
		Name:   "New Machine",
		UserId: user.ID,
	}
	assert.NoError(t, ut.TrainingMachine.Create(&tm))

	trainingTask := &models.TrainingTask{
		Name:              "Task to Upload",
		UserId:            user.ID,
		TrainingDatasetId: td.ID,
		TrainingMachineId: &tm.ID,
		Status:            models.Completed,
		Configuration:     "",
	}
	assert.NoError(t, ut.TrainingTask.Create(trainingTask))

	now := uint64(time.Now().UTC().UnixMilli())
	ut.MockedServices.JAliEn.On("ListAndParseDirectory", "/alice/sim/2024/LHC24b1b/0").Return(&jalien.DirectoryContents{
		Subdirs: []jalien.Dir{
			{Name: "560000", Path: "/alice/sim/2024/LHC24b1b/0/560000"},
			{Name: "567454", Path: "/alice/sim/2024/LHC24b1b/0/567454"},
			{Name: "567458", Path: "/alice/sim/2024/LHC24b1b/0/567458"},
			{Name: "570000", Path: "/alice/sim/2024/LHC24b1b/0/570000"},
		},
	}, nil)
	ut.MockedServices.CCDB.On("GetRunInformation", uint64(560000)).Return(&ccdb.RunInformation{
		RunNumber: 560000,
		SOR:       now - 10000,
		EOR:       now,
	}, nil)
	ut.MockedServices.CCDB.On("GetRunInformation", uint64(570000)).Return(&ccdb.RunInformation{
		RunNumber: 570000,
		SOR:       now,
		EOR:       now + 10000,
	}, nil)

	if withOnnxFile {
		for localName, uploadName := range ut.NNArch.GetExpectedResults().Onnx {
			ttr := models.TrainingTaskResult{
				Name:        "Local file",
				Type:        models.Onnx,
				Description: "some local file",
				File: models.File{
					Path: fmt.Sprintf("./%s", localName),
					Name: localName,
					Size: 12312,
				},
				TrainingTaskId: trainingTask.ID,
			}
			assert.NoError(t, ut.TrainingTaskResult.Create(&ttr))

			ut.FileService.On("OpenFile", ttr.File.Path).Return(nil)
			ut.CCDB.On("UploadFile", now-10000, now+10000, uploadName, mock.Anything).Return(nil)
		}
	} else {
		ut.FileService.On("OpenFile", mock.Anything).Return(errors.New("file do not exist"))
	}

	return trainingTask
}

func TestTrainingTaskHandler_UploadToCCDB_LocalFileNotFound(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	trainingTask := prepareUploadToCCDB(t, ut, user, false)

	req, err := http.NewRequest("POST", fmt.Sprintf("/training-tasks/%d/upload-to-ccdb", trainingTask.ID), nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "local_file.onnx")
}

func TestTrainingTaskHandler_UploadToCCDB_BadStatus(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	trainingTask := prepareUploadToCCDB(t, ut, user, false)
	trainingTask.Status = models.Benchmarking
	assert.NoError(t, ut.TrainingTask.Update(trainingTask))

	req, err := http.NewRequest("POST", fmt.Sprintf("/training-tasks/%d/upload-to-ccdb", trainingTask.ID), nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	assert.Contains(t, rr.Body.String(), "must be completed")
}

func TestTrainingTaskHandler_UploadToCCDB_Success(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	trainingTask := prepareUploadToCCDB(t, ut, user, true)

	req, err := http.NewRequest("POST", fmt.Sprintf("/training-tasks/%d/upload-to-ccdb", trainingTask.ID), nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "local_file.onnx")
}

func TestTrainingTaskHandler_Index_Unauthorized(t *testing.T) {
	testUnauthorized(t, "GET", "/training-tasks", nil)
}
func TestTrainingTaskHandler_List_Unauthorized(t *testing.T) {
	testUnauthorized(t, "GET", "/training-tasks/list", nil)
}

func TestTrainingTaskHandler_Show_Unauthorized(t *testing.T) {
	testUnauthorized(t, "GET", "/training-tasks/1", nil)
}

func TestTrainingTaskHandler_New_Unauthorized(t *testing.T) {
	testUnauthorized(t, "GET", "/training-tasks/new", nil)
}

func TestTrainingTaskHandler_Create_Unauthorized(t *testing.T) {
	trainingMachine := models.TrainingMachine{
		Name:   "New Machine",
		UserId: 1, // Arbitrary user ID
	}
	body, err := json.Marshal(trainingMachine)
	assert.NoError(t, err)

	testUnauthorized(t, "POST", "/training-tasks", body)
}
func TestTrainingTaskHandler_UploadToCCDB_Unauthorized(t *testing.T) {
	testUnauthorized(t, "POST", "/training-tasks/1/upload-to-ccdb", nil)
}
