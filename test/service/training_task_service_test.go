package service_test

import (
	"errors"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/mytkom/AliceTraINT/internal/ccdb"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type trainingTaskServiceTestUtils struct {
	TTRepo      *repository.MockTrainingTaskRepository
	TDRepo      *repository.MockTrainingDatasetRepository
	TTRRepo     *repository.MockTrainingTaskResultRepository
	CCDBService *service.MockCCDBService
	FileService *service.MockFileService
	NNArch      *service.NNArchServiceInMemory
}

func newTrainingTaskService() (*service.TrainingTaskService, *trainingTaskServiceTestUtils) {
	ttRepo := repository.NewMockTrainingTaskRepository()
	tdRepo := repository.NewMockTrainingDatasetRepository()
	ttrRepo := repository.NewMockTrainingTaskResultRepository()
	ccdbService := service.NewMockCCDBService()
	fileService := service.NewMockFileService()
	nnArch := service.NewNNArchServiceInMemory(&service.NNFieldConfigs{
		"fieldName": service.NNConfigField{
			FullName:     "Full field name",
			Type:         "uint",
			DefaultValue: uint(512),
			Min:          uint(128),
			Max:          uint(1024),
			Step:         uint(1),
			Description:  "Field description",
		},
	}, &service.NNExpectedResults{
		Onnx: map[string]string{
			"local_file.onnx": "uploaded_file.onnx",
		},
	})

	return service.NewTrainingTaskService(&repository.RepositoryContext{
			TrainingTask:       ttRepo,
			TrainingDataset:    tdRepo,
			TrainingTaskResult: ttrRepo,
		}, ccdbService, fileService, nnArch), &trainingTaskServiceTestUtils{
			TTRepo:      ttRepo,
			TDRepo:      tdRepo,
			TTRRepo:     ttrRepo,
			CCDBService: ccdbService,
			FileService: fileService,
			NNArch:      nnArch,
		}
}

func TestTrainingTaskService_GetAll_Global(t *testing.T) {
	// Arrange
	ttService, ut := newTrainingTaskService()
	userId := uint(1)
	tdId := uint(1)
	tmId := uint(1)
	tts := []models.TrainingTask{
		{Name: "task1", UserId: userId, Status: models.Queued, TrainingMachineId: nil, TrainingDatasetId: tdId, Configuration: ""},
		{Name: "task2", UserId: userId, Status: models.Benchmarking, TrainingMachineId: &tmId, TrainingDatasetId: tdId, Configuration: ""},
	}
	ut.TTRepo.On("GetAll").Return(tts, nil)

	// Act
	tasks, err := ttService.GetAll(userId, false)

	// Assert
	assert.NoError(t, err)
	ut.TTRepo.AssertCalled(t, "GetAll")
	ut.TTRepo.AssertNotCalled(t, "GetAllUser", mock.Anything)
	assert.Equal(t, 2, len(tasks))
	assert.Equal(t, tts[0].Name, tasks[0].Name)
	assert.Equal(t, tts[1].Name, tasks[1].Name)
}

func TestTrainingTaskService_GetAll_UserScoped(t *testing.T) {
	// Arrange
	ttService, ut := newTrainingTaskService()
	userId := uint(1)
	tdId := uint(1)
	tmId := uint(1)
	tts := []models.TrainingTask{
		{Name: "task1", UserId: userId, Status: models.Queued, TrainingMachineId: nil, TrainingDatasetId: tdId, Configuration: ""},
		{Name: "task2", UserId: userId, Status: models.Benchmarking, TrainingMachineId: &tmId, TrainingDatasetId: tdId, Configuration: ""},
	}
	ut.TTRepo.On("GetAllUser", userId).Return(tts, nil)

	// Act
	tasks, err := ttService.GetAll(userId, true)

	// Assert
	assert.NoError(t, err)
	ut.TTRepo.AssertNotCalled(t, "GetAll")
	ut.TTRepo.AssertCalled(t, "GetAllUser", userId)
	assert.Equal(t, 2, len(tasks))
	assert.Equal(t, tts[0].Name, tasks[0].Name)
	assert.Equal(t, tts[1].Name, tasks[1].Name)
}

