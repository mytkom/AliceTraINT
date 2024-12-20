package repository

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"gorm.io/gorm"
)

type TrainingDatasetRepository interface {
	Create(trainingDataset *models.TrainingDataset) error
	GetByID(id uint) (*models.TrainingDataset, error)
	GetAll() ([]models.TrainingDataset, error)
	GetAllUser(userId uint) ([]models.TrainingDataset, error)
	Delete(userId uint, id uint) error
}

type trainingDatasetRepository struct {
	db *gorm.DB
}

func NewTrainingDatasetRepository(db *gorm.DB) TrainingDatasetRepository {
	return &trainingDatasetRepository{db: db}
}

func (r *trainingDatasetRepository) Create(trainingDataset *models.TrainingDataset) error {
	return r.db.Create(trainingDataset).Error
}

func (r *trainingDatasetRepository) allWithDependencies() *gorm.DB {
	return r.db.Joins("User").Order("\"training_datasets\".\"created_at\" desc")
}

func (r *trainingDatasetRepository) GetByID(id uint) (*models.TrainingDataset, error) {
	var trainingDataset models.TrainingDataset
	if err := r.allWithDependencies().First(&trainingDataset, id).Error; err != nil {
		return nil, err
	}
	return &trainingDataset, nil
}

func (r *trainingDatasetRepository) GetAll() ([]models.TrainingDataset, error) {
	var trainingDatasets []models.TrainingDataset
	if err := r.allWithDependencies().Find(&trainingDatasets).Error; err != nil {
		return nil, err
	}
	return trainingDatasets, nil
}

func (r *trainingDatasetRepository) GetAllUser(userId uint) ([]models.TrainingDataset, error) {
	var trainingDatasets []models.TrainingDataset
	if err := r.allWithDependencies().Find(&trainingDatasets, r.db.Where("\"user_id\" = ?", userId)).Error; err != nil {
		return nil, err
	}
	return trainingDatasets, nil
}

func (r *trainingDatasetRepository) Delete(userId uint, id uint) error {
	return r.db.Where("\"user_id\" = ?", userId).Delete(&models.TrainingDataset{}, id).Error
}
