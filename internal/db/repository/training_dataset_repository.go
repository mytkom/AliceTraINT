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
	Update(userId uint, trainingDataset *models.TrainingDataset) error
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

func (r *trainingDatasetRepository) GetByID(id uint) (*models.TrainingDataset, error) {
	var trainingDataset models.TrainingDataset
	if err := r.db.First(&trainingDataset, id).Error; err != nil {
		return nil, err
	}
	return &trainingDataset, nil
}

func (r *trainingDatasetRepository) GetAll() ([]models.TrainingDataset, error) {
	var trainingDatasets []models.TrainingDataset
	if err := r.db.Find(&trainingDatasets).Error; err != nil {
		return nil, err
	}
	return trainingDatasets, nil
}

func (r *trainingDatasetRepository) GetAllUser(userId uint) ([]models.TrainingDataset, error) {
	var trainingDatasets []models.TrainingDataset
	if err := r.db.Order("\"created_at\" desc").Where("\"user_id\" = ?", userId).Find(&trainingDatasets).Error; err != nil {
		return nil, err
	}
	return trainingDatasets, nil
}

func (r *trainingDatasetRepository) Update(userId uint, trainingDataset *models.TrainingDataset) error {
	// TODO: authorize user

	return r.db.Save(trainingDataset).Error
}

func (r *trainingDatasetRepository) Delete(userId uint, id uint) error {
	return r.db.Where("\"user_id\" = ?", userId).Delete(&models.TrainingDataset{}, id).Error
}