func TestTrainingTaskService_Create(t *testing.T) {
	// Arrange
	ttService, ut := newTrainingTaskService()
	userId := uint(1)
	tdId := uint(1)
	tt := models.TrainingTask{
		Name:              "task2",
		UserId:            userId,
		Status:            models.Failed, // must be changed to queued
		TrainingDatasetId: tdId,
		Configuration:     "",
	}
	ut.TTRepo.On("Create", &tt).Return(nil)

	// Act
	err := ttService.Create(&tt)

	// Assert
	assert.NoError(t, err)
	ut.TTRepo.AssertCalled(t, "Create", &tt)
	assert.Equal(t, models.Queued, tt.Status)
	assert.Equal(t, (*uint)(nil), tt.TrainingMachineId)
}

func TestTrainingTaskService_GetHelpers(t *testing.T) {
	// Arrange
	ttService, ut := newTrainingTaskService()
	userId := uint(1)
	tds := []models.TrainingDataset{
		{Name: "LHC24b1b", UserId: userId, AODFiles: []jalien.AODFile{}},
		{Name: "LHC24b1b2", UserId: userId, AODFiles: []jalien.AODFile{}},
	}
	ut.TDRepo.On("GetAllUser", userId).Return(tds, nil)

	// Act
	helpers, err := ttService.GetHelpers(userId)

	// Assert
	assert.NoError(t, err)
	ut.TDRepo.AssertCalled(t, "GetAllUser", userId)
	assert.Equal(t, tds[0].Name, helpers.TrainingDatasets[0].Name)
	assert.Equal(t, tds[1].Name, helpers.TrainingDatasets[1].Name)
	assert.True(t, reflect.DeepEqual(helpers.FieldConfigs, ut.NNArch.FieldConfigs))
}

func TestTrainingTaskService_GetByID_Queued(t *testing.T) {
	// Arrange
	ttService, ut := newTrainingTaskService()
	userId := uint(1)
	tdId := uint(1)
	ttId := uint(1)
	tt := models.TrainingTask{
		Name:              "task2",
		UserId:            userId,
		Status:            models.Queued,
		TrainingDatasetId: tdId,
		Configuration:     "",
	}
	ut.TTRepo.On("GetByID", ttId).Return(&tt, nil)
	ut.TTRepo.On("GetByType", ttId, models.Onnx).Return([]models.TrainingTaskResult{}, nil)
	ut.TTRepo.On("GetByType", ttId, models.Image).Return([]models.TrainingTaskResult{}, nil)

	// Act
	ttWithRes, err := ttService.GetByID(ttId)

	// Assert
	assert.NoError(t, err)
	ut.TTRepo.AssertCalled(t, "GetByID", ttId)
	ut.TTRRepo.AssertNotCalled(t, "GetByType", ttId, models.Onnx)
	ut.TTRRepo.AssertNotCalled(t, "GetByType", ttId, models.Image)
	assert.Equal(t, tt.Status, ttWithRes.TrainingTask.Status)
	assert.Equal(t, []models.TrainingTaskResult(nil), ttWithRes.OnnxFiles)
	assert.Equal(t, []models.TrainingTaskResult(nil), ttWithRes.ImageFiles)
}

