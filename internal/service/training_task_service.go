package service

import (
	"errors"
	"log"
	"slices"
	"sort"

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

	imageFiles, err := s.TrainingTaskResult.GetByType(trainingTask.ID, models.Image)
	if err != nil {
		return nil, err
	}

	onnxFiles, err := s.TrainingTaskResult.GetByType(trainingTask.ID, models.Onnx)
	if err != nil {
		return nil, err
	}

	return &TrainingTaskWithResults{
		TrainingTask: trainingTask,
		ImageFiles:   imageFiles,
		OnnxFiles:    onnxFiles,
	}, nil
}

func (s *TrainingTaskService) UploadOnnxResults(id uint) error {
	trainingTask, err := s.TrainingTask.GetByID(uint(id))
	if err != nil {
		return err
	}

	runs := []uint64{}
	for _, aod := range trainingTask.TrainingDataset.AODFiles {
		if !slices.Contains(runs, aod.RunNumber) {
			runs = append(runs, aod.RunNumber)
		}
	}

	if len(runs) == 0 {
		return errors.New("unexpected behaviour: empty training dataset")
	}

	sort.Slice(runs, func(i, j int) bool {
		return runs[i] < runs[j]
	})

	firstRunInfo, err := s.CCDBService.GetRunInformation(runs[0])
	if err != nil {
		return err
	}

	lastRunInfo, err := s.CCDBService.GetRunInformation(runs[len(runs)-1])
	if err != nil {
		return err
	}

	log.Printf("From run %d, SOR %d", firstRunInfo.RunNumber, firstRunInfo.SOR)
	log.Printf("to run %d, EOR %d", lastRunInfo.RunNumber, lastRunInfo.SOR)

	onnxFiles, err := s.TrainingTaskResult.GetByType(trainingTask.ID, models.Onnx)
	if err != nil {
		return err
	}

	for _, onnxFile := range onnxFiles {
		file, close, err := s.FileService.OpenFile(onnxFile.File.Path)
		if err != nil {
			return err
		}
		defer close(file)

		if upload_filename, ok := s.NNArch.GetUploadFilename(onnxFile.Name); ok {
			err = s.CCDBService.UploadFile(firstRunInfo.SOR, lastRunInfo.EOR, upload_filename, file)
			if err != nil {
				return err
			}
		} else {
			log.Printf("not expected file: %s", onnxFile.Name)
			continue
		}
	}

	trainingTask.Status = models.Uploaded
	s.TrainingTask.Update(trainingTask)

	return nil
}
