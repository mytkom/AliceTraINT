package models

import (
	"time"

	"github.com/mytkom/AliceTraINT/internal/jalien"
	"gorm.io/gorm"
)

type TrainingDataset struct {
	gorm.Model
	Name                    string           `gorm:"type:varchar(255);not null;uniqueIndex"`
	AODFiles                []jalien.AODFile `gorm:"serializer:json"`
	IsONe                   bool
	AnchorProdTimestampFrom time.Time
	AnchorProdTimestampTo   time.Time
	UserId                  uint
	User                    User
}
