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
	TTRepo        *repository.MockTrainingTaskRepository
	TDRepo        *repository.MockTrainingDatasetRepository
	TTRRepo       *repository.MockTrainingTaskResultRepository
	CCDBService   *service.MockCCDBService
	JAliEnService *service.MockJAliEnService
	FileService   *service.MockFileService
	NNArch        *service.NNArchServiceInMemory
}

func newTrainingTaskService() (*service.TrainingTaskService, *trainingTaskServiceTestUtils) {
	ttRepo := repository.NewMockTrainingTaskRepository()
	tdRepo := repository.NewMockTrainingDatasetRepository()
	ttrRepo := repository.NewMockTrainingTaskResultRepository()
	jalienService := service.NewMockJAliEnService()
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
		}, ccdbService, jalienService, fileService, nnArch), &trainingTaskServiceTestUtils{
			TTRepo:        ttRepo,
			TDRepo:        tdRepo,
			TTRRepo:       ttrRepo,
			CCDBService:   ccdbService,
			JAliEnService: jalienService,
			FileService:   fileService,
			NNArch:        nnArch,
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
		Model:             gorm.Model{ID: ttId},
		Name:              "task2",
		UserId:            userId,
		Status:            models.Training,
		TrainingDatasetId: tdId,
		Configuration:     "",
	}
	ut.TTRepo.On("GetByID", ttId).Return(&tt, nil)
	ut.TTRRepo.On("GetByType", ttId, models.Onnx).Return([]models.TrainingTaskResult{}, nil)
	ut.TTRRepo.On("GetByType", ttId, models.Image).Return([]models.TrainingTaskResult{}, nil)

	// Act
	ttWithRes, err := ttService.GetByID(ttId)

	// Assert
	assert.NoError(t, err)
	ut.TTRepo.AssertCalled(t, "GetByID", ttId)
	ut.TTRRepo.AssertNotCalled(t, "GetByType", ttId, models.Onnx)
	ut.TTRRepo.AssertCalled(t, "GetByType", ttId, models.Image)
	assert.Equal(t, tt.Status, ttWithRes.TrainingTask.Status)
	assert.Equal(t, []models.TrainingTaskResult(nil), ttWithRes.OnnxFiles)
	assert.Equal(t, []models.TrainingTaskResult{}, ttWithRes.ImageFiles)
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
	ut.TTRRepo.AssertCalled(t, "GetByType", ttId, models.Image)
	assert.Equal(t, tt.Status, ttWithRes.TrainingTask.Status)
	assert.True(t, reflect.DeepEqual(onnxFiles, ttWithRes.OnnxFiles))
	assert.Equal(t, []models.TrainingTaskResult{}, ttWithRes.ImageFiles)
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

func TestTrainingTaskService_UploadToCCDB_OnePeriod(t *testing.T) {
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
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/321321/AOD/002", RunNumber: 321321, LHCPeriod: "LHC24f3", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/321326/AOD/002", RunNumber: 321326, LHCPeriod: "LHC24f3", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/321338/AOD/002", RunNumber: 321338, LHCPeriod: "LHC24f3", AODNumber: 2},
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
	ut.JAliEnService.On("ListAndParseDirectory", "/alice/sim/2024/LHC24f3/0").Return(&jalien.DirectoryContents{
		Subdirs: []jalien.Dir{
			{Name: "321000", Path: "/alice/sim/2024/LHC24f3/0/321000"},
			{Name: "321100", Path: "/alice/sim/2024/LHC24f3/0/321100"},
			{Name: "321321", Path: "/alice/sim/2024/LHC24f3/0/321321"},
			{Name: "321326", Path: "/alice/sim/2024/LHC24f3/0/321326"},
			{Name: "321338", Path: "/alice/sim/2024/LHC24f3/0/321338"},
			{Name: "321400", Path: "/alice/sim/2024/LHC24f3/0/321400"},
			{Name: "321500", Path: "/alice/sim/2024/LHC24f3/0/321500"},
		},
	}, nil)
	now := time.Now().UnixMilli()
	ut.CCDBService.On("GetRunInformation", uint64(321000)).Return(&ccdb.RunInformation{
		RunNumber: 321000,
		SOR:       uint64(now - 10000),
		EOR:       uint64(now - 9000),
	}, nil)
	ut.CCDBService.On("GetRunInformation", uint64(321500)).Return(&ccdb.RunInformation{
		RunNumber: 321500,
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
	ut.JAliEnService.AssertCalled(t, "ListAndParseDirectory", "/alice/sim/2024/LHC24f3/0")
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(321000))
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(321500))
	ut.CCDBService.AssertCalled(t, "UploadFile", uint64(now-10000), uint64(now+10000), "uploaded_file.onnx", file)
}

