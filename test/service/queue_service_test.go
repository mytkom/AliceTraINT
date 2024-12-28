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

func newQueueService() (*service.QueueService, *service.MockHasher, *repository.MockTrainingTaskRepository, *repository.MockTrainingMachineRepository, *repository.MockTrainingTaskResultRepository, *service.MockFileService) {
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

	return service.NewQueueService(mockFileService, repoContext, mockHasher), mockHasher, mockTaskRepo, mockMachineRepo, mockTaskResultRepo, mockFileService
}

func TestQueueService_UpdateTrainingTaskStatus_Success(t *testing.T) {
	// Arrange
	queueService, _, mockTaskRepo, _, _, _ := newQueueService()
	taskID := uint(1)
	newStatus := models.Training
	mockTask := &models.TrainingTask{Model: gorm.Model{ID: taskID}, Status: models.Queued}

	mockTaskRepo.On("GetByID", taskID).Return(mockTask, nil)
	mockTaskRepo.On("Update", mock.AnythingOfType("*models.TrainingTask")).Return(nil)

	// Act
	err := queueService.UpdateTrainingTaskStatus(taskID, newStatus)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, newStatus, mockTask.Status)
	mockTaskRepo.AssertCalled(t, "GetByID", taskID)
	mockTaskRepo.AssertCalled(t, "Update", mockTask)
}

