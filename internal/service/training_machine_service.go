package service

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
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
		return "", err
	}

	return secretKey, nil
}

func (s *TrainingMachineService) GetAll(loggedUserId uint, userScoped bool) ([]models.TrainingMachine, error) {
	var trainingMachines []models.TrainingMachine
	var err error

	if userScoped {
		trainingMachines, err = s.TrainingMachine.GetAllUser(loggedUserId)
		if err != nil {
			return nil, err
		}
	} else {
		trainingMachines, err = s.TrainingMachine.GetAll()
		if err != nil {
			return nil, err
		}
	}

	return trainingMachines, nil
}

func (s *TrainingMachineService) GetByID(id uint) (*models.TrainingMachine, error) {
	return s.TrainingMachine.GetByID(uint(id))
}

func (s *TrainingMachineService) Delete(loggedUserId uint, id uint) error {
	return s.TrainingMachine.Delete(loggedUserId, id)
}
