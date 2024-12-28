package service

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type NNExpectedResults struct {
	Onnx map[string]string `json:"onnx"`
}

type NNConfigField struct {
	FullName     string      `json:"full_name"`
	Type         string      `json:"type"`
	DefaultValue interface{} `json:"default_value"`
	Min          interface{} `json:"min"`
	Max          interface{} `json:"max"`
	Step         interface{} `json:"step"`
	Description  string      `json:"description"`
}

type NNFieldConfigs map[string]NNConfigField

type NNArchSpec struct {
	FieldConfigs    NNFieldConfigs    `json:"field_configs"`
	ExpectedResults NNExpectedResults `json:"expected_results"`
}

func loadSpec(filename string) (*NNArchSpec, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck
	defer file.Close()

	var arch NNArchSpec
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &arch)
	if err != nil {
		return nil, err
	}

	return &arch, nil
}

type INNArchService interface {
	GetFieldConfigs() NNFieldConfigs
	GetExpectedResults() NNExpectedResults
}

type NNArchService struct {
	*NNArchSpec
}

func NewNNArchService(filename string) *NNArchService {
	spec, err := loadSpec(filename)
	if err != nil {
		log.Fatal(err.Error())
		return nil
	}

	return &NNArchService{
		NNArchSpec: spec,
	}
}

func (s *NNArchService) GetFieldConfigs() NNFieldConfigs {
	return s.FieldConfigs
}

func (s *NNArchService) GetExpectedResults() NNExpectedResults {
	return s.ExpectedResults
}

type NNArchServiceInMemory struct {
	*NNArchSpec
}

func NewNNArchServiceInMemory(fieldConfigs *NNFieldConfigs, expectedResults *NNExpectedResults) *NNArchServiceInMemory {
	return &NNArchServiceInMemory{
		NNArchSpec: &NNArchSpec{
			FieldConfigs:    *fieldConfigs,
			ExpectedResults: *expectedResults,
		},
	}
}

func (s *NNArchServiceInMemory) GetFieldConfigs() NNFieldConfigs {
	return s.FieldConfigs
}

func (s *NNArchServiceInMemory) GetExpectedResults() NNExpectedResults {
	return s.ExpectedResults
}
