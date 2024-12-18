package repository

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/mock"
)

type MockTrainingMachineRepository struct {
	mock.Mock
}

func (m *MockTrainingMachineRepository) Create(trainingMachine *models.TrainingMachine) error {
	args := m.Called(trainingMachine)
	return args.Error(0)
}

func (m *MockTrainingMachineRepository) GetAll() ([]models.TrainingMachine, error) {
	args := m.Called()

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]models.TrainingMachine), args.Error(1)
}

func (m *MockTrainingMachineRepository) GetAllUser(userId uint) ([]models.TrainingMachine, error) {
	args := m.Called(userId)
	return args.Get(0).([]models.TrainingMachine), args.Error(1)
}

func (m *MockTrainingMachineRepository) GetByID(id uint) (*models.TrainingMachine, error) {
	args := m.Called(id)
	return args.Get(0).(*models.TrainingMachine), args.Error(1)
}

func (m *MockTrainingMachineRepository) Update(tm *models.TrainingMachine) error {
	args := m.Called(tm)
	return args.Error(0)
}

func (m *MockTrainingMachineRepository) Delete(userId uint, id uint) error {
	args := m.Called(userId, id)
	return args.Error(0)
}
