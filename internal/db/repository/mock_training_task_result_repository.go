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

func (m *MockTrainingTaskResultRepository) GetAll(id uint) ([]models.TrainingTaskResult, error) {
	args := m.Called(id)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]models.TrainingTaskResult), args.Error(1)
}

func (m *MockTrainingTaskResultRepository) GetByType(ttId uint, resultType models.TrainingTaskResultType) ([]models.TrainingTaskResult, error) {
	args := m.Called(ttId, resultType)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]models.TrainingTaskResult), args.Error(1)
}

func (m *MockTrainingTaskResultRepository) GetByID(id uint) (*models.TrainingTaskResult, error) {
	args := m.Called(id)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.TrainingTaskResult), args.Error(1)
}

func (m *MockTrainingTaskResultRepository) Update(ttr *models.TrainingTaskResult) error {
	args := m.Called(ttr)
	return args.Error(0)
}

func (m *MockTrainingTaskResultRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}
