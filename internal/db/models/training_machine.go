package models

import (
	"time"

	"gorm.io/gorm"
)

type TrainingMachine struct {
	gorm.Model
	Name            string `gorm:"type:varchar(255);not null"`
	LastActivityAt  time.Time
	SecretKeyHashed string
	UserId          uint
	User            User
}
