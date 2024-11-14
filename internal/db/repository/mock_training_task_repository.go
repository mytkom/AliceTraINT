package repository

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/mock"
)

type MockTrainingTaskRepository struct {
	mock.Mock
}

func (m *MockTrainingTaskRepository) Create(trainingTask *models.TrainingTask) error {
	args := m.Called(trainingTask)
	return args.Error(0)
}

func (m *MockTrainingTaskRepository) GetAll() ([]models.TrainingTask, error) {
	args := m.Called()
	return args.Get(0).([]models.TrainingTask), args.Error(1)
}

func (m *MockTrainingTaskRepository) GetAllUser(userId uint) ([]models.TrainingTask, error) {
	args := m.Called()
	return args.Get(0).([]models.TrainingTask), args.Error(1)
}

func (m *MockTrainingTaskRepository) GetByID(id uint) (*models.TrainingTask, error) {
	args := m.Called()
	return args.Get(0).(*models.TrainingTask), args.Error(1)
}

func (m *MockTrainingTaskRepository) Update(updatedTrainingTask *models.TrainingTask) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTrainingTaskRepository) Delete(userId uint, id uint) error {
	args := m.Called()
	return args.Error(0)
}
