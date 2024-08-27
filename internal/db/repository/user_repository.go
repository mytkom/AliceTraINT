package repository

import (
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByID(id int) (*models.User, error)
	GetUserByCernPersonId(cern_person_id string) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id int) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetUserByID(id int) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByCernPersonId(cern_person_id string) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "cern_person_id", cern_person_id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetAllUsers() ([]models.User, error) {
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) DeleteUser(id int) error {
	return r.db.Delete(&models.User{}, id).Error
}
