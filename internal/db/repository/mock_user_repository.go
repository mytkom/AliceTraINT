package repository

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetAll() ([]models.User, error) {
	args := m.Called()
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(id uint) (*models.User, error) {
	args := m.Called()
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByCernPersonId(cern_person_id string) (*models.User, error) {
	args := m.Called()
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(updatedUser *models.User) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called()
	return args.Error(0)
}
