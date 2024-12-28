package repository

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByCernPersonId(cern_person_id string) (*models.User, error)
	GetAll() ([]models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByCernPersonId(cern_person_id string) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "cern_person_id", cern_person_id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetAll() ([]models.User, error) {
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

type MockUserRepository struct {
	mock.Mock
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{}
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
