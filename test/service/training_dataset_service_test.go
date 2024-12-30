package service_test

import (
	"testing"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/stretchr/testify/assert"
)

type trainingDatasetServiceTestUtils struct {
	TDRepo        *repository.MockTrainingDatasetRepository
	JAliEnService *service.MockJAliEnService
}

func newTrainingDatasetService() (*service.TrainingDatasetService, *trainingDatasetServiceTestUtils) {
	tdRepo := repository.NewMockTrainingDatasetRepository()
	jalienService := service.NewMockJAliEnService()
	return service.NewTrainingDatasetService(&repository.RepositoryContext{
			TrainingDataset: tdRepo,
		}, jalienService), &trainingDatasetServiceTestUtils{
			TDRepo:        tdRepo,
			JAliEnService: jalienService,
		}
}

func TestTrainingDatasetService_GetAll_Global(t *testing.T) {
	// Arrange
	tdService, ut := newTrainingDatasetService()

	userId := uint(1)
	tds := []models.TrainingDataset{
		{Name: "LHC24b1b", UserId: userId, AODFiles: []jalien.AODFile{}},
		{Name: "LHC24b1b2", UserId: userId, AODFiles: []jalien.AODFile{}},
	}
	ut.TDRepo.On("GetAll").Return(tds, nil)

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
	tdService, ut := newTrainingDatasetService()

	userId := uint(1)
	tds := []models.TrainingDataset{
		{Name: "LHC24b1b", UserId: userId, AODFiles: []jalien.AODFile{}},
		{Name: "LHC24b1b2", UserId: userId, AODFiles: []jalien.AODFile{}},
	}
	ut.TDRepo.On("GetAllUser", userId).Return(tds, nil)

	// Act
	datasets, err := tdService.GetAll(userId, true)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, len(datasets))
	assert.Equal(t, tds[0].Name, datasets[0].Name)
	assert.Equal(t, tds[1].Name, datasets[1].Name)
}
