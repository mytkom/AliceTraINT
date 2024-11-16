package repository

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/mock"
)

type MockTrainingDatasetRepository struct {
	mock.Mock
}

func (m *MockTrainingDatasetRepository) Create(trainingDataset *models.TrainingDataset) error {
	args := m.Called(trainingDataset)
	return args.Error(0)
}

func (m *MockTrainingDatasetRepository) GetAll() ([]models.TrainingDataset, error) {
	args := m.Called()
	return args.Get(0).([]models.TrainingDataset), args.Error(1)
}

func (m *MockTrainingDatasetRepository) GetAllUser(userId uint) ([]models.TrainingDataset, error) {
	args := m.Called()
	return args.Get(0).([]models.TrainingDataset), args.Error(1)
}

func (m *MockTrainingDatasetRepository) GetByID(id uint) (*models.TrainingDataset, error) {
	args := m.Called()
	return args.Get(0).(*models.TrainingDataset), args.Error(1)
}

func (m *MockTrainingDatasetRepository) Delete(userId uint, id uint) error {
	args := m.Called()
	return args.Error(0)
}
