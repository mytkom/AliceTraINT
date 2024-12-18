package repository

type RepositoryContext struct {
	User               UserRepository
	TrainingMachine    TrainingMachineRepository
	TrainingDataset    TrainingDatasetRepository
	TrainingTask       TrainingTaskRepository
	TrainingTaskResult TrainingTaskResultRepository
}

func NewRepositoryContext(user UserRepository, trainingMachine TrainingMachineRepository, trainingDataset TrainingDatasetRepository, trainingTask TrainingTaskRepository, trainingTaskResult TrainingTaskResultRepository) *RepositoryContext {
	return &RepositoryContext{
		User:               user,
		TrainingMachine:    trainingMachine,
		TrainingDataset:    trainingDataset,
		TrainingTask:       trainingTask,
		TrainingTaskResult: trainingTaskResult,
	}
}
