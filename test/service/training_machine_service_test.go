package service_test

import (
	"testing"
	"time"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTrainingMachineService_GetAll_Global(t *testing.T) {
	tmRepo := repository.NewMockTrainingMachineRepository()
	hasher := service.NewMockHasher()
	tdService := service.NewTrainingMachineService(&repository.RepositoryContext{
		TrainingMachine: tmRepo,
	}, hasher)

	userId := uint(1)
	tms := []models.TrainingMachine{
		{Name: "awm1", UserId: userId, SecretKeyHashed: "secret1", LastActivityAt: time.Now()},
		{Name: "awm2", UserId: userId, SecretKeyHashed: "secret2", LastActivityAt: time.Now().Add(5 * time.Hour)},
	}
	tmRepo.On("GetAll").Return(tms, nil)

	machines, err := tdService.GetAll(userId, false)
	assert.NoError(t, err)
	tmRepo.AssertCalled(t, "GetAll")
	tmRepo.AssertNotCalled(t, "GetAllUser", mock.Anything)
	assert.Equal(t, 2, len(machines))
	assert.Equal(t, tms[0].Name, machines[0].Name)
	assert.Equal(t, tms[1].Name, machines[1].Name)
}

func TestTrainingMachineService_GetAll_UserScoped(t *testing.T) {
	tmRepo := repository.NewMockTrainingMachineRepository()
	hasher := service.NewMockHasher()
	tdService := service.NewTrainingMachineService(&repository.RepositoryContext{
		TrainingMachine: tmRepo,
	}, hasher)

	userId := uint(1)
	tms := []models.TrainingMachine{
		{Name: "awm1", UserId: userId, SecretKeyHashed: "secret1", LastActivityAt: time.Now()},
		{Name: "awm2", UserId: userId, SecretKeyHashed: "secret2", LastActivityAt: time.Now().Add(5 * time.Hour)},
	}
	tmRepo.On("GetAllUser", userId).Return(tms, nil)

	machines, err := tdService.GetAll(userId, true)
	assert.NoError(t, err)
	tmRepo.AssertNotCalled(t, "GetAll")
	tmRepo.AssertCalled(t, "GetAllUser", userId)
	assert.Equal(t, 2, len(machines))
	assert.Equal(t, tms[0].Name, machines[0].Name)
	assert.Equal(t, tms[1].Name, machines[1].Name)
}

func TestTrainingMachineService_Create(t *testing.T) {
	tmRepo := repository.NewMockTrainingMachineRepository()
	hasher := service.NewMockHasher()
	tmService := service.NewTrainingMachineService(&repository.RepositoryContext{
		TrainingMachine: tmRepo,
	}, hasher)

	userId := uint(1)
	tm := models.TrainingMachine{
		Name:           "awm1",
		UserId:         userId,
		LastActivityAt: time.Now(),
	}
	tmRepo.On("Create", mock.Anything).Return(nil)
	hasher.On("GenerateKey").Return("secret", nil)
	hasher.On("HashKey", "secret").Return("secretHashed", nil)

	secretKey, err := tmService.Create(&tm)
	assert.NoError(t, err)

	hasher.AssertCalled(t, "GenerateKey")
	hasher.AssertCalled(t, "HashKey", "secret")
	tmRepo.AssertCalled(t, "Create", &tm)
	assert.Equal(t, "secret", secretKey)
	assert.Equal(t, "secretHashed", tm.SecretKeyHashed)
}
