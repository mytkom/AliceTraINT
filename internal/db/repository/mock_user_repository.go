package repository

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetAllUsers() ([]models.User, error) {
	args := m.Called()
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(id int) (*models.User, error) {
	args := m.Called()
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByCernPersonId(cern_person_id string) (*models.User, error) {
	args := m.Called()
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(updatedUser *models.User) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(id int) error {
	args := m.Called()
	return args.Error(0)
}