func TestTrainingTaskService_GetByID_Training(t *testing.T) {
	// Arrange
	ttService, ut := newTrainingTaskService()
	userId := uint(1)
	tdId := uint(1)
	ttId := uint(1)
	tt := models.TrainingTask{
		Name:              "task2",
		UserId:            userId,
		Status:            models.Training,
		TrainingDatasetId: tdId,
		Configuration:     "",
	}
	ut.TTRepo.On("GetByID", ttId).Return(&tt, nil)
	ut.TTRepo.On("GetByType", ttId, models.Onnx).Return([]models.TrainingTaskResult{}, nil)
	ut.TTRepo.On("GetByType", ttId, models.Image).Return([]models.TrainingTaskResult{}, nil)

	// Act
	ttWithRes, err := ttService.GetByID(ttId)

	// Assert
	assert.NoError(t, err)
	ut.TTRepo.AssertCalled(t, "GetByID", ttId)
	ut.TTRRepo.AssertNotCalled(t, "GetByType", ttId, models.Onnx)
	ut.TTRRepo.AssertNotCalled(t, "GetByType", ttId, models.Image)
	assert.Equal(t, tt.Status, ttWithRes.TrainingTask.Status)
	assert.Equal(t, []models.TrainingTaskResult(nil), ttWithRes.OnnxFiles)
	assert.Equal(t, []models.TrainingTaskResult(nil), ttWithRes.ImageFiles)
}

func TestTrainingTaskService_GetByID_Benchmarking(t *testing.T) {
	// Arrange
	ttService, ut := newTrainingTaskService()
	userId := uint(1)
	tdId := uint(1)
	ttId := uint(1)
	tt := models.TrainingTask{
		Model:             gorm.Model{ID: ttId},
		Name:              "task2",
		UserId:            userId,
		Status:            models.Benchmarking,
		TrainingDatasetId: tdId,
		Configuration:     "",
	}
	onnxFiles := []models.TrainingTaskResult{
		{Name: "ONNX result file", Type: models.Onnx, Description: "example", FileId: 1, File: models.File{
			Name: "file.onnx", Path: "./file.onnx", Size: 12312,
		}, TrainingTaskId: ttId},
	}
	ut.TTRepo.On("GetByID", ttId).Return(&tt, nil)
	ut.TTRRepo.On("GetByType", ttId, models.Onnx).Return(onnxFiles, nil)
	ut.TTRRepo.On("GetByType", ttId, models.Image).Return([]models.TrainingTaskResult{}, nil)

	// Act
	ttWithRes, err := ttService.GetByID(ttId)

	// Assert
	assert.NoError(t, err)
	ut.TTRepo.AssertCalled(t, "GetByID", ttId)
	ut.TTRRepo.AssertCalled(t, "GetByType", ttId, models.Onnx)
	ut.TTRRepo.AssertNotCalled(t, "GetByType", ttId, models.Image)
	assert.Equal(t, tt.Status, ttWithRes.TrainingTask.Status)
	assert.True(t, reflect.DeepEqual(onnxFiles, ttWithRes.OnnxFiles))
	assert.Equal(t, []models.TrainingTaskResult(nil), ttWithRes.ImageFiles)
}

func TestTrainingTaskService_GetByID_Completed(t *testing.T) {
	// Arrange
	ttService, ut := newTrainingTaskService()
	userId := uint(1)
	tdId := uint(1)
	ttId := uint(1)
	tt := models.TrainingTask{
		Model:             gorm.Model{ID: ttId},
		Name:              "task2",
		UserId:            userId,
		Status:            models.Completed,
		TrainingDatasetId: tdId,
		Configuration:     "",
	}
	onnxFiles := []models.TrainingTaskResult{
		{Name: "ONNX result file", Type: models.Onnx, Description: "example", FileId: 1, File: models.File{
			Name: "file.onnx", Path: "./file.onnx", Size: 12312,
		}, TrainingTaskId: ttId},
	}
	imageFiles := []models.TrainingTaskResult{
		{Name: "Image result file", Type: models.Image, Description: "image", FileId: 2, File: models.File{
			Name: "file.png", Path: "./file.png", Size: 12312231,
		}, TrainingTaskId: ttId},
	}
	ut.TTRepo.On("GetByID", ttId).Return(&tt, nil)
	ut.TTRRepo.On("GetByType", ttId, models.Onnx).Return(onnxFiles, nil)
	ut.TTRRepo.On("GetByType", ttId, models.Image).Return(imageFiles, nil)

	// Act
	ttWithRes, err := ttService.GetByID(ttId)

	// Assert
	assert.NoError(t, err)
	ut.TTRepo.AssertCalled(t, "GetByID", ttId)
	ut.TTRRepo.AssertCalled(t, "GetByType", ttId, models.Onnx)
	ut.TTRRepo.AssertCalled(t, "GetByType", ttId, models.Image)
	assert.Equal(t, tt.Status, ttWithRes.TrainingTask.Status)
	assert.True(t, reflect.DeepEqual(onnxFiles, ttWithRes.OnnxFiles))
	assert.True(t, reflect.DeepEqual(imageFiles, ttWithRes.ImageFiles))
}

