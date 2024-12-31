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

type trainingMachineServiceTestUtils struct {
	TMRepo *repository.MockTrainingMachineRepository
	Hasher *service.MockHasher
}

func newTrainingMachineService() (*service.TrainingMachineService, *trainingMachineServiceTestUtils) {
	tmRepo := repository.NewMockTrainingMachineRepository()
	hasher := service.NewMockHasher()

	return service.NewTrainingMachineService(&repository.RepositoryContext{
			TrainingMachine: tmRepo,
		}, hasher), &trainingMachineServiceTestUtils{
			TMRepo: tmRepo,
			Hasher: hasher,
		}
}

func TestTrainingMachineService_GetAll_Global(t *testing.T) {
	// Arrange
	tmService, ut := newTrainingMachineService()
	userId := uint(1)
	tms := []models.TrainingMachine{
		{Name: "awm1", UserId: userId, SecretKeyHashed: "secret1", LastActivityAt: time.Now()},
		{Name: "awm2", UserId: userId, SecretKeyHashed: "secret2", LastActivityAt: time.Now().Add(5 * time.Hour)},
	}
	ut.TMRepo.On("GetAll").Return(tms, nil)

	// Act
	machines, err := tmService.GetAll(userId, false)

	// Assert
	assert.NoError(t, err)
	ut.TMRepo.AssertCalled(t, "GetAll")
	ut.TMRepo.AssertNotCalled(t, "GetAllUser", mock.Anything)
	assert.Equal(t, 2, len(machines))
	assert.Equal(t, tms[0].Name, machines[0].Name)
	assert.Equal(t, tms[1].Name, machines[1].Name)
}

func TestTrainingMachineService_GetAll_UserScoped(t *testing.T) {
	// Arrange
	tmService, ut := newTrainingMachineService()
	userId := uint(1)
	tms := []models.TrainingMachine{
		{Name: "awm1", UserId: userId, SecretKeyHashed: "secret1", LastActivityAt: time.Now()},
		{Name: "awm2", UserId: userId, SecretKeyHashed: "secret2", LastActivityAt: time.Now().Add(5 * time.Hour)},
	}
	ut.TMRepo.On("GetAllUser", userId).Return(tms, nil)

	// Act
	machines, err := tmService.GetAll(userId, true)

	// Assert
	assert.NoError(t, err)
	ut.TMRepo.AssertNotCalled(t, "GetAll")
	ut.TMRepo.AssertCalled(t, "GetAllUser", userId)
	assert.Equal(t, 2, len(machines))
	assert.Equal(t, tms[0].Name, machines[0].Name)
	assert.Equal(t, tms[1].Name, machines[1].Name)
}

func TestTrainingMachineService_Create(t *testing.T) {
	// Arrange
	tmService, ut := newTrainingMachineService()
	userId := uint(1)
	tm := models.TrainingMachine{
		Name:           "awm1",
		UserId:         userId,
		LastActivityAt: time.Now(),
	}
	ut.TMRepo.On("Create", mock.Anything).Return(nil)
	ut.Hasher.On("GenerateKey").Return("secret", nil)
	ut.Hasher.On("HashKey", "secret").Return("secretHashed", nil)

	// Act
	secretKey, err := tmService.Create(&tm)

	// Assert
	assert.NoError(t, err)
	ut.Hasher.AssertCalled(t, "GenerateKey")
	ut.Hasher.AssertCalled(t, "HashKey", "secret")
	ut.TMRepo.AssertCalled(t, "Create", &tm)
	assert.Equal(t, "secret", secretKey)
	assert.Equal(t, "secretHashed", tm.SecretKeyHashed)
}
