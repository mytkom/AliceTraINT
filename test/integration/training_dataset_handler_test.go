package integration_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/stretchr/testify/assert"
)

func TestTrainingDatasetHandler_Index(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	req, err := http.NewRequest("GET", "/training-datasets", nil)
	assert.NoError(t, err)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Training Datasets")
}

func TestTrainingDatasetHandler_List_All(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1", Email: "1@gmail.com"}
	assert.NoError(t, ut.User.Create(user))

	otherUser := &models.User{CernPersonId: "123456", Username: "user12", Email: "2@gmail.com"}
	assert.NoError(t, ut.User.Create(otherUser))

	datasets := []models.TrainingDataset{
		{Name: "LHC24b1b", UserId: user.ID, AODFiles: []jalien.AODFile{}},
		{Name: "LHC24b1b2", UserId: otherUser.ID, AODFiles: []jalien.AODFile{}},
	}
	for _, dataset := range datasets {
		assert.NoError(t, ut.TrainingDataset.Create(&dataset))
	}

	req, err := http.NewRequest("GET", "/training-datasets/list", nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "LHC24b1b")
	assert.Contains(t, rr.Body.String(), "LHC24b1b2")
}

func TestTrainingDatasetHandler_List_UserScoped(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1", Email: "1@gmail.com"}
	assert.NoError(t, ut.User.Create(user))

	otherUser := &models.User{CernPersonId: "123456", Username: "user12", Email: "2@gmail.com"}
	assert.NoError(t, ut.User.Create(otherUser))

	datasets := []models.TrainingDataset{
		{Name: "LHC24b1b", UserId: user.ID, AODFiles: []jalien.AODFile{}},
		{Name: "LHC24b1b2", UserId: otherUser.ID, AODFiles: []jalien.AODFile{}},
	}
	for _, dataset := range datasets {
		assert.NoError(t, ut.TrainingDataset.Create(&dataset))
	}

	req, err := http.NewRequest("GET", "/training-datasets/list?userScoped=on", nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "LHC24b1b")
	assert.NotContains(t, rr.Body.String(), "LHC24b1b2")
}

func TestTrainingDatasetHandler_Show(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1", Email: "1@gmail.com"}
	assert.NoError(t, ut.User.Create(user))

	otherUser := &models.User{CernPersonId: "123456", Username: "user12", Email: "2@gmail.com"}
	assert.NoError(t, ut.User.Create(otherUser))

	datasets := []models.TrainingDataset{
		{Name: "LHC24b1b", UserId: user.ID, AODFiles: []jalien.AODFile{}},
		{Name: "LHC24b1b2", UserId: otherUser.ID, AODFiles: []jalien.AODFile{}},
	}
	for _, dataset := range datasets {
		assert.NoError(t, ut.TrainingDataset.Create(&dataset))
	}

	req, err := http.NewRequest("GET", "/training-datasets/2", nil)
	assert.NoError(t, err)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "LHC24b1b2")
}

func TestTrainingDatasetHandler_Show_NotFound(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1", Email: "1@gmail.com"}
	assert.NoError(t, ut.User.Create(user))

	datasets := []models.TrainingDataset{
		{Name: "LHC24b1b", UserId: user.ID, AODFiles: []jalien.AODFile{}},
	}
	for _, dataset := range datasets {
		assert.NoError(t, ut.TrainingDataset.Create(&dataset))
	}

	req, err := http.NewRequest("GET", "/training-datasets/2", nil)
	assert.NoError(t, err)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "TrainingDataset not found")
}

func TestTrainingDatasetHandler_New(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	req, err := http.NewRequest("GET", "/training-datasets/new", nil)
	assert.NoError(t, err)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Create New Training Dataset")
}

type createTrainingDatasetPayload struct {
	Name     string
	AODFiles []jalien.AODFile
	UserId   uint
}

