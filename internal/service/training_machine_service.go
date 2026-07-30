package service

import (
	"errors"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"gorm.io/gorm"
)

type ITrainingMachineService interface {
	Create(tm *models.TrainingMachine) (string, error)
	GetAll(loggedUserId uint, userScoped bool) ([]models.TrainingMachine, error)
	GetByID(id uint) (*models.TrainingMachine, error)
	Delete(loggedUserId uint, id uint) error
}

type TrainingMachineService struct {
	*repository.RepositoryContext
	Hasher Hasher
}

func NewTrainingMachineService(repo *repository.RepositoryContext, hasher Hasher) *TrainingMachineService {
	return &TrainingMachineService{
		RepositoryContext: repo,
		Hasher:            hasher,
	}
}

var errMachineNotFound = NewErrHandlerNotFound("TrainingMachine")

func (s *TrainingMachineService) Create(tm *models.TrainingMachine) (string, error) {
	secretKey, err := s.Hasher.GenerateKey()
	if err != nil {
		return "", err
	}

	tm.SecretKeyHashed, err = s.Hasher.HashKey(secretKey)
	if err != nil {
		return "", err
	}

	err = s.TrainingMachine.Create(tm)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return "", &ErrHandlerValidation{
				Field: "Name",
				Msg:   errMsgNotUnique,
			}
		} else {
			return "", InternalError("create training machine", err)
		}
	}

	return secretKey, nil
}

func (s *TrainingMachineService) GetAll(loggedUserId uint, userScoped bool) ([]models.TrainingMachine, error) {
	var trainingMachines []models.TrainingMachine
	var err error

	if userScoped {
		trainingMachines, err = s.TrainingMachine.GetAllUser(loggedUserId)
		if err != nil {
			return nil, InternalError("get training machines", err)
		}
	} else {
		trainingMachines, err = s.TrainingMachine.GetAll()
		if err != nil {
			return nil, InternalError("get training machines", err)
		}
	}

	return trainingMachines, nil
}

func (s *TrainingMachineService) GetByID(id uint) (*models.TrainingMachine, error) {
	tm, err := s.TrainingMachine.GetByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errMachineNotFound
		} else {
			return nil, InternalError("get training machine by id", err)
		}
	}

	return tm, nil
}

func (s *TrainingMachineService) Delete(loggedUserId uint, id uint) error {
	err := s.TrainingMachine.Delete(loggedUserId, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errMachineNotFound
		} else {
			return InternalError("delete training machine", err)
		}
	}

	return nil
}
