package repository

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"gorm.io/gorm"
)

type TrainingTaskResultRepository interface {
	Create(ttr *models.TrainingTaskResult) error
	GetByID(id uint) (*models.TrainingTaskResult, error)
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
	var trainingTask models.TrainingTaskResult
	if err := r.db.First(&trainingTask, id).Error; err != nil {
		return nil, err
	}
	return &trainingTask, nil
}

func (r *trainingTaskResultRepository) GetAll(taskId uint) ([]models.TrainingTaskResult, error) {
	var trainingTasks []models.TrainingTaskResult
	if err := r.db.Order("\"created_at\" desc").Find(&trainingTasks, r.db.Where("\"training_task_id\" = ?", taskId)).Error; err != nil {
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