func TestTrainingDatasetHandler_Create_Failure(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	td := models.TrainingDataset{
		Name: "Unique Dataset Name",
		AODFiles: []jalien.AODFile{{
			Name:      "AO2D.root",
			Path:      "/alice/sim/2024/LHC24b1b/0/567454/AOD/002/AO2D.root",
			Size:      2312421213,
			LHCPeriod: "LHC24b1b",
			RunNumber: 567454,
			AODNumber: 2,
		}},
		UserId: 1,
	}
	assert.NoError(t, ut.TrainingDataset.Create(&td))

	tdSameName := createTrainingDatasetPayload{
		Name: "Unique Dataset Name",
		AODFiles: []jalien.AODFile{{
			Name:      "AO2D.root",
			Path:      "/alice/sim/2024/LHC24b1b/0/567456/AOD/002/AO2D.root",
			Size:      33425234,
			LHCPeriod: "LHC24b1b",
			RunNumber: 567456,
			AODNumber: 2,
		}},
		UserId: 1,
	}
	body, err := json.Marshal(tdSameName)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/training-datasets", bytes.NewReader(body))
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

	responseBody := rr.Body.String()
	assert.Contains(t, responseBody, "Name must be unique")
}

func TestTrainingDatasetHandler_Create(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	trainingDataset := models.TrainingDataset{
		Name: "New Machine",
		AODFiles: []jalien.AODFile{{
			Name:      "AO2D.root",
			Path:      "/alice/sim/2024/LHC24b1b/0/567454/AOD/002/AO2D.root",
			Size:      2312421213,
			LHCPeriod: "LHC24b1b",
			RunNumber: 567454,
			AODNumber: 2,
		}},
		UserId: 1,
	}
	body, err := json.Marshal(trainingDataset)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/training-datasets", bytes.NewReader(body))
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	responseBody := rr.Body.String()
	assert.Empty(t, responseBody)
	assert.Equal(t, "/training-datasets", rr.Header().Get("Hx-Redirect"))
}

