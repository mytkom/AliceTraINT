package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	CernPersonId string `gorm:"type:varchar(255);not null;unique"`
	Username     string `gorm:"type:varchar(255);not null"`
	FirstName    string `gorm:"type:varchar(255);not null"`
	FamilyName   string `gorm:"type:varchar(255);not null"`
	Email        string `gorm:"type:varchar(255);unique;not null"`
}
