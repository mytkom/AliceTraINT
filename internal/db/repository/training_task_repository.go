package repository

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type TrainingTaskRepository interface {
	Create(trainingTask *models.TrainingTask) error
	GetByID(id uint) (*models.TrainingTask, error)
	GetAll() ([]models.TrainingTask, error)
	GetAllUser(userId uint) ([]models.TrainingTask, error)
	GetFirstQueued() (*models.TrainingTask, error)
	Update(trainingTask *models.TrainingTask) error
	Delete(userId uint, id uint) error
}

type trainingTaskRepository struct {
	db *gorm.DB
}

func NewTrainingTaskRepository(db *gorm.DB) TrainingTaskRepository {
	return &trainingTaskRepository{db: db}
}

func (r *trainingTaskRepository) Create(trainingTask *models.TrainingTask) error {
	return r.db.Create(trainingTask).Error
}

func (r *trainingTaskRepository) withDependencies() *gorm.DB {
	return r.db.
		Preload("TrainingDataset", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		}).
		Joins("User")
}

func (r *trainingTaskRepository) GetByID(id uint) (*models.TrainingTask, error) {
	var trainingTask models.TrainingTask
	if err := r.withDependencies().First(&trainingTask, id).Error; err != nil {
		return nil, err
	}
	return &trainingTask, nil
}

func (r *trainingTaskRepository) GetFirstQueued() (*models.TrainingTask, error) {
	var trainingTask models.TrainingTask
	if err := r.withDependencies().Where("\"status\" = ?", models.Queued).Order("\"training_tasks\".\"created_at\" asc").First(&trainingTask).Error; err != nil {
		return nil, err
	}

	return &trainingTask, nil
}

func (r *trainingTaskRepository) GetAll() ([]models.TrainingTask, error) {
	var trainingTasks []models.TrainingTask
	if err := r.withDependencies().Order("\"training_tasks\".\"created_at\" desc").Find(&trainingTasks).Error; err != nil {
		return nil, err
	}
	return trainingTasks, nil
}

func (r *trainingTaskRepository) GetAllUser(userId uint) ([]models.TrainingTask, error) {
	var trainingTasks []models.TrainingTask
	if err := r.withDependencies().Order("\"training_tasks\".\"created_at\" desc").Find(&trainingTasks, r.db.Where(&models.TrainingTask{UserId: userId})).Error; err != nil {
		return nil, err
	}
	return trainingTasks, nil
}

func (r *trainingTaskRepository) Update(trainingTask *models.TrainingTask) error {
	return r.db.Save(trainingTask).Error
}

func (r *trainingTaskRepository) Delete(userId uint, id uint) error {
	return r.db.Where("\"user_id\" = ?", userId).Delete(&models.TrainingTask{}, id).Error
}

type MockTrainingTaskRepository struct {
	mock.Mock
}

func NewMockTrainingTaskRepository() *MockTrainingTaskRepository {
	return &MockTrainingTaskRepository{}
}

func (m *MockTrainingTaskRepository) Create(trainingTask *models.TrainingTask) error {
	args := m.Called(trainingTask)
	return args.Error(0)
}

func (m *MockTrainingTaskRepository) GetAll() ([]models.TrainingTask, error) {
	args := m.Called()

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]models.TrainingTask), args.Error(1)
}

func (m *MockTrainingTaskRepository) GetAllUser(userId uint) ([]models.TrainingTask, error) {
	args := m.Called(userId)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]models.TrainingTask), args.Error(1)
}

func (m *MockTrainingTaskRepository) GetByID(id uint) (*models.TrainingTask, error) {
	args := m.Called(id)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.TrainingTask), args.Error(1)
}
func (m *MockTrainingTaskRepository) GetFirstQueued() (*models.TrainingTask, error) {
	args := m.Called()

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*models.TrainingTask), args.Error(1)
}

func (m *MockTrainingTaskRepository) Update(trainingTask *models.TrainingTask) error {
	args := m.Called(trainingTask)
	return args.Error(0)
}

func (m *MockTrainingTaskRepository) Delete(userId uint, id uint) error {
	args := m.Called(userId, id)
	return args.Error(0)
}
