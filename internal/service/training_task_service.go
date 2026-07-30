package service

import (
	"errors"
	"fmt"

	"github.com/mytkom/AliceTraINT/internal/applog"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"gorm.io/gorm"
)

type TrainingTaskWithResults struct {
	TrainingTask *models.TrainingTask
	ImageFiles   []models.TrainingTaskResult
	OnnxFiles    []models.TrainingTaskResult
	LogFiles     []models.TrainingTaskResult
}

type TrainingTaskHelpers struct {
	TrainingDatasets []models.TrainingDataset
	FieldGroups      []NNFieldGroup
}

type ITrainingTaskService interface {
	Create(tm *models.TrainingTask) error
	GetAll(loggedUserId uint, userScoped bool) ([]models.TrainingTask, error)
	GetHelpers(loggedUserId uint) (*TrainingTaskHelpers, error)
	GetByID(id uint) (*TrainingTaskWithResults, error)
	UploadOnnxResults(id uint) error
}

type TrainingTaskService struct {
	*repository.RepositoryContext
	CCDBService     ICCDBService
	MonalisaService IMonalisaService
	FileService     IFileService
	NNArch          INNArchService
}

func NewTrainingTaskService(repo *repository.RepositoryContext, ccdbService ICCDBService, monalisaService IMonalisaService, fileService IFileService, nnArch INNArchService) *TrainingTaskService {
	return &TrainingTaskService{
		RepositoryContext: repo,
		CCDBService:       ccdbService,
		MonalisaService:   monalisaService,
		FileService:       fileService,
		NNArch:            nnArch,
	}
}

var errTaskNotFound = NewErrHandlerNotFound("TrainingTask")

func (s *TrainingTaskService) Create(tt *models.TrainingTask) error {
	// Status must start with Queued
	tt.Status = models.Queued

	err := s.TrainingTask.Create(tt)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return &ErrHandlerValidation{
				Field: "Name",
				Msg:   errMsgNotUnique,
			}
		} else {
			return InternalError("create training task", err)
		}
	}

	return nil
}

func (s *TrainingTaskService) GetAll(loggedUserId uint, userScoped bool) ([]models.TrainingTask, error) {
	var trainingTasks []models.TrainingTask
	var err error

	if userScoped {
		trainingTasks, err = s.TrainingTask.GetAllUser(loggedUserId)
		if err != nil {
			return nil, InternalError("get training tasks", err)
		}
	} else {
		trainingTasks, err = s.TrainingTask.GetAll()
		if err != nil {
			return nil, InternalError("get training tasks", err)
		}
	}

	return trainingTasks, nil
}

func (s *TrainingTaskService) GetHelpers(loggedUserId uint) (*TrainingTaskHelpers, error) {
	trainingDatasets, err := s.TrainingDataset.GetAllUser(loggedUserId)
	if err != nil {
		return nil, InternalError("get training task helpers", err)
	}

	return &TrainingTaskHelpers{
		TrainingDatasets: trainingDatasets,
		FieldGroups:      s.NNArch.GetFieldGroups(),
	}, nil
}

func (s *TrainingTaskService) GetByID(id uint) (*TrainingTaskWithResults, error) {
	trainingTask, err := s.TrainingTask.GetByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errTaskNotFound
		} else {
			return nil, InternalError("get training task by id", err)
		}
	}

	var imageFiles []models.TrainingTaskResult
	if trainingTask.Status >= models.Training {
		imageFiles, err = s.TrainingTaskResult.GetByType(trainingTask.ID, models.Image)
		if err != nil {
			return nil, InternalError("get training task image results", err)
		}
	}

	var onnxFiles []models.TrainingTaskResult
	if trainingTask.Status >= models.Benchmarking {
		onnxFiles, err = s.TrainingTaskResult.GetByType(trainingTask.ID, models.Onnx)
		if err != nil {
			return nil, InternalError("get training task onnx results", err)
		}
	}

	var logFiles []models.TrainingTaskResult
	if trainingTask.Status != models.Queued {
		logFiles, err = s.TrainingTaskResult.GetByType(trainingTask.ID, models.Log)
		if err != nil {
			return nil, InternalError("get training task log results", err)
		}
	}

	return &TrainingTaskWithResults{
		TrainingTask: trainingTask,
		ImageFiles:   imageFiles,
		OnnxFiles:    onnxFiles,
		LogFiles:     logFiles,
	}, nil
}

