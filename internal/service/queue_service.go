package service

import (
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"time"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
)

type IQueueService interface {
	AuthorizeTrainingMachine(secretID string, tmID uint) (*models.TrainingMachine, error)
	UpdateTrainingTaskStatus(taskID uint, status models.TrainingTaskStatus) error
	AssignTaskToMachine(tmID uint) (*models.TrainingTask, error)
	CreateTrainingTaskResult(ttID uint, file multipart.File, handler *multipart.FileHeader, name, description, fileType string) (*models.TrainingTaskResult, error)
}

type QueueService struct {
	*repository.RepositoryContext
	FileService IFileService
	Hasher      Hasher
}

func NewQueueService(fileService IFileService, repo *repository.RepositoryContext, hasher Hasher) *QueueService {
	return &QueueService{
		RepositoryContext: repo,
		FileService:       fileService,
		Hasher:            hasher,
	}
}

func (qs *QueueService) AuthorizeTrainingMachine(secretID string, tmID uint) (*models.TrainingMachine, error) {
	trainingMachine, err := qs.TrainingMachine.GetByID(tmID)
	if err != nil {
		return nil, err
	}

	ok, err := qs.Hasher.VerifyKey(secretID, trainingMachine.SecretKeyHashed)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errors.New("authorization failure")
	}

	trainingMachine.LastActivityAt = time.Now()
	err = qs.TrainingMachine.Update(trainingMachine)
	if err != nil {
		return nil, errors.New("machine activity timestamp error")
	}

	return trainingMachine, nil
}

func (qs *QueueService) UpdateTrainingTaskStatus(taskID uint, status models.TrainingTaskStatus) error {
	tt, err := qs.TrainingTask.GetByID(taskID)
	if err != nil {
		return err
	}

	tt.Status = status
	return qs.TrainingTask.Update(tt)
}

func (qs *QueueService) AssignTaskToMachine(tmID uint) (*models.TrainingTask, error) {
	tt, err := qs.TrainingTask.GetFirstQueued()
	if err != nil {
		return nil, errors.New("no task to run")
	}

	tt.TrainingMachineId = &tmID
	tt.Status = models.Training

	err = qs.TrainingTask.Update(tt)
	if err != nil {
		return nil, fmt.Errorf("cannot assign task to machine: %w", err)
	}

	return tt, nil
}

func (qs *QueueService) CreateTrainingTaskResult(ttID uint, file multipart.File, handler *multipart.FileHeader, name, description, fileType string) (*models.TrainingTaskResult, error) {
	tt, err := qs.TrainingTask.GetByID(ttID)
	if err != nil {
		return nil, errors.New("training task does not exist")
	}

	fileModel, err := qs.FileService.SaveFile(file, handler)
	if err != nil {
		return nil, fmt.Errorf("error saving file: %w", err)
	}

	fileTypeUint, _ := strconv.ParseUint(fileType, 9, 64)

	ttr := &models.TrainingTaskResult{
		File:           *fileModel,
		Name:           name,
		Description:    description,
		Type:           models.TrainingTaskResultType(fileTypeUint),
		TrainingTaskId: tt.ID,
	}

	err = qs.TrainingTaskResult.Create(ttr)
	if err != nil {
		return nil, fmt.Errorf("error during task result creation: %w", err)
	}

	return ttr, nil
}
