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
	SelectRandomAODSubset(path string, runCount, filesPerRun int, minSizeBytes uint64) ([]jalien.AODFile, error)
}

type TrainingDatasetService struct {
	*repository.RepositoryContext
	JAliEn          IJAliEnService
	CCDBService     ICCDBService
	MonalisaService IMonalisaService
}

var errCCDBUnreachable = NewErrExternalServiceTimeout("CCDB")
var errJAlienUnreachable = NewErrExternalServiceTimeout("JAlien")
var errMonalisaUnreachable = NewErrExternalServiceTimeout("Monalisa")
var errDatasetNotFound = NewErrHandlerNotFound("TrainingDataset")

func NewTrainingDatasetService(repo *repository.RepositoryContext, jalien IJAliEnService, ccdbService ICCDBService, monalisaService IMonalisaService) *TrainingDatasetService {
	return &TrainingDatasetService{
		RepositoryContext: repo,
		JAliEn:            jalien,
		CCDBService:       ccdbService,
		MonalisaService:   monalisaService,
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
			return nil, InternalError("get training dataset by id", err)
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

	timestampFrom, timestampTo, err := resolveDatasetTimestampRange(td.AODFiles, s.MonalisaService, s.CCDBService)
	if err != nil {
		var external *ErrExternalServiceTimeout
		if errors.As(err, &external) {
			return err
		}
		return InternalError("resolve training dataset timestamp range", err)
	}
	td.AnchorProdTimestampFrom = timestampFrom
	td.AnchorProdTimestampTo = timestampTo

	isONe, err := resolveDatasetIsONe(td.AODFiles, s.MonalisaService)
	if err != nil {
		var external *ErrExternalServiceTimeout
		if errors.As(err, &external) {
			return err
		}
		return InternalError("resolve training dataset IsONe", err)
	}
	td.IsONe = isONe

	err = s.TrainingDataset.Create(td)

	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return &ErrHandlerValidation{
				Field: "Name",
				Msg:   errMsgNotUnique,
			}
		} else {
			return InternalError("create training dataset", err)
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
			return InternalError("delete training dataset", err)
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
		return nil, "", mapJAlienError("ListAndParseDirectory", err)
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
		return nil, mapJAlienError("FindAODFiles", err)
	}

	return aodFiles, nil
}

func resolveDatasetIsONe(aodFiles []jalien.AODFile, monalisaService IMonalisaService) (bool, error) {
	runList, err := monalisaService.GetRunList()
	if err != nil {
		return false, mapMonalisaError("GetRunList", err)
	}

	for _, aod := range aodFiles {
		if aod.LHCPeriod != "" {
			desc := runList.TagToDesc[aod.LHCPeriod]
			if strings.Contains(desc, "O-O") || strings.Contains(desc, "OO") ||
				strings.Contains(desc, "Ne-Ne") || strings.Contains(desc, "NeNe") {
				return true, nil
			}
		}
	}

	return false, nil
}
