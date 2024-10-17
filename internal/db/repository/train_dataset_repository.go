package repository

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"gorm.io/gorm"
)

type TrainDatasetRepository interface {
	Create(trainDataset *models.TrainDataset) error
	GetByID(id uint) (*models.TrainDataset, error)
	GetAll() ([]models.TrainDataset, error)
	GetAllUser(userId uint) ([]models.TrainDataset, error)
	Update(userId uint, trainDataset *models.TrainDataset) error
	Delete(userId uint, id uint) error
}

type trainDatasetRepository struct {
	db *gorm.DB
}

func NewTrainDatasetRepository(db *gorm.DB) TrainDatasetRepository {
	return &trainDatasetRepository{db: db}
}

func (r *trainDatasetRepository) Create(trainDataset *models.TrainDataset) error {
	return r.db.Create(trainDataset).Error
}

func (r *trainDatasetRepository) GetByID(id uint) (*models.TrainDataset, error) {
	var trainDataset models.TrainDataset
	if err := r.db.First(&trainDataset, id).Error; err != nil {
		return nil, err
	}
	return &trainDataset, nil
}

func (r *trainDatasetRepository) GetAll() ([]models.TrainDataset, error) {
	var trainDatasets []models.TrainDataset
	if err := r.db.Find(&trainDatasets).Error; err != nil {
		return nil, err
	}
	return trainDatasets, nil
}

func (r *trainDatasetRepository) GetAllUser(userId uint) ([]models.TrainDataset, error) {
	var trainDatasets []models.TrainDataset
	if err := r.db.Where("\"user_id\" = ?", userId).Find(&trainDatasets).Error; err != nil {
		return nil, err
	}
	return trainDatasets, nil
}

func (r *trainDatasetRepository) Update(userId uint, trainDataset *models.TrainDataset) error {
	// TODO: authorize user

	return r.db.Save(trainDataset).Error
}

func (r *trainDatasetRepository) Delete(userId uint, id uint) error {
	return r.db.Where("\"user_id\" = ?", userId).Delete(&models.TrainDataset{}, id).Error
}