func TestTrainingTaskService_GetByID_Uploaded(t *testing.T) {
	// Arrange
	ttService, ut := newTrainingTaskService()
	userId := uint(1)
	tdId := uint(1)
	ttId := uint(1)
	tt := models.TrainingTask{
		Model:             gorm.Model{ID: ttId},
		Name:              "task2",
		UserId:            userId,
		Status:            models.Uploaded,
		TrainingDatasetId: tdId,
		Configuration:     "",
	}
	onnxFiles := []models.TrainingTaskResult{
		{Name: "ONNX result file", Type: models.Onnx, Description: "example", FileId: 1, File: models.File{
			Name: "file.onnx", Path: "./file.onnx", Size: 12312,
		}, TrainingTaskId: ttId},
	}
	imageFiles := []models.TrainingTaskResult{
		{Name: "Image result file", Type: models.Image, Description: "image", FileId: 2, File: models.File{
			Name: "file.png", Path: "./file.png", Size: 12312231,
		}, TrainingTaskId: ttId},
	}
	ut.TTRepo.On("GetByID", ttId).Return(&tt, nil)
	ut.TTRRepo.On("GetByType", ttId, models.Onnx).Return(onnxFiles, nil)
	ut.TTRRepo.On("GetByType", ttId, models.Image).Return(imageFiles, nil)

	// Act
	ttWithRes, err := ttService.GetByID(ttId)

	// Assert
	assert.NoError(t, err)
	ut.TTRepo.AssertCalled(t, "GetByID", ttId)
	ut.TTRRepo.AssertCalled(t, "GetByType", ttId, models.Onnx)
	ut.TTRRepo.AssertCalled(t, "GetByType", ttId, models.Image)
	assert.Equal(t, tt.Status, ttWithRes.TrainingTask.Status)
	assert.True(t, reflect.DeepEqual(onnxFiles, ttWithRes.OnnxFiles))
	assert.True(t, reflect.DeepEqual(imageFiles, ttWithRes.ImageFiles))
}

type mockReadCloser struct{}

func (m mockReadCloser) Read(p []byte) (int, error) { return 0, nil }
func (m mockReadCloser) Close() error               { return nil }

