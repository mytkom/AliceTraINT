package service_test

import (
	"testing"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/stretchr/testify/assert"
)

// Helper to create a new TrainingDatasetService instance
func newTrainingDatasetService() (*repository.MockTrainingDatasetRepository, *service.MockJAliEnService, *service.TrainingDatasetService) {
	tdRepo := repository.NewMockTrainingDatasetRepository()
	jalienService := service.NewMockJAliEnService()
	tdService := service.NewTrainingDatasetService(&repository.RepositoryContext{
		TrainingDataset: tdRepo,
	}, jalienService)
	return tdRepo, jalienService, tdService
}

func TestTrainingDatasetService_GetAll_Global(t *testing.T) {
	// Arrange
	tdRepo, _, tdService := newTrainingDatasetService()

	userId := uint(1)
	tds := []models.TrainingDataset{
		{Name: "LHC24b1b", UserId: userId, AODFiles: []jalien.AODFile{}},
		{Name: "LHC24b1b2", UserId: userId, AODFiles: []jalien.AODFile{}},
	}
	tdRepo.On("GetAll").Return(tds, nil)

	// Act
	datasets, err := tdService.GetAll(userId, false)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, len(datasets))
	assert.Equal(t, tds[0].Name, datasets[0].Name)
	assert.Equal(t, tds[1].Name, datasets[1].Name)
}

func TestTrainingDatasetService_GetAll_UserScoped(t *testing.T) {
	// Arrange
	tdRepo, _, tdService := newTrainingDatasetService()

	userId := uint(1)
	tds := []models.TrainingDataset{
		{Name: "LHC24b1b", UserId: userId, AODFiles: []jalien.AODFile{}},
		{Name: "LHC24b1b2", UserId: userId, AODFiles: []jalien.AODFile{}},
	}
	tdRepo.On("GetAllUser").Return(tds, nil)

	// Act
	datasets, err := tdService.GetAll(userId, true)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, len(datasets))
	assert.Equal(t, tds[0].Name, datasets[0].Name)
	assert.Equal(t, tds[1].Name, datasets[1].Name)
}
