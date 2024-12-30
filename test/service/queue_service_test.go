package service_test

import (
	"errors"
	"testing"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type queueServiceTestUtils struct {
	TTRepo      *repository.MockTrainingTaskRepository
	TTRRepo     *repository.MockTrainingTaskResultRepository
	TMRepo      *repository.MockTrainingMachineRepository
	FileService *service.MockFileService
	Hasher      *service.MockHasher
}

func newQueueService() (*service.QueueService, *queueServiceTestUtils) {
	mockHasher := service.NewMockHasher()
	mockTaskRepo := repository.NewMockTrainingTaskRepository()
	mockMachineRepo := repository.NewMockTrainingMachineRepository()
	mockTaskResultRepo := repository.NewMockTrainingTaskResultRepository()
	mockFileService := service.NewMockFileService()

	repoContext := &repository.RepositoryContext{
		TrainingTask:       mockTaskRepo,
		TrainingMachine:    mockMachineRepo,
		TrainingTaskResult: mockTaskResultRepo,
	}

	return service.NewQueueService(mockFileService, repoContext, mockHasher), &queueServiceTestUtils{
		TTRepo:      mockTaskRepo,
		TTRRepo:     mockTaskResultRepo,
		TMRepo:      mockMachineRepo,
		FileService: mockFileService,
		Hasher:      mockHasher,
	}
}

func TestQueueService_UpdateTrainingTaskStatus_Success(t *testing.T) {
	// Arrange
	queueService, ut := newQueueService()
	taskID := uint(1)
	newStatus := models.Training
	mockTask := &models.TrainingTask{Model: gorm.Model{ID: taskID}, Status: models.Queued}

	ut.TTRepo.On("GetByID", taskID).Return(mockTask, nil)
	ut.TTRepo.On("Update", mock.AnythingOfType("*models.TrainingTask")).Return(nil)

	// Act
	err := queueService.UpdateTrainingTaskStatus(taskID, newStatus)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, newStatus, mockTask.Status)
	ut.TTRepo.AssertCalled(t, "GetByID", taskID)
	ut.TTRepo.AssertCalled(t, "Update", mockTask)
}

