package models

import "gorm.io/gorm"

type TrainingTaskStatus string

const (
	Queued       TrainingTaskStatus = "queued"
	Training     TrainingTaskStatus = "training"
	Benchmarking TrainingTaskStatus = "benchmarking"
	Completed    TrainingTaskStatus = "completed"
)

type TrainingTaskConfig struct {
	BatchSize         uint    `json:"bs"`
	MaxEpochs         uint    `json:"max_epochs"`
	DropoutRate       float64 `json:"dropout"`
	Gamma             float64 `json:"gamma"`
	Patience          uint    `json:"patience"`
	PatienceThreshold float64 `json:"patience_threshold"`
	EmbedHidden       uint    `json:"embed_hidden"`
	DModel            uint    `json:"d_model"`
	FFHidden          uint    `json:"ff_hidden"`
	PoolHidden        uint    `json:"pool_hidden"`
	NumHeads          uint    `json:"num_heads"`
	NumBlocks         uint    `json:"num_blocks"`
	StartLearningRate float64 `json:"start_lr"`
}

type TrainingTask struct {
	gorm.Model
	Name              string             `gorm:"type:varchar(255);not null"`
	Status            TrainingTaskStatus `gorm:"type:enum('queued', 'training', 'benchmarking', 'completed')"`
	UserId            uint
	User              User
	TrainingDatasetId uint
	TrainingDataset   TrainingDataset
	Configuration     TrainingTaskConfig `gorm:"serializer:json"`
	// TODO: benchmarks' files
}
