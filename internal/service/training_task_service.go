package service

import (
	"errors"
	"fmt"
	"log"

	"github.com/mytkom/AliceTraINT/internal/ccdb"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
)

type TrainingTaskWithResults struct {
	TrainingTask *models.TrainingTask
	ImageFiles   []models.TrainingTaskResult
	OnnxFiles    []models.TrainingTaskResult
}

type TrainingTaskHelpers struct {
	TrainingDatasets []models.TrainingDataset
	FieldConfigs     NNFieldConfigs
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
	CCDBService ICCDBService
	FileService IFileService
	NNArch      INNArchService
}

func NewTrainingTaskService(repo *repository.RepositoryContext, ccdbService ICCDBService, fileService IFileService, nnArch INNArchService) *TrainingTaskService {
	return &TrainingTaskService{
		RepositoryContext: repo,
		CCDBService:       ccdbService,
		FileService:       fileService,
		NNArch:            nnArch,
	}
}

func (s *TrainingTaskService) Create(tt *models.TrainingTask) error {
	return s.TrainingTask.Create(tt)
}

func (s *TrainingTaskService) GetAll(loggedUserId uint, userScoped bool) ([]models.TrainingTask, error) {
	var trainingTasks []models.TrainingTask
	var err error

	if userScoped {
		trainingTasks, err = s.TrainingTask.GetAllUser(loggedUserId)
		if err != nil {
			return nil, err
		}
	} else {
		trainingTasks, err = s.TrainingTask.GetAll()
		if err != nil {
			return nil, err
		}
	}

	return trainingTasks, nil
}

func (s *TrainingTaskService) GetHelpers(loggedUserId uint) (*TrainingTaskHelpers, error) {
	trainingDatasets, err := s.TrainingDataset.GetAllUser(loggedUserId)
	if err != nil {
		return nil, err
	}

	return &TrainingTaskHelpers{
		TrainingDatasets: trainingDatasets,
		FieldConfigs:     s.NNArch.GetFieldConfigs(),
	}, nil
}

func (s *TrainingTaskService) GetByID(id uint) (*TrainingTaskWithResults, error) {
	trainingTask, err := s.TrainingTask.GetByID(uint(id))
	if err != nil {
		return nil, err
	}

	var imageFiles []models.TrainingTaskResult
	if trainingTask.Status >= models.Completed {
		imageFiles, err = s.TrainingTaskResult.GetByType(trainingTask.ID, models.Image)
		if err != nil {
			return nil, err
		}
	} else {
		imageFiles = nil
	}

	var onnxFiles []models.TrainingTaskResult
	if trainingTask.Status >= models.Benchmarking {
		onnxFiles, err = s.TrainingTaskResult.GetByType(trainingTask.ID, models.Onnx)
		if err != nil {
			return nil, err
		}
	} else {
		onnxFiles = nil
	}

	return &TrainingTaskWithResults{
		TrainingTask: trainingTask,
		ImageFiles:   imageFiles,
		OnnxFiles:    onnxFiles,
	}, nil
}

func (s *TrainingTaskService) UploadOnnxResults(id uint) error {
	trainingTask, err := s.TrainingTask.GetByID(id)
	if err != nil {
		return err
	}

	if trainingTask.Status < models.Completed {
		return &ErrHandlerValidation{
			Field: "Status",
			Msg:   "must be completed or uploaded",
		}
	}

	smallestRun, greatestRun, err := s.findRunNumberRange(trainingTask)
	if err != nil {
		return err
	}

	firstRunInfo, lastRunInfo, err := s.getRunInfoRange(smallestRun, greatestRun)
	if err != nil {
		return err
	}

	mappedOnnxFiles, err := s.filterOnnxFiles(trainingTask.ID)
	if err != nil {
		return err
	}

	for uploadName, file := range mappedOnnxFiles {
		if err := s.uploadOnnxFile(firstRunInfo.SOR, lastRunInfo.EOR, file, uploadName); err != nil {
			return err
		}
	}

	trainingTask.Status = models.Uploaded
	if err := s.TrainingTask.Update(trainingTask); err != nil {
		return err
	}

	return nil
}

func (s *TrainingTaskService) findRunNumberRange(task *models.TrainingTask) (uint64, uint64, error) {
	var smallestRun, greatestRun uint64
	initialized := false

	for _, aod := range task.TrainingDataset.AODFiles {
		if !initialized || aod.RunNumber < smallestRun {
			smallestRun = aod.RunNumber
		}
		if !initialized || aod.RunNumber > greatestRun {
			greatestRun = aod.RunNumber
		}
		initialized = true
	}

	if !initialized {
		return 0, 0, errors.New("unexpected behaviour: empty training dataset")
	}

	return smallestRun, greatestRun, nil
}

func (s *TrainingTaskService) getRunInfoRange(smallestRun, greatestRun uint64) (*ccdb.RunInformation, *ccdb.RunInformation, error) {
	firstRunInfo, err := s.CCDBService.GetRunInformation(smallestRun)
	if err != nil {
		return nil, nil, handleCCDBError(err)
	}

	lastRunInfo, err := s.CCDBService.GetRunInformation(greatestRun)
	if err != nil {
		return nil, nil, handleCCDBError(err)
	}

	log.Printf("From run %d, SOR %d", firstRunInfo.RunNumber, firstRunInfo.SOR)
	log.Printf("to run %d, EOR %d", lastRunInfo.RunNumber, lastRunInfo.EOR)

	return firstRunInfo, lastRunInfo, nil
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
			log.Printf("expected file not present: %s", localName)
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
		return handleCCDBError(err)
	}

	return nil
}
