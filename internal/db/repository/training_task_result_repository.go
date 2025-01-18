package repository

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type TrainingTaskResultRepository interface {
	Create(ttr *models.TrainingTaskResult) error
	GetByID(id uint) (*models.TrainingTaskResult, error)
	GetByType(ttId uint, resultType models.TrainingTaskResultType) ([]models.TrainingTaskResult, error)
	GetAll(taskId uint) ([]models.TrainingTaskResult, error)
	Update(ttr *models.TrainingTaskResult) error
	Delete(id uint) error
}

type trainingTaskResultRepository struct {
	db *gorm.DB
}

func NewTrainingTaskResultRepository(db *gorm.DB) TrainingTaskResultRepository {
	return &trainingTaskResultRepository{db: db}
}

func (r *trainingTaskResultRepository) Create(ttr *models.TrainingTaskResult) error {
	return r.db.Create(ttr).Error
}

func (r *trainingTaskResultRepository) GetByID(id uint) (*models.TrainingTaskResult, error) {
	var ttr models.TrainingTaskResult
	if err := r.db.Table("training_task_results").Joins("File").First(&ttr, id).Error; err != nil {
		return nil, err
	}
	return &ttr, nil
}

func (r *trainingTaskResultRepository) getAll(taskId uint) *gorm.DB {
	return r.db.Order("\"training_task_results\".\"created_at\" desc").Where("\"training_task_id\" = ?", taskId).Joins("File")
}

func (r *trainingTaskResultRepository) GetByType(ttId uint, resultType models.TrainingTaskResultType) ([]models.TrainingTaskResult, error) {
	var trainingTasks []models.TrainingTaskResult
	if err := r.getAll(ttId).Find(&trainingTasks, r.db.Where("\"type\" = ?", int(resultType))).Error; err != nil {
		return nil, err
	}
	return trainingTasks, nil
}

func (r *trainingTaskResultRepository) GetAll(taskId uint) ([]models.TrainingTaskResult, error) {
	var trainingTasks []models.TrainingTaskResult
	if err := r.getAll(taskId).Find(&trainingTasks).Error; err != nil {
		return nil, err
	}
	return trainingTasks, nil
}

func (r *trainingTaskResultRepository) Update(ttr *models.TrainingTaskResult) error {
	return r.db.Save(ttr).Error
}

func (r *trainingTaskResultRepository) Delete(id uint) error {
	return r.db.Delete(&models.TrainingTaskResult{}, id).Error
}

type MockTrainingTaskResultRepository struct {
	mock.Mock
}

func NewMockTrainingTaskResultRepository() *MockTrainingTaskResultRepository {
	return &MockTrainingTaskResultRepository{}
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
