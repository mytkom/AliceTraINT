package repository

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/mock"
)

type MockTrainDatasetRepository struct {
	mock.Mock
}

func (m *MockTrainDatasetRepository) Create(trainDataset *models.TrainDataset) error {
	args := m.Called(trainDataset)
	return args.Error(0)
}

func (m *MockTrainDatasetRepository) GetAll() ([]models.TrainDataset, error) {
	args := m.Called()
	return args.Get(0).([]models.TrainDataset), args.Error(1)
}

func (m *MockTrainDatasetRepository) GetAllUser(userId uint) ([]models.TrainDataset, error) {
	args := m.Called()
	return args.Get(0).([]models.TrainDataset), args.Error(1)
}

func (m *MockTrainDatasetRepository) GetByID(id uint) (*models.TrainDataset, error) {
	args := m.Called()
	return args.Get(0).(*models.TrainDataset), args.Error(1)
}

func (m *MockTrainDatasetRepository) Update(updatedTrainDataset *models.TrainDataset) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTrainDatasetRepository) Delete(userId uint, id uint) error {
	args := m.Called()
	return args.Error(0)
}