func TestTrainingTaskService_UploadToCCDB_Success(t *testing.T) {
	// Arrange
	ttService, ut := newTrainingTaskService()
	userId := uint(1)
	tdId := uint(1)
	ttId := uint(1)
	tt := models.TrainingTask{
		Model:             gorm.Model{ID: ttId},
		Name:              "task2",
		UserId:            userId,
		Status:            models.Completed,
		TrainingDatasetId: tdId,
		TrainingDataset: models.TrainingDataset{
			Model:  gorm.Model{ID: tdId},
			UserId: userId,
			AODFiles: []jalien.AODFile{
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/AOD/002/321321", RunNumber: 321321, LHCPeriod: "LHC24f3", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/AOD/002/321326", RunNumber: 321326, LHCPeriod: "LHC24f3", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/AOD/002/321338", RunNumber: 321338, LHCPeriod: "LHC24f3", AODNumber: 2},
			},
		},
		Configuration: "",
	}
	onnxFiles := []models.TrainingTaskResult{
		{Name: "local_file.onnx", Type: models.Onnx, Description: "example", FileId: 1, File: models.File{
			Name: "local_file_temp.onnx", Path: "./local_file_temp.onnx", Size: 12312,
		}, TrainingTaskId: ttId},
	}
	ut.TTRepo.On("GetByID", ttId).Return(&tt, nil)
	ut.TTRepo.On("Update", &tt).Return(nil)
	ut.TTRRepo.On("GetByType", ttId, models.Onnx).Return(onnxFiles, nil)
	file := &mockReadCloser{}
	ut.FileService.On("OpenFile", "./local_file_temp.onnx").Return(file, func(r io.ReadCloser) { r.Close() }, nil)
	now := time.Now().UnixMilli()
	ut.CCDBService.On("GetRunInformation", uint64(321321)).Return(&ccdb.RunInformation{
		RunNumber: 321321,
		SOR:       uint64(now - 10000),
		EOR:       uint64(now - 9000),
	}, nil)
	ut.CCDBService.On("GetRunInformation", uint64(321338)).Return(&ccdb.RunInformation{
		RunNumber: 321338,
		SOR:       uint64(now + 7000),
		EOR:       uint64(now + 10000),
	}, nil)
	ut.CCDBService.On("UploadFile", uint64(now-10000), uint64(now+10000), "uploaded_file.onnx", file).Return(nil)

	// Act
	err := ttService.UploadOnnxResults(ttId)

	// Assert
	assert.NoError(t, err)
	ut.TTRepo.AssertCalled(t, "GetByID", ttId)
	ut.TTRepo.AssertCalled(t, "Update", &tt)
	assert.Equal(t, models.Uploaded, tt.Status)
	ut.TTRRepo.AssertCalled(t, "GetByType", ttId, models.Onnx)
	ut.FileService.AssertCalled(t, "OpenFile", "./local_file_temp.onnx")
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(321321))
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(321338))
	ut.CCDBService.AssertCalled(t, "UploadFile", uint64(now-10000), uint64(now+10000), "uploaded_file.onnx", file)
}

func TestTrainingTaskService_UploadToCCDB_MissingExpectedFile(t *testing.T) {
	// Arrange
	ttService, ut := newTrainingTaskService()
	userId := uint(1)
	tdId := uint(1)
	ttId := uint(1)
	tt := models.TrainingTask{
		Model:             gorm.Model{ID: ttId},
		Name:              "task2",
		UserId:            userId,
		Status:            models.Completed,
		TrainingDatasetId: tdId,
		TrainingDataset: models.TrainingDataset{
			Model:  gorm.Model{ID: tdId},
			UserId: userId,
			AODFiles: []jalien.AODFile{
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/AOD/002/321321", RunNumber: 321321, LHCPeriod: "LHC24f3", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/AOD/002/321326", RunNumber: 321326, LHCPeriod: "LHC24f3", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/AOD/002/321338", RunNumber: 321338, LHCPeriod: "LHC24f3", AODNumber: 2},
			},
		},
		Configuration: "",
	}
	onnxFiles := []models.TrainingTaskResult{
		{Name: "not_expected_file.onnx", Type: models.Onnx, Description: "example", FileId: 1, File: models.File{
			Name: "local_file_temp.onnx", Path: "./local_file_temp.onnx", Size: 12312,
		}, TrainingTaskId: ttId},
	}
	ut.TTRepo.On("GetByID", ttId).Return(&tt, nil)
	ut.TTRepo.On("Update", &tt).Return(nil)
	ut.TTRRepo.On("GetByType", ttId, models.Onnx).Return(onnxFiles, nil)
	file := &mockReadCloser{}
	ut.FileService.On("OpenFile", "./local_file_temp.onnx").Return(file, func(r io.ReadCloser) { r.Close() }, nil)
	now := time.Now().UnixMilli()
	ut.CCDBService.On("GetRunInformation", uint64(321321)).Return(&ccdb.RunInformation{
		RunNumber: 321321,
		SOR:       uint64(now - 10000),
		EOR:       uint64(now - 9000),
	}, nil)
	ut.CCDBService.On("GetRunInformation", uint64(321338)).Return(&ccdb.RunInformation{
		RunNumber: 321338,
		SOR:       uint64(now + 7000),
		EOR:       uint64(now + 10000),
	}, nil)
	ut.CCDBService.On("UploadFile", uint64(now-10000), uint64(now+10000), "uploaded_file.onnx", file).Return(nil)

	// Act
	err := ttService.UploadOnnxResults(ttId)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TrainingTask's result file: local_file.onnx not found")
	ut.TTRepo.AssertCalled(t, "GetByID", ttId)
	ut.TTRepo.AssertNotCalled(t, "Update", &tt)
	ut.TTRRepo.AssertCalled(t, "GetByType", ttId, models.Onnx)
	ut.FileService.AssertNotCalled(t, "OpenFile", "./local_file_temp.onnx")
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(321321))
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(321338))
	ut.CCDBService.AssertNotCalled(t, "UploadFile", uint64(now-10000), uint64(now+10000), "uploaded_file.onnx", file)
}

