package repository

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"gorm.io/gorm"
)

type TrainingMachineRepository interface {
	Create(trainingMachine *models.TrainingMachine) error
	GetByID(id uint) (*models.TrainingMachine, error)
	GetAll() ([]models.TrainingMachine, error)
	GetAllUser(userId uint) ([]models.TrainingMachine, error)
	Update(userId uint, trainingMachine *models.TrainingMachine) error
	Delete(userId uint, id uint) error
}

type trainingMachineRepository struct {
	db *gorm.DB
}

func NewTrainingMachineRepository(db *gorm.DB) TrainingMachineRepository {
	return &trainingMachineRepository{db: db}
}

func (r *trainingMachineRepository) Create(trainingMachine *models.TrainingMachine) error {
	return r.db.Create(trainingMachine).Error
}

func (r *trainingMachineRepository) withDependencies() *gorm.DB {
	return r.db.Joins("User")
}

func (r *trainingMachineRepository) GetByID(id uint) (*models.TrainingMachine, error) {
	var trainingTask models.TrainingMachine
	if err := r.withDependencies().First(&trainingTask, id).Error; err != nil {
		return nil, err
	}
	return &trainingTask, nil
}

func (r *trainingMachineRepository) GetAll() ([]models.TrainingMachine, error) {
	var trainingTasks []models.TrainingMachine
	if err := r.withDependencies().Order("\"last_activity_at\" desc").Find(&trainingTasks).Error; err != nil {
		return nil, err
	}
	return trainingTasks, nil
}

func (r *trainingMachineRepository) GetAllUser(userId uint) ([]models.TrainingMachine, error) {
	var trainingTasks []models.TrainingMachine
	if err := r.withDependencies().Order("\"last_activity_at\" desc").Find(&trainingTasks, r.db.Where(&models.TrainingMachine{UserId: userId})).Error; err != nil {
		return nil, err
	}
	return trainingTasks, nil
}

func (r *trainingMachineRepository) Update(userId uint, trainingMachine *models.TrainingMachine) error {
	// TODO: authorize user

	return r.db.Save(trainingMachine).Error
}

func (r *trainingMachineRepository) Delete(userId uint, id uint) error {
	return r.db.Where("\"user_id\" = ?", userId).Delete(&models.TrainingMachine{}, id).Error
}