// 3 different periods in one dataset
// expected: take min SOR and max EOR from all periods' runs
func TestTrainingTaskService_UploadToCCDB_ManyPeriods(t *testing.T) {
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
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24b1b/0/320000/AOD/002", RunNumber: 320000, LHCPeriod: "LHC24b1b", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24b1b/0/320100/AOD/002", RunNumber: 320100, LHCPeriod: "LHC24b1b", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24b1b/0/320200/AOD/002", RunNumber: 320200, LHCPeriod: "LHC24b1b", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/321321/AOD/002", RunNumber: 321321, LHCPeriod: "LHC24f3", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/321326/AOD/002", RunNumber: 321326, LHCPeriod: "LHC24f3", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/321338/AOD/002", RunNumber: 321338, LHCPeriod: "LHC24f3", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24c1/322000/AOD/002", RunNumber: 322000, LHCPeriod: "LHC24c1", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24c1/322100/AOD/002", RunNumber: 322100, LHCPeriod: "LHC24c1", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24c1/322200/AOD/002", RunNumber: 322200, LHCPeriod: "LHC24c1", AODNumber: 2},
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

	// Mock JAliEn and CCDB with timestamp information
	now := time.Now().UnixMilli()
	// LHC24b1b
	ut.JAliEnService.On("ListAndParseDirectory", "/alice/sim/2024/LHC24b1b/0").Return(&jalien.DirectoryContents{
		Subdirs: []jalien.Dir{
			{Name: "319900", Path: "/alice/sim/2024/LHC24b1b/0/319900"}, // min run number LHC24b1b
			{Name: "320000", Path: "/alice/sim/2024/LHC24b1b/0/320000"},
			{Name: "320100", Path: "/alice/sim/2024/LHC24b1b/0/320100"},
			{Name: "320200", Path: "/alice/sim/2024/LHC24b1b/0/320200"},
			{Name: "320300", Path: "/alice/sim/2024/LHC24b1b/0/320300"}, // max run number LHC24b1b
		},
	}, nil)
	ut.CCDBService.On("GetRunInformation", uint64(319900)).Return(&ccdb.RunInformation{
		RunNumber: 319900,
		SOR:       uint64(now - 20000),
		EOR:       uint64(now - 19000),
	}, nil)
	ut.CCDBService.On("GetRunInformation", uint64(320300)).Return(&ccdb.RunInformation{
		RunNumber: 320300,
		SOR:       uint64(now - 12000),
		EOR:       uint64(now - 11000),
	}, nil)
	// LHC24f3
	ut.JAliEnService.On("ListAndParseDirectory", "/alice/sim/2024/LHC24f3/0").Return(&jalien.DirectoryContents{
		Subdirs: []jalien.Dir{
			{Name: "321000", Path: "/alice/sim/2024/LHC24f3/0/321000"}, // min run number LHC24f3
			{Name: "321321", Path: "/alice/sim/2024/LHC24f3/0/321321"},
			{Name: "321326", Path: "/alice/sim/2024/LHC24f3/0/321326"},
			{Name: "321338", Path: "/alice/sim/2024/LHC24f3/0/321338"},
			{Name: "321500", Path: "/alice/sim/2024/LHC24f3/0/321500"}, // max run number LHC24f3
		},
	}, nil)
	ut.CCDBService.On("GetRunInformation", uint64(321000)).Return(&ccdb.RunInformation{
		RunNumber: 321000,
		SOR:       uint64(now - 10000),
		EOR:       uint64(now - 9000),
	}, nil)
	ut.CCDBService.On("GetRunInformation", uint64(321500)).Return(&ccdb.RunInformation{
		RunNumber: 321500,
		SOR:       uint64(now - 2000),
		EOR:       uint64(now - 1000),
	}, nil)
	// LHC24c1
	// 	simulate 2023 datasets (without numbered subdir in period dir)
	ut.JAliEnService.On("ListAndParseDirectory", "/alice/sim/2024/LHC24c1").Return(&jalien.DirectoryContents{
		Subdirs: []jalien.Dir{
			{Name: "321900", Path: "/alice/sim/2024/LHC24c1/321900"}, // min run number LHC24c1
			{Name: "322000", Path: "/alice/sim/2024/LHC24c1/322000"},
			{Name: "322100", Path: "/alice/sim/2024/LHC24c1/322100"},
			{Name: "322200", Path: "/alice/sim/2024/LHC24c1/322200"},
			{Name: "322300", Path: "/alice/sim/2024/LHC24c1/322300"}, // max run number LHC24c1
		},
	}, nil)
	ut.CCDBService.On("GetRunInformation", uint64(321900)).Return(&ccdb.RunInformation{
		RunNumber: 321900,
		SOR:       uint64(now),
		EOR:       uint64(now + 1000),
	}, nil)
	ut.CCDBService.On("GetRunInformation", uint64(322300)).Return(&ccdb.RunInformation{
		RunNumber: 322300,
		SOR:       uint64(now + 9000),
		EOR:       uint64(now + 10000),
	}, nil)

	// TODO b1b and c1
	ut.CCDBService.On("UploadFile", uint64(now-20000), uint64(now+10000), "uploaded_file.onnx", file).Return(nil)

	// Act
	err := ttService.UploadOnnxResults(ttId)

	// Assert
	assert.NoError(t, err)
	ut.TTRepo.AssertCalled(t, "GetByID", ttId)
	ut.TTRepo.AssertCalled(t, "Update", &tt)
	assert.Equal(t, models.Uploaded, tt.Status)
	ut.TTRRepo.AssertCalled(t, "GetByType", ttId, models.Onnx)
	ut.FileService.AssertCalled(t, "OpenFile", "./local_file_temp.onnx")
	// LHC24b1b info
	ut.JAliEnService.AssertCalled(t, "ListAndParseDirectory", "/alice/sim/2024/LHC24b1b/0")
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(319900))
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(320300))
	// LHC24f3 info
	ut.JAliEnService.AssertCalled(t, "ListAndParseDirectory", "/alice/sim/2024/LHC24f3/0")
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(321000))
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(321500))
	// LHC24c1 info
	ut.JAliEnService.AssertCalled(t, "ListAndParseDirectory", "/alice/sim/2024/LHC24c1")
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(321900))
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(322300))
	// Upload
	ut.CCDBService.AssertCalled(t, "UploadFile", uint64(now-20000), uint64(now+10000), "uploaded_file.onnx", file)
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
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/321321/AOD/002", RunNumber: 321321, LHCPeriod: "LHC24f3", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/321326/AOD/002", RunNumber: 321326, LHCPeriod: "LHC24f3", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/321338/AOD/002", RunNumber: 321338, LHCPeriod: "LHC24f3", AODNumber: 2},
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
	ut.JAliEnService.On("ListAndParseDirectory", "/alice/sim/2024/LHC24f3/0").Return(&jalien.DirectoryContents{
		Subdirs: []jalien.Dir{
			{Name: "321000", Path: "/alice/sim/2024/LHC24f3/0/321000"},
			{Name: "321100", Path: "/alice/sim/2024/LHC24f3/0/321100"},
			{Name: "321321", Path: "/alice/sim/2024/LHC24f3/0/321321"},
			{Name: "321326", Path: "/alice/sim/2024/LHC24f3/0/321326"},
			{Name: "321338", Path: "/alice/sim/2024/LHC24f3/0/321338"},
			{Name: "321400", Path: "/alice/sim/2024/LHC24f3/0/321400"},
			{Name: "321500", Path: "/alice/sim/2024/LHC24f3/0/321500"},
		},
	}, nil)
	now := time.Now().UnixMilli()
	ut.CCDBService.On("GetRunInformation", uint64(321000)).Return(&ccdb.RunInformation{
		RunNumber: 321000,
		SOR:       uint64(now - 10000),
		EOR:       uint64(now - 9000),
	}, nil)
	ut.CCDBService.On("GetRunInformation", uint64(321500)).Return(&ccdb.RunInformation{
		RunNumber: 321500,
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
	ut.JAliEnService.AssertCalled(t, "ListAndParseDirectory", "/alice/sim/2024/LHC24f3/0")
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(321000))
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(321500))
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
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/321321/AOD/002", RunNumber: 321321, LHCPeriod: "LHC24f3", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/321326/AOD/002", RunNumber: 321326, LHCPeriod: "LHC24f3", AODNumber: 2},
				{Name: "AO2D.root", Path: "/alice/sim/2024/LHC24f3/0/321338/AOD/002", RunNumber: 321338, LHCPeriod: "LHC24f3", AODNumber: 2},
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
	ut.JAliEnService.On("ListAndParseDirectory", "/alice/sim/2024/LHC24f3/0").Return(&jalien.DirectoryContents{
		Subdirs: []jalien.Dir{
			{Name: "321000", Path: "/alice/sim/2024/LHC24f3/0/321000"},
			{Name: "321100", Path: "/alice/sim/2024/LHC24f3/0/321100"},
			{Name: "321321", Path: "/alice/sim/2024/LHC24f3/0/321321"},
			{Name: "321326", Path: "/alice/sim/2024/LHC24f3/0/321326"},
			{Name: "321338", Path: "/alice/sim/2024/LHC24f3/0/321338"},
			{Name: "321400", Path: "/alice/sim/2024/LHC24f3/0/321400"},
			{Name: "321500", Path: "/alice/sim/2024/LHC24f3/0/321500"},
		},
	}, nil)
	now := time.Now().UnixMilli()
	ut.CCDBService.On("GetRunInformation", uint64(321000)).Return(&ccdb.RunInformation{
		RunNumber: 321000,
		SOR:       uint64(now - 10000),
		EOR:       uint64(now - 9000),
	}, nil)
	ut.CCDBService.On("GetRunInformation", uint64(321500)).Return(&ccdb.RunInformation{
		RunNumber: 321500,
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
	ut.JAliEnService.AssertCalled(t, "ListAndParseDirectory", "/alice/sim/2024/LHC24f3/0")
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(321000))
	ut.CCDBService.AssertCalled(t, "GetRunInformation", uint64(321500))
	ut.CCDBService.AssertNotCalled(t, "UploadFile", uint64(now-10000), uint64(now+10000), "uploaded_file.onnx", file)
}