func (s *TrainingTaskService) UploadOnnxResults(id uint) error {
	trainingTask, err := s.TrainingTask.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errTaskNotFound
		}
		return InternalError("get training task for upload", err)
	}

	if trainingTask.Status < models.Completed {
		return &ErrHandlerValidation{
			Field: "Status",
			Msg:   "must be completed or uploaded",
		}
	}

	minSOR, maxEOR, err := s.getUploadTimestampRange(trainingTask)
	if err != nil {
		return err
	}

	mappedOnnxFiles, err := s.filterOnnxFiles(trainingTask.ID)
	if err != nil {
		return err
	}

	for uploadName, file := range mappedOnnxFiles {
		if err := s.uploadOnnxFile(minSOR, maxEOR, file, uploadName); err != nil {
			return err
		}
	}

	trainingTask.Status = models.Uploaded
	if err := s.TrainingTask.Update(trainingTask); err != nil {
		return InternalError("update training task status after upload", err)
	}

	return nil
}

func (s *TrainingTaskService) getUploadTimestampRange(trainingTask *models.TrainingTask) (uint64, uint64, error) {
	if !trainingTask.TrainingDataset.AnchorProdTimestampFrom.IsZero() &&
		!trainingTask.TrainingDataset.AnchorProdTimestampTo.IsZero() {
		return uint64(trainingTask.TrainingDataset.AnchorProdTimestampFrom.UnixMilli()),
			uint64(trainingTask.TrainingDataset.AnchorProdTimestampTo.UnixMilli()),
			nil
	}

	timestampFrom, timestampTo, err := resolveDatasetTimestampRange(trainingTask.TrainingDataset.AODFiles, s.MonalisaService, s.CCDBService)
	if err == nil {
		return uint64(timestampFrom.UnixMilli()), uint64(timestampTo.UnixMilli()), nil
	}

	applog.Fallback("resolve dataset timestamp range", err, "dataset AOD run range")

	minRun, maxRun, err := getDatasetAODRunRange(trainingTask.TrainingDataset.AODFiles)
	if err != nil {
		return 0, 0, err
	}

	firstRunInfo, lastRunInfo, err := getRunInfoRangeFromCCDB(s.CCDBService, minRun, maxRun)
	if err != nil {
		return 0, 0, err
	}

	return firstRunInfo.SOR, lastRunInfo.EOR, nil
}

func (s *TrainingTaskService) filterOnnxFiles(ttId uint) (map[string]*models.TrainingTaskResult, error) {
	expectedOnnxFilenames := s.NNArch.GetExpectedResults().Onnx
	mappedResults := make(map[string]*models.TrainingTaskResult, len(expectedOnnxFilenames))
	onnxFiles, err := s.TrainingTaskResult.GetByType(ttId, models.Onnx)
	if err != nil {
		return nil, err
	}

	for localName, expectedName := range expectedOnnxFilenames {
		found := false
		for _, file := range onnxFiles {
			if file.Name == localName {
				mappedResults[expectedName] = &file
				found = true
				break
			}
		}

		if !found {
			return nil, NewErrHandlerNotFound(fmt.Sprintf("TrainingTask's result file: %s", localName))
		}
	}

	return mappedResults, nil
}

func (s *TrainingTaskService) uploadOnnxFile(sor, eor uint64, onnxFile *models.TrainingTaskResult, uploadFilename string) error {
	f, closeFile, err := s.FileService.OpenFile(onnxFile.File.Path)
	if err != nil {
		return err
	}
	defer closeFile(f)

	if err := s.CCDBService.UploadFile(sor, eor, uploadFilename, f); err != nil {
		return mapCCDBError("UploadFile", err)
	}

	return nil
}
