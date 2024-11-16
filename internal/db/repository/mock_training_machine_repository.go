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
	return args.Get(0).([]models.TrainingMachine), args.Error(1)
}

func (m *MockTrainingMachineRepository) GetAllUser(userId uint) ([]models.TrainingMachine, error) {
	args := m.Called()
	return args.Get(0).([]models.TrainingMachine), args.Error(1)
}

func (m *MockTrainingMachineRepository) GetByID(id uint) (*models.TrainingMachine, error) {
	args := m.Called()
	return args.Get(0).(*models.TrainingMachine), args.Error(1)
}

func (m *MockTrainingMachineRepository) Delete(userId uint, id uint) error {
	args := m.Called()
	return args.Error(0)
}
