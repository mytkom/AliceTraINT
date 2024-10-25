package repository

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"gorm.io/gorm"
)

type TrainingTaskRepository interface {
	Create(trainingTask *models.TrainingTask) error
	GetByID(id uint) (*models.TrainingTask, error)
	GetAll() ([]models.TrainingTask, error)
	GetAllUser(userId uint) ([]models.TrainingTask, error)
	Update(userId uint, trainingTask *models.TrainingTask) error
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

func (r *trainingTaskRepository) GetByID(id uint) (*models.TrainingTask, error) {
	var trainingTask models.TrainingTask
	if err := r.db.First(&trainingTask, id).Error; err != nil {
		return nil, err
	}
	return &trainingTask, nil
}

func (r *trainingTaskRepository) GetAll() ([]models.TrainingTask, error) {
	var trainingTasks []models.TrainingTask
	if err := r.db.Find(&trainingTasks).Error; err != nil {
		return nil, err
	}
	return trainingTasks, nil
}

func (r *trainingTaskRepository) GetAllUser(userId uint) ([]models.TrainingTask, error) {
	var trainingTasks []models.TrainingTask
	if err := r.db.Order("\"created_at\" desc").Where("\"user_id\" = ?", userId).Find(&trainingTasks).Error; err != nil {
		return nil, err
	}
	return trainingTasks, nil
}

func (r *trainingTaskRepository) Update(userId uint, trainingTask *models.TrainingTask) error {
	// TODO: authorize user

	return r.db.Save(trainingTask).Error
}

func (r *trainingTaskRepository) Delete(userId uint, id uint) error {
	return r.db.Where("\"user_id\" = ?", userId).Delete(&models.TrainingTask{}, id).Error
}
