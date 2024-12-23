package service_test

import (
	"testing"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestTrainingDatasetService_GetAll_Global(t *testing.T) {
	tdRepo := &repository.MockTrainingDatasetRepository{}
	jalienService := service.NewMockJAliEnService()
	tdService := service.NewTrainingDatasetService(&repository.RepositoryContext{
		TrainingDataset: tdRepo,
	}, jalienService)

	userId := uint(1)
	tds := []models.TrainingDataset{
		{Name: "LHC24b1b", UserId: userId, AODFiles: []jalien.AODFile{}},
		{Name: "LHC24b1b2", UserId: userId, AODFiles: []jalien.AODFile{}},
	}
	tdRepo.On("GetAll").Return(tds, nil)

	datasets, err := tdService.GetAll(userId, false)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(datasets))
	assert.Equal(t, tds[0].Name, datasets[0].Name)
	assert.Equal(t, tds[1].Name, datasets[1].Name)
}

func TestTrainingDatasetService_GetAll_UserScoped(t *testing.T) {
	tdRepo := &repository.MockTrainingDatasetRepository{}
	jalienService := service.NewMockJAliEnService()
	tdService := service.NewTrainingDatasetService(&repository.RepositoryContext{
		TrainingDataset: tdRepo,
	}, jalienService)

	userId := uint(1)
	tds := []models.TrainingDataset{
		{Name: "LHC24b1b", UserId: userId, AODFiles: []jalien.AODFile{}},
		{Name: "LHC24b1b2", UserId: userId, AODFiles: []jalien.AODFile{}},
	}
	tdRepo.On("GetAllUser").Return(tds, nil)

	datasets, err := tdService.GetAll(userId, true)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(datasets))
	assert.Equal(t, tds[0].Name, datasets[0].Name)
	assert.Equal(t, tds[1].Name, datasets[1].Name)
}
