package repository

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/mock"
)

type MockTrainingTaskResultRepository struct {
	mock.Mock
}

func (m *MockTrainingTaskResultRepository) Create(ttr *models.TrainingTaskResult) error {
	args := m.Called(ttr)
	return args.Error(0)
}

func (m *MockTrainingTaskResultRepository) GetAll() ([]models.TrainingTaskResult, error) {
	args := m.Called()
	return args.Get(0).([]models.TrainingTaskResult), args.Error(1)
}

func (m *MockTrainingTaskResultRepository) GetByID(id uint) (*models.TrainingTaskResult, error) {
	args := m.Called()
	return args.Get(0).(*models.TrainingTaskResult), args.Error(1)
}

func (m *MockTrainingTaskResultRepository) Update(ttr *models.TrainingTaskResult) error {
	args := m.Called(ttr)
	return args.Error(0)
}

func (m *MockTrainingTaskResultRepository) Delete(userId uint, id uint) error {
	args := m.Called()
	return args.Error(0)
}