func TestQueueService_UpdateTrainingTaskStatus_TaskNotFound(t *testing.T) {
	// Arrange
	queueService, _, mockTaskRepo, _, _, _ := newQueueService()
	taskID := uint(1)
	newStatus := models.Training

	mockTaskRepo.On("GetByID", taskID).Return(nil, errors.New("task not found"))

	// Act
	err := queueService.UpdateTrainingTaskStatus(taskID, newStatus)

	// Assert
	assert.Error(t, err)
	assert.EqualError(t, err, "task not found")
	mockTaskRepo.AssertCalled(t, "GetByID", taskID)
	mockTaskRepo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestQueueService_AuthorizeTrainingMachine_Success(t *testing.T) {
	// Arrange
	queueService, mockHasher, _, mockMachineRepo, _, _ := newQueueService()
	tmID := uint(1)
	secretID := "valid_secret"
	hashedSecret := "hashed_secret"
	trainingMachine := &models.TrainingMachine{SecretKeyHashed: hashedSecret, Model: gorm.Model{ID: tmID}}

	mockMachineRepo.On("GetByID", tmID).Return(trainingMachine, nil)
	mockMachineRepo.On("Update", trainingMachine).Return(nil)
	mockHasher.On("VerifyKey", secretID, hashedSecret).Return(true, nil)

	// Act
	result, err := queueService.AuthorizeTrainingMachine(secretID, tmID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, trainingMachine, result)
	mockMachineRepo.AssertCalled(t, "GetByID", tmID)
	mockHasher.AssertCalled(t, "VerifyKey", secretID, hashedSecret)
}

func TestQueueService_AuthorizeTrainingMachine_Failure(t *testing.T) {
	// Arrange
	queueService, mockHasher, _, mockMachineRepo, _, _ := newQueueService()
	tmID := uint(1)
	secretID := "invalid_secret"
	hashedSecret := "hashed_secret"
	trainingMachine := &models.TrainingMachine{SecretKeyHashed: hashedSecret, Model: gorm.Model{ID: tmID}}

	mockMachineRepo.On("GetByID", tmID).Return(trainingMachine, nil)
	mockHasher.On("VerifyKey", secretID, hashedSecret).Return(false, nil)

	// Act
	result, err := queueService.AuthorizeTrainingMachine(secretID, tmID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "authorization failure")
	mockMachineRepo.AssertCalled(t, "GetByID", tmID)
	mockHasher.AssertCalled(t, "VerifyKey", secretID, hashedSecret)
}

func TestQueueService_AssignTaskToMachine_Success(t *testing.T) {
	// Arrange
	queueService, _, mockTaskRepo, _, _, _ := newQueueService()
	tmID := uint(1)
	mockTask := &models.TrainingTask{Model: gorm.Model{ID: 1}, Status: models.Queued}

	mockTaskRepo.On("GetFirstQueued").Return(mockTask, nil)
	mockTaskRepo.On("Update", mockTask).Return(nil)

	// Act
	task, err := queueService.AssignTaskToMachine(tmID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, task, mockTask)
	assert.Equal(t, tmID, *mockTask.TrainingMachineId)
	assert.Equal(t, models.Training, mockTask.Status)
	mockTaskRepo.AssertCalled(t, "GetFirstQueued")
	mockTaskRepo.AssertCalled(t, "Update", mockTask)
}

func TestQueueService_AssignTaskToMachine_NoTask(t *testing.T) {
	// Arrange
	queueService, _, mockTaskRepo, _, _, _ := newQueueService()

	mockTaskRepo.On("GetFirstQueued").Return(nil, errors.New("no task to run"))

	// Act
	task, err := queueService.AssignTaskToMachine(1)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.EqualError(t, err, "no task to run")
	mockTaskRepo.AssertCalled(t, "GetFirstQueued")
	mockTaskRepo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestQueueService_AssignTaskToMachine_UpdateError(t *testing.T) {
	// Arrange
	queueService, _, mockTaskRepo, _, _, _ := newQueueService()
	tmID := uint(1)
	mockTask := &models.TrainingTask{Model: gorm.Model{ID: 1}, Status: models.Queued}

	mockTaskRepo.On("GetFirstQueued").Return(mockTask, nil)
	mockTaskRepo.On("Update", mockTask).Return(errors.New("update failed"))

	// Act
	task, err := queueService.AssignTaskToMachine(tmID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.EqualError(t, err, "cannot assign task to machine: update failed")
	mockTaskRepo.AssertCalled(t, "GetFirstQueued")
	mockTaskRepo.AssertCalled(t, "Update", mockTask)
}

func TestQueueService_CreateTrainingTaskResult_Success(t *testing.T) {
	// Arrange
	queueService, _, mockTaskRepo, _, mockTaskResultRepo, mockFileService := newQueueService()
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

	mockTaskRepo.On("GetByID", taskID).Return(mockTask, nil)
	mockFileService.On("SaveFile", mock.Anything, mock.Anything).Return(mockFileModel, nil)
	mockTaskResultRepo.On("Create", mock.Anything).Return(nil)

	// Act
	result, err := queueService.CreateTrainingTaskResult(taskID, nil, nil, fileName, description, fileType)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, mockResult.Name, result.Name)
	assert.Equal(t, mockResult.Description, result.Description)
	mockTaskRepo.AssertCalled(t, "GetByID", taskID)
	mockFileService.AssertCalled(t, "SaveFile", mock.Anything, mock.Anything)
	mockTaskResultRepo.AssertCalled(t, "Create", mock.Anything)
}

func TestQueueService_CreateTrainingTaskResult_TaskNotFound(t *testing.T) {
	// Arrange
	queueService, _, mockTaskRepo, _, mockTaskResultRepo, mockFileService := newQueueService()
	taskID := uint(1)

	mockTaskRepo.On("GetByID", taskID).Return(nil, errors.New("training task does not exist"))

	// Act
	result, err := queueService.CreateTrainingTaskResult(taskID, nil, nil, "test", "desc", "1")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "training task does not exist")
	mockTaskRepo.AssertCalled(t, "GetByID", taskID)
	mockFileService.AssertNotCalled(t, "SaveFile", mock.Anything, mock.Anything)
	mockTaskResultRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestQueueService_CreateTrainingTaskResult_FileSaveError(t *testing.T) {
	// Arrange
	queueService, _, mockTaskRepo, _, mockTaskResultRepo, mockFileService := newQueueService()
	taskID := uint(1)
	mockTask := &models.TrainingTask{Model: gorm.Model{ID: taskID}}

	mockTaskRepo.On("GetByID", taskID).Return(mockTask, nil)
	mockFileService.On("SaveFile", mock.Anything, mock.Anything).Return(nil, errors.New("file save error"))

	// Act
	result, err := queueService.CreateTrainingTaskResult(taskID, nil, nil, "test", "desc", "1")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "error saving file: file save error")
	mockTaskRepo.AssertCalled(t, "GetByID", taskID)
	mockFileService.AssertCalled(t, "SaveFile", mock.Anything, mock.Anything)
	mockTaskResultRepo.AssertNotCalled(t, "Create", mock.Anything)
}