func TestTrainingTaskService_UploadToCCDB_ErrorReadingFile(t *testing.T) {
	// Arrange
	ttService, ut := newTrainingTaskService()
	userId := uint(1)
	tdId := uint(1)
	ttId := uint(1)
	tt := models.TrainingTask{
		Model:             gorm.Model{ID: ttId},
		Name:              "task2",
		UserId:            userId,
		Status:            models.Completed,
		TrainingDatasetId: tdId,
		TrainingDataset: models.TrainingDataset{
			Model:  gorm.Model{ID: tdId},
			UserId: userId,
			AODFiles: []jalien.AODFile{
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/AOD/002/321321", RunNumber: 321321, LHCPeriod: "LHC24f3", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/AOD/002/321326", RunNumber: 321326, LHCPeriod: "LHC24f3", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/AOD/002/321338", RunNumber: 321338, LHCPeriod: "LHC24f3", AODNumber: 2},
			},
		},
		Configuration: "",
	}
	onnxFiles := []models.TrainingTaskResult{
		{Name: "local_file.onnx", Type: models.Onnx, Description: "example", FileId: 1, File: models.File{
			Name: "local_file_temp.onnx", Path: "./local_file_temp.onnx", Size: 12312,
		}, TrainingTaskId: ttId},
	}
	ut.TTRepo.On("GetByID", ttId).Return(&tt, nil)
	ut.TTRepo.On("Update", &tt).Return(nil)
	ut.TTRRepo.On("GetByType", ttId, models.Onnx).Return(onnxFiles, nil)
	file := &mockReadCloser{}
	ut.FileService.On("OpenFile", "./local_file_temp.onnx").Return(nil, nil, errors.New("error reading file"))
	now := time.Now().UnixMilli()
	ut.CCDBService.On("GetRunInformation", uint64(321321)).Return(&ccdb.RunInformation{
		RunNumber: 321321,
		SOR:       uint64(now - 10000),
		EOR:       uint64(now - 9000),
	}, nil)
	ut.CCDBService.On("GetRunInformation", uint64(321338)).Return(&ccdb.RunInformation{
		RunNumber: 321338,
		SOR:       uint64(now + 7000),
		EOR:       uint64(now + 10000),
	}, nil)
	ut.CCDBService.On("UploadFile", uint64(now-10000), uint64(now+10000), "uploaded_file.onnx", file).Return(nil)

	// Act
	err := ttService.UploadOnnxResults(ttId)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error reading file")
	ut.TTRepo.AssertCalled(t, "GetByID", ttId)
	ut.TTRepo.AssertNotCalled(t, "Update", &tt)
	ut.TTRRepo.AssertCalled(t, "GetByType", ttId, models.Onnx)
	ut.FileService.AssertCalled(t, "OpenFile", "./local_file_temp.onnx")
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(321321))
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(321338))
	ut.CCDBService.AssertNotCalled(t, "UploadFile", uint64(now-10000), uint64(now+10000), "uploaded_file.onnx", file)
}
