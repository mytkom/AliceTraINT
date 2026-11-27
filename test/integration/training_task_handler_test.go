package integration_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/mytkom/AliceTraINT/internal/ccdb"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
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
	assert.Contains(t, rr.Body.String(), "Machines")
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

func TestTrainingTaskListQuery(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	owner := models.User{CernPersonId: "owner", Username: "aeinstein", FirstName: "Albert", FamilyName: "Einstein", Email: "owner@example.test"}
	other := models.User{CernPersonId: "other", Username: "other", Email: "other@example.test"}
	require.NoError(t, ut.User.Create(&owner))
	require.NoError(t, ut.User.Create(&other))
	dataset := models.TrainingDataset{Name: "Shared dataset", UserId: other.ID}
	require.NoError(t, ut.TrainingDataset.Create(&dataset))
	require.NoError(t, ut.TrainingDataset.Delete(other.ID, dataset.ID))
	otherDataset := models.TrainingDataset{Name: "Other dataset", UserId: other.ID}
	require.NoError(t, ut.TrainingDataset.Create(&otherDataset))

	day := time.Date(2024, 11, 20, 0, 0, 0, 0, time.UTC)
	create := func(name string, userID uint, created time.Time, status models.TrainingTaskStatus) uint {
		t.Helper()
		task := models.TrainingTask{
			Model: gorm.Model{CreatedAt: created}, Name: name, UserId: userID,
			TrainingDatasetId: dataset.ID, Status: status, Configuration: map[string]int{"unused": 123},
		}
		require.NoError(t, ut.TrainingTask.Create(&task))
		return task.ID
	}

	const count = repository.TrainingTaskPageSize + 2
	var ids []uint
	for i := 0; i < count; i++ {
		ids = append(ids, create(fmt.Sprintf("Task_%% %02d", i), owner.ID, day, models.Failed))
	}
	create("TaskXY 91", owner.ID, day, models.Failed)
	create("Task_% 92", other.ID, day, models.Failed)
	create("Task_% 93", owner.ID, day.AddDate(0, 0, 1), models.Failed)
	create("Task_% 94", owner.ID, day.Add(-time.Second), models.Failed)
	create("Task_% 95", owner.ID, day, models.Queued)
	deleted := create("Task_% 96", owner.ID, day, models.Failed)
	require.NoError(t, ut.TrainingTask.Delete(owner.ID, deleted))
	require.NoError(t, ut.TrainingTask.Create(&models.TrainingTask{
		Model: gorm.Model{CreatedAt: day}, Name: "Task_% 97", UserId: owner.ID,
		TrainingDatasetId: otherDataset.ID, Status: models.Failed,
	}))

	choices, err := ut.TrainingTask.FilterChoices("dataset", "OTHER", dataset.ID)
	require.NoError(t, err)
	assert.Equal(t, []repository.TrainingTaskFilterChoice{
		{ID: dataset.ID, Name: dataset.Name}, {ID: otherDataset.ID, Name: otherDataset.Name},
	}, choices)

	owners, err := ut.TrainingTask.FilterChoices("owner", "EINSTEIN", 0)
	require.NoError(t, err)
	assert.Equal(t, []repository.TrainingTaskFilterChoice{
		{ID: owner.ID, Name: "Albert Einstein (aeinstein)"},
	}, owners)

	failed := models.Failed
	query := repository.TrainingTaskListQuery{
		DatasetID: dataset.ID, OwnerID: owner.ID, Name: "TASK_%", Status: &failed,
		CreatedFrom: day, CreatedBefore: day.AddDate(0, 0, 1), Sort: "oldest", Page: 1,
	}
	first, err := ut.TrainingTask.List(query)
	require.NoError(t, err)
	require.True(t, first.HasNext)
	require.Len(t, first.Tasks, repository.TrainingTaskPageSize)
	for index, task := range first.Tasks {
		assert.Equal(t, ids[index], task.ID)
		assert.Equal(t, dataset.Name, task.TrainingDataset.Name)
		assert.True(t, task.TrainingDataset.DeletedAt.Valid)
		assert.Nil(t, task.Configuration)
	}

	query.Page = 2
	last, err := ut.TrainingTask.List(query)
	require.NoError(t, err)
	require.Len(t, last.Tasks, 2)
	assert.False(t, last.HasNext)
	assert.Equal(t, ids[count-2], last.Tasks[0].ID)
	assert.Equal(t, ids[count-1], last.Tasks[1].ID)

	query.Page = 3
	empty, err := ut.TrainingTask.List(query)
	require.NoError(t, err)
	assert.Empty(t, empty.Tasks)
	assert.False(t, empty.HasNext)

	req, err := http.NewRequest("GET", fmt.Sprintf("/training-tasks/list?name=TASK_%%25&owner=%d&status=0&from=20/11/2024&to=2024-11-20&sort=oldest&dataset=%d&page=2", owner.ID, dataset.ID), nil)
	require.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, owner.ID)
	ut.Router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 2, strings.Count(rr.Body.String(), `class="task-name"`))
	assert.Contains(t, rr.Body.String(), ">Task_% 25<")
	assert.Contains(t, rr.Body.String(), ">Task_% 26<")
	assert.Contains(t, rr.Body.String(), `rel="prev"`)
	assert.Contains(t, rr.Body.String(), `from=20%2F11%2F2024`)
	assert.Contains(t, rr.Body.String(), `to=20%2F11%2F2024`)
	assert.NotContains(t, rr.Body.String(), `rel="next"`)
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
	assert.Contains(t, responseBody, "configuration.training.batch_size")
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
		Name:                "TrainingTask3",
		UserId:              user.ID,
		TrainingDatasetId:   td.ID,
		TrainingMachineId:   nil,
		Status:              models.Completed, // it should be overwritten
		Configuration:       "",
		SubsampleEventCount: 400000,
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
	ut.MockedServices.Monalisa.On("GetRunList").Return(nil, errors.New("monalisa down"))
	ut.MockedServices.CCDB.On("GetRunInformation", uint64(567454)).Return(&ccdb.RunInformation{
		RunNumber: 567454,
		SOR:       now - 10000,
		EOR:       now,
	}, nil)
	ut.MockedServices.CCDB.On("GetRunInformation", uint64(567458)).Return(&ccdb.RunInformation{
		RunNumber: 567458,
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

func TestTrainingTaskHandler_Restart(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	owner := &models.User{CernPersonId: "12345", Username: "owner", Email: "owner@example.com"}
	other := &models.User{CernPersonId: "67890", Username: "other", Email: "other@example.com"}
	require.NoError(t, ut.User.Create(owner))
	require.NoError(t, ut.User.Create(other))
	dataset := &models.TrainingDataset{Name: "Restart dataset", UserId: owner.ID}
	require.NoError(t, ut.TrainingDataset.Create(dataset))

	machine := &models.TrainingMachine{Name: "Previous machine", UserId: owner.ID}
	require.NoError(t, ut.TrainingMachine.Create(machine))
	task := &models.TrainingTask{
		Name: "Failed task", UserId: owner.ID, TrainingDatasetId: dataset.ID,
		TrainingMachineId: &machine.ID, Status: models.Failed,
		Configuration: map[string]interface{}{"data": map[string]interface{}{"use_event_data": false}},
	}
	require.NoError(t, ut.TrainingTask.Create(task))

	taskURL := fmt.Sprintf("/training-tasks/%d", task.ID)
	for _, attempt := range []struct {
		userID     uint
		canRestart bool
		status     int
	}{
		{other.ID, false, http.StatusNotFound},
		{owner.ID, true, http.StatusOK},
		{owner.ID, false, http.StatusNotFound},
	} {
		req, err := http.NewRequest(http.MethodGet, taskURL, nil)
		require.NoError(t, err)
		rr := addSessionCookie(t, ut.Auth, req, attempt.userID)
		ut.Router.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, attempt.canRestart, strings.Contains(rr.Body.String(), "Restart task"))

		req, err = http.NewRequest(http.MethodPost, taskURL+"/restart", nil)
		require.NoError(t, err)
		HTMXReq(req)
		rr = addSessionCookie(t, ut.Auth, req, attempt.userID)
		ut.Router.ServeHTTP(rr, req)
		assert.Equal(t, attempt.status, rr.Code)
	}

	queued, err := ut.TrainingTask.GetFirstQueued()
	require.NoError(t, err)
	assert.Equal(t, task.ID, queued.ID)
	assert.Nil(t, queued.TrainingMachineId)
	assert.Equal(t, task.Configuration, queued.Configuration)
}