func TestQueueService_UpdateTrainingTaskStatus_TaskNotFound(t *testing.T) {
	// Arrange
	queueService, ut := newQueueService()
	taskID := uint(1)
	newStatus := models.Training

	ut.TTRepo.On("GetByID", taskID).Return(nil, errors.New("task not found"))

	// Act
	err := queueService.UpdateTrainingTaskStatus(taskID, newStatus)

	// Assert
	assert.Error(t, err)
	assert.EqualError(t, err, "task not found")
	ut.TTRepo.AssertCalled(t, "GetByID", taskID)
	ut.TTRepo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestQueueService_AuthorizeTrainingMachine_Success(t *testing.T) {
	// Arrange
	queueService, ut := newQueueService()
	tmID := uint(1)
	secretID := "valid_secret"
	hashedSecret := "hashed_secret"
	trainingMachine := &models.TrainingMachine{SecretKeyHashed: hashedSecret, Model: gorm.Model{ID: tmID}}

	ut.TMRepo.On("GetByID", tmID).Return(trainingMachine, nil)
	ut.TMRepo.On("Update", trainingMachine).Return(nil)
	ut.Hasher.On("VerifyKey", secretID, hashedSecret).Return(true, nil)

	// Act
	result, err := queueService.AuthorizeTrainingMachine(secretID, tmID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, trainingMachine, result)
	ut.TMRepo.AssertCalled(t, "GetByID", tmID)
	ut.Hasher.AssertCalled(t, "VerifyKey", secretID, hashedSecret)
}

func TestQueueService_AuthorizeTrainingMachine_Failure(t *testing.T) {
	// Arrange
	queueService, ut := newQueueService()
	tmID := uint(1)
	secretID := "invalid_secret"
	hashedSecret := "hashed_secret"
	trainingMachine := &models.TrainingMachine{SecretKeyHashed: hashedSecret, Model: gorm.Model{ID: tmID}}

	ut.TMRepo.On("GetByID", tmID).Return(trainingMachine, nil)
	ut.Hasher.On("VerifyKey", secretID, hashedSecret).Return(false, nil)

	// Act
	result, err := queueService.AuthorizeTrainingMachine(secretID, tmID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "authorization failure")
	ut.TMRepo.AssertCalled(t, "GetByID", tmID)
	ut.Hasher.AssertCalled(t, "VerifyKey", secretID, hashedSecret)
}

func TestQueueService_AssignTaskToMachine_Success(t *testing.T) {
	// Arrange
	queueService, ut := newQueueService()
	tmID := uint(1)
	mockTask := &models.TrainingTask{Model: gorm.Model{ID: 1}, Status: models.Queued}

	ut.TTRepo.On("GetFirstQueued").Return(mockTask, nil)
	ut.TTRepo.On("Update", mockTask).Return(nil)

	// Act
	task, err := queueService.AssignTaskToMachine(tmID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, task, mockTask)
	assert.Equal(t, tmID, *mockTask.TrainingMachineId)
	assert.Equal(t, models.Training, mockTask.Status)
	ut.TTRepo.AssertCalled(t, "GetFirstQueued")
	ut.TTRepo.AssertCalled(t, "Update", mockTask)
}

func TestQueueService_AssignTaskToMachine_NoTask(t *testing.T) {
	// Arrange
	queueService, ut := newQueueService()

	ut.TTRepo.On("GetFirstQueued").Return(nil, errors.New("no task to run"))

	// Act
	task, err := queueService.AssignTaskToMachine(1)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.EqualError(t, err, "no task to run")
	ut.TTRepo.AssertCalled(t, "GetFirstQueued")
	ut.TTRepo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestQueueService_AssignTaskToMachine_UpdateError(t *testing.T) {
	// Arrange
	queueService, ut := newQueueService()
	tmID := uint(1)
	mockTask := &models.TrainingTask{Model: gorm.Model{ID: 1}, Status: models.Queued}

	ut.TTRepo.On("GetFirstQueued").Return(mockTask, nil)
	ut.TTRepo.On("Update", mockTask).Return(errors.New("update failed"))

	// Act
	task, err := queueService.AssignTaskToMachine(tmID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.EqualError(t, err, "cannot assign task to machine: update failed")
	ut.TTRepo.AssertCalled(t, "GetFirstQueued")
	ut.TTRepo.AssertCalled(t, "Update", mockTask)
}

func TestQueueService_CreateTrainingTaskResult_Success(t *testing.T) {
	// Arrange
	queueService, ut := newQueueService()
	taskID := uint(1)
	fileName := "test-file.txt"
	description := "Test description"
	fileType := "1"
	mockFileModel := &models.File{Name: fileName, Model: gorm.Model{ID: 1}}
	mockTask := &models.TrainingTask{Model: gorm.Model{ID: taskID}}
	mockResult := &models.TrainingTaskResult{
		Name:           fileName,
		Description:    description,
		Type:           models.TrainingTaskResultType(1),
		TrainingTaskId: taskID,
		File:           *mockFileModel,
	}

	ut.TTRepo.On("GetByID", taskID).Return(mockTask, nil)
	ut.FileService.On("SaveFile", mock.Anything, mock.Anything).Return(mockFileModel, nil)
	ut.TTRRepo.On("Create", mock.Anything).Return(nil)

	// Act
	result, err := queueService.CreateTrainingTaskResult(taskID, nil, nil, fileName, description, fileType)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, mockResult.Name, result.Name)
	assert.Equal(t, mockResult.Description, result.Description)
	ut.TTRepo.AssertCalled(t, "GetByID", taskID)
	ut.FileService.AssertCalled(t, "SaveFile", mock.Anything, mock.Anything)
	ut.TTRRepo.AssertCalled(t, "Create", mock.Anything)
}

func TestQueueService_CreateTrainingTaskResult_TaskNotFound(t *testing.T) {
	// Arrange
	queueService, ut := newQueueService()
	taskID := uint(1)

	ut.TTRepo.On("GetByID", taskID).Return(nil, errors.New("training task does not exist"))

	// Act
	result, err := queueService.CreateTrainingTaskResult(taskID, nil, nil, "test", "desc", "1")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "training task does not exist")
	ut.TTRepo.AssertCalled(t, "GetByID", taskID)
	ut.FileService.AssertNotCalled(t, "SaveFile", mock.Anything, mock.Anything)
	ut.TTRRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestQueueService_CreateTrainingTaskResult_FileSaveError(t *testing.T) {
	// Arrange
	queueService, ut := newQueueService()
	taskID := uint(1)
	mockTask := &models.TrainingTask{Model: gorm.Model{ID: taskID}}

	ut.TTRepo.On("GetByID", taskID).Return(mockTask, nil)
	ut.FileService.On("SaveFile", mock.Anything, mock.Anything).Return(nil, errors.New("file save error"))

	// Act
	result, err := queueService.CreateTrainingTaskResult(taskID, nil, nil, "test", "desc", "1")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "error saving file: file save error")
	ut.TTRepo.AssertCalled(t, "GetByID", taskID)
	ut.FileService.AssertCalled(t, "SaveFile", mock.Anything, mock.Anything)
	ut.TTRRepo.AssertNotCalled(t, "Create", mock.Anything)
}
