package models

import (
	"database/sql/driver"
	"fmt"

	"gorm.io/gorm"
)

type TrainingTaskStatus uint

const (
	Queued TrainingTaskStatus = iota
	Training
	Benchmarking
	Completed
)

func (s *TrainingTaskStatus) Scan(value interface{}) error {
	val, ok := value.(int64)
	if !ok {
		return fmt.Errorf("failed to scan TrainingTaskStatus")
	}
	*s = TrainingTaskStatus(val)
	return nil
}

func (s TrainingTaskStatus) Value() (driver.Value, error) {
	return int64(s), nil
}

func (s TrainingTaskStatus) String() string {
	switch s {
	case Queued:
		return "Queued"
	case Training:
		return "Training"
	case Benchmarking:
		return "Benchmarking"
	case Completed:
		return "Completed"
	default:
		return "Unknown"
	}
}

type TrainingTask struct {
	gorm.Model
	Name              string             `gorm:"type:varchar(255);not null"`
	Status            TrainingTaskStatus `gorm:"type:smallint"`
	UserId            uint
	User              User
	TrainingDatasetId uint
	TrainingDataset   TrainingDataset
	Configuration     interface{} `gorm:"serializer:json"`
	// TODO: benchmarks' files
}
