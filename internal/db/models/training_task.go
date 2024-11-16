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
	Failed
)

func (s *TrainingTaskStatus) Scan(value interface{}) error {
	val, ok := value.(int64)
	if !ok {
		return fmt.Errorf("failed to scan status")
	}
	*s = TrainingTaskStatus(val)
	return nil
}

func (s TrainingTaskStatus) Value() (driver.Value, error) {
	if s < Queued || s > Completed {
		return nil, fmt.Errorf("bad status")
	}

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
	case Failed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// returns tailwind color suffix and this classes should be included in tailwind's safelist
func (s TrainingTaskStatus) Color() string {
	switch s {
	case Queued:
		return "emerald-600"
	case Training:
		return "yellow-200"
	case Benchmarking:
		return "yellow-600"
	case Completed:
		return "green-600"
	default:
		return "gray-400"
	}
}

type TrainingTask struct {
	gorm.Model
	Name              string             `gorm:"type:varchar(255);not null;uniqueIndex"`
	Status            TrainingTaskStatus `gorm:"type:smallint"`
	UserId            uint
	User              User
	TrainingDatasetId uint
	TrainingDataset   TrainingDataset
	TrainingMachineId *uint
	TrainingMachine   TrainingMachine
	Configuration     interface{} `gorm:"serializer:json"`
}
