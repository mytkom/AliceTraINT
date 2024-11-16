package models

import (
	"database/sql/driver"
	"fmt"

	"gorm.io/gorm"
)

type TrainingTaskResultType uint

const (
	Log TrainingTaskResultType = iota
	Image
	Onnx
)

func (s *TrainingTaskResultType) Scan(value interface{}) error {
	val, ok := value.(uint)
	if !ok {
		return fmt.Errorf("failed to scan TrainingTaskStatus")
	}
	*s = TrainingTaskResultType(val)
	return nil
}

func (s TrainingTaskResultType) Value() (driver.Value, error) {
	return uint(s), nil
}

type TrainingTaskResult struct {
	gorm.Model
	Name           string
	Type           TrainingTaskResultType
	Description    string
	File           []byte
	TrainingTaskId uint
	TrainingTask   TrainingTask
}