func TestTrainingDatasetHandler_Delete(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	trainingDataset := models.TrainingDataset{
		Name: "New Machine",
		AODFiles: []jalien.AODFile{{
			Name:      "AO2D.root",
			Path:      "/alice/sim/2024/LHC24b1b/0/567454/AOD/002/AO2D.root",
			Size:      2312421213,
			LHCPeriod: "LHC24b1b",
			RunNumber: 567454,
			AODNumber: 2,
		}},
		UserId: 1, // Arbitrary user ID
	}
	assert.NoError(t, ut.TrainingDataset.Create(&trainingDataset))

	req, err := http.NewRequest("DELETE", fmt.Sprintf("/training-datasets/%d", trainingDataset.ID), nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	deletedMachine, err := ut.TrainingMachine.GetByID(trainingDataset.ID)
	assert.Error(t, err)
	assert.Nil(t, deletedMachine)
}

func TestTrainingDatasetHandler_ExploreDirectory_Success(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	path := "/some/path"
	ut.JAliEn.On("ListAndParseDirectory", path).Return(&jalien.DirectoryContents{
		AODFiles: []jalien.AODFile{{
			Name:      "AO2D.root",
			Path:      "/some/path/AO2D.root",
			Size:      2312421213,
			LHCPeriod: "LHC24b1b",
			RunNumber: 567454,
			AODNumber: 2,
		}},
		OtherFiles: []jalien.File{{
			Name: "other_file.ext",
			Path: "some/path/other_file.ext",
			Size: 4096,
		}},
		Subdirs: []jalien.Dir{
			{
				Name: "subdir1",
				Path: "/some/path/subdir1",
			},
			{
				Name: "subdir2",
				Path: "/some/path/subdir2",
			},
		},
	}, nil)

	req, err := http.NewRequest("GET", fmt.Sprintf("/training-datasets/explore-directory?path=%s", path), nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	responseBody := rr.Body.String()
	assert.Contains(t, responseBody, "/some/path")
	assert.Contains(t, responseBody, "subdir1")
	assert.Contains(t, responseBody, "subdir2")
	assert.NotContains(t, responseBody, "LHC24b1b")
	assert.NotContains(t, responseBody, "AO2D.root")
	assert.NotContains(t, responseBody, "other_file.ext")
}

func TestTrainingDatasetHandler_ExploreDirectory_InternalFailure(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	path := "/some/path"
	ut.JAliEn.On("ListAndParseDirectory", path).Return(nil, errors.New("cannot obtain directory contents"))

	req, err := http.NewRequest("GET", fmt.Sprintf("/training-datasets/explore-directory?path=%s", path), nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	responseBody := rr.Body.String()
	assert.Equal(t, "unexpected internal server error\n", responseBody)
}

func TestTrainingDatasetHandler_ExploreDirectory_CCDBUnreachable(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	path := "/some/path"
	ut.JAliEn.On("ListAndParseDirectory", path).Return(nil, NewMockTimeoutError())

	req, err := http.NewRequest("GET", fmt.Sprintf("/training-datasets/explore-directory?path=%s", path), nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	responseBody := rr.Body.String()
	assert.Contains(t, responseBody, `"CCDB" external service is unreachable`)
}

func TestTrainingDatasetHandler_FindAods_Success(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	path := "/some/path"
	ut.JAliEn.On("FindAODFiles", path).Return([]jalien.AODFile{
		{
			Name:      "AO2D.root",
			Path:      "/some/path/001/AO2D.root",
			Size:      2312421213,
			LHCPeriod: "LHC24b1b",
			RunNumber: 567454,
			AODNumber: 1,
		},
		{
			Name:      "AO2D.root",
			Path:      "/some/path/002/AO2D.root",
			Size:      10000000,
			LHCPeriod: "LHC24c1",
			RunNumber: 567451,
			AODNumber: 2,
		},
	}, nil)

	req, err := http.NewRequest("GET", fmt.Sprintf("/training-datasets/find-aods?path=%s", path), nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	responseBody := rr.Body.String()
	assert.Contains(t, responseBody, "LHC24b1b")
	assert.Contains(t, responseBody, "567454")
	assert.Contains(t, responseBody, "LHC24c1")
	assert.Contains(t, responseBody, "567451")
}

func TestTrainingDatasetHandler_FindAods_InternalFailure(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	path := "/some/path"
	ut.JAliEn.On("FindAODFiles", path).Return(nil, errors.New("cannot execute find command"))

	req, err := http.NewRequest("GET", fmt.Sprintf("/training-datasets/find-aods?path=%s", path), nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	responseBody := rr.Body.String()
	assert.Equal(t, "unexpected internal server error\n", responseBody)
}

func TestTrainingDatasetHandler_FindAods_CCDBUnreachable(t *testing.T) {
	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	user := &models.User{CernPersonId: "12345", Username: "user1"}
	assert.NoError(t, ut.User.Create(user))

	path := "/some/path"
	ut.JAliEn.On("FindAODFiles", path).Return(nil, NewMockTimeoutError())

	req, err := http.NewRequest("GET", fmt.Sprintf("/training-datasets/find-aods?path=%s", path), nil)
	assert.NoError(t, err)
	HTMXReq(req)
	rr := addSessionCookie(t, ut.Auth, req, user.ID)

	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	responseBody := rr.Body.String()
	assert.Contains(t, responseBody, `"CCDB" external service is unreachable`)
}

func TestTrainingDatasetHandler_Index_Unauthorized(t *testing.T) {
	testUnauthorized(t, "GET", "/training-datasets", nil)
}

func TestTrainingDatasetHandler_List_Unauthorized(t *testing.T) {
	testUnauthorized(t, "GET", "/training-datasets/list", nil)
}

func TestTrainingDatasetHandler_Show_Unauthorized(t *testing.T) {
	testUnauthorized(t, "GET", "/training-datasets/1", nil)
}

func TestTrainingDatasetHandler_New_Unauthorized(t *testing.T) {
	testUnauthorized(t, "GET", "/training-datasets/new", nil)
}

func TestTrainingDatasetHandler_Create_Unauthorized(t *testing.T) {
	trainingDataset := models.TrainingDataset{
		Name: "New Machine",
		AODFiles: []jalien.AODFile{{
			Name:      "AO2D.root",
			Path:      "/alice/sim/2024/LHC24b1b/0/567454/AOD/002/AO2D.root",
			Size:      2312421213,
			LHCPeriod: "LHC24b1b",
			RunNumber: 567454,
			AODNumber: 2,
		}},
		UserId: 1, // Arbitrary user ID
	}
	body, err := json.Marshal(trainingDataset)
	assert.NoError(t, err)

	testUnauthorized(t, "POST", "/training-datasets", body)
}

func TestTrainingDatasetHandler_Delete_Unauthorized(t *testing.T) {
	testUnauthorized(t, "DELETE", "/training-datasets/1", nil)
}

func TestTrainingDatasetHandler_ExploreDirectory_Unauthorized(t *testing.T) {
	testUnauthorized(t, "GET", "/training-datasets/explore-directory", nil)
}

func TestTrainingDatasetHandler_FindAods_Unauthorized(t *testing.T) {
	testUnauthorized(t, "GET", "/training-datasets/find-aods", nil)
}
