package service

import (
	"path/filepath"
	"strings"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/jalien"
)

type ITrainingDatasetService interface {
	GetAll(loggedUserId uint, userScoped bool) ([]models.TrainingDataset, error)
	GetByID(id uint) (*models.TrainingDataset, error)
	Create(td *models.TrainingDataset) error
	Delete(userId uint, id uint) error
	ExploreDirectory(path string) (*jalien.DirectoryContents, string, error)
	FindAods(path string) ([]jalien.AODFile, error)
}

type TrainingDatasetService struct {
	*repository.RepositoryContext
	JAliEn IJAliEnService
}

func NewTrainingDatasetService(repo *repository.RepositoryContext, jalien IJAliEnService) *TrainingDatasetService {
	return &TrainingDatasetService{
		RepositoryContext: repo,
		JAliEn:            jalien,
	}
}

func (s *TrainingDatasetService) GetAll(loggedUserId uint, userScoped bool) ([]models.TrainingDataset, error) {
	var trainingTasks []models.TrainingDataset
	var err error

	if userScoped {
		trainingTasks, err = s.TrainingDataset.GetAllUser(loggedUserId)
		if err != nil {
			return nil, err
		}
	} else {
		trainingTasks, err = s.TrainingDataset.GetAll()
		if err != nil {
			return nil, err
		}
	}

	return trainingTasks, nil
}

func (s *TrainingDatasetService) GetByID(id uint) (*models.TrainingDataset, error) {
	return s.TrainingDataset.GetByID(id)
}

func (s *TrainingDatasetService) Create(td *models.TrainingDataset) error {
	return s.TrainingDataset.Create(td)
}

func (s *TrainingDatasetService) Delete(userId uint, id uint) error {
	return s.TrainingDataset.Delete(userId, id)
}

func (s *TrainingDatasetService) ExploreDirectory(path string) (*jalien.DirectoryContents, string, error) {
	if path == "" {
		path = "/"
	}

	dirContents, err := s.JAliEn.ListAndParseDirectory(path)
	if err != nil {
		return nil, "", err
	}

	parentDir := "/"
	if path != "/" {
		parentDir = filepath.Dir(strings.TrimSuffix(path, "/"))
		if parentDir != "/" {
			parentDir += "/"
		}
	}

	return dirContents, parentDir, nil
}

func (s *TrainingDatasetService) FindAods(path string) ([]jalien.AODFile, error) {
	return s.JAliEn.FindAODFiles(path)
}
