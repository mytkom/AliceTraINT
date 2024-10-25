package models

import (
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"gorm.io/gorm"
)

type TrainingDataset struct {
	gorm.Model
	Name     string           `gorm:"type:varchar(255);not null"`
	AODFiles []jalien.AODFile `gorm:"serializer:json"`
	UserId   uint
	User     User
}
