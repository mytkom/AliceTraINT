package repository

import "gorm.io/gorm"

type RepositoryContext struct {
	User               UserRepository
	TrainingMachine    TrainingMachineRepository
	TrainingDataset    TrainingDatasetRepository
	TrainingTask       TrainingTaskRepository
	TrainingTaskResult TrainingTaskResultRepository
}

func NewRepositoryContext(db *gorm.DB) *RepositoryContext {
	return &RepositoryContext{
		User:               NewUserRepository(db),
		TrainingMachine:    NewTrainingMachineRepository(db),
		TrainingDataset:    NewTrainingDatasetRepository(db),
		TrainingTask:       NewTrainingTaskRepository(db),
		TrainingTaskResult: NewTrainingTaskResultRepository(db),
	}
}
