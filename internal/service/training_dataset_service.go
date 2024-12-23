package service

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"gorm.io/gorm"
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

var errCCDBUnreachable = NewErrExternalServiceTimeout("CCDB")
var errDatasetNotFound = NewErrHandlerNotFound("TrainingDataset")

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
	td, err := s.TrainingDataset.GetByID(id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errDatasetNotFound
		} else {
			return nil, errInternalServerError
		}
	}

	return td, nil
}

func (s *TrainingDatasetService) Create(td *models.TrainingDataset) error {
	if len(td.AODFiles) == 0 {
		return &ErrHandlerValidation{
			Field: "AODFiles",
			Msg:   errMsgMissing,
		}
	}

	err := s.TrainingDataset.Create(td)

	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return &ErrHandlerValidation{
				Field: "Name",
				Msg:   errMsgNotUnique,
			}
		} else {
			return errInternalServerError
		}
	}

	return nil
}

func (s *TrainingDatasetService) Delete(userId uint, id uint) error {
	err := s.TrainingDataset.Delete(userId, id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errDatasetNotFound
		} else {
			return errInternalServerError
		}
	}

	return nil
}

func (s *TrainingDatasetService) ExploreDirectory(path string) (*jalien.DirectoryContents, string, error) {
	if path == "" {
		path = "/"
	}

	dirContents, err := s.JAliEn.ListAndParseDirectory(path)
	if err != nil {
		return nil, "", handleCCDBError(err)
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
	aodFiles, err := s.JAliEn.FindAODFiles(path)

	if err != nil {
		return nil, handleCCDBError(err)
	}

	return aodFiles, nil
}
