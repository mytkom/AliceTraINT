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

type NNOption struct {
	Value interface{} `json:"value"`
	Label string      `json:"label"`
}

type NNShowWhen struct {
	Field  string      `json:"field"`
	Equals interface{} `json:"equals"`
}

type NNConfigField struct {
	Name         string      `json:"name"`
	FullName     string      `json:"full_name"`
	Type         string      `json:"type"`
	DefaultValue interface{} `json:"default_value"`
	Min          interface{} `json:"min,omitempty"`
	Max          interface{} `json:"max,omitempty"`
	Step         interface{} `json:"step,omitempty"`
	Options      []NNOption  `json:"options,omitempty"`
	AllowNull    bool        `json:"allow_null,omitempty"`
	Description  string      `json:"description"`
}

type NNFieldGroup struct {
	Name     string          `json:"name"`
	Path     string          `json:"path"`
	FullName string          `json:"full_name"`
	ShowWhen *NNShowWhen     `json:"show_when,omitempty"`
	Fields   []NNConfigField `json:"fields,omitempty"`
	Groups   []NNFieldGroup  `json:"groups,omitempty"`
}

type NNArchSpec struct {
	FieldGroups     []NNFieldGroup    `json:"field_groups"`
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
	GetFieldGroups() []NNFieldGroup
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

func (s *NNArchService) GetFieldGroups() []NNFieldGroup {
	return s.FieldGroups
}

func (s *NNArchService) GetExpectedResults() NNExpectedResults {
	return s.ExpectedResults
}

type NNArchServiceInMemory struct {
	*NNArchSpec
}

func NewNNArchServiceInMemory(fieldGroups []NNFieldGroup, expectedResults *NNExpectedResults) *NNArchServiceInMemory {
	return &NNArchServiceInMemory{
		NNArchSpec: &NNArchSpec{
			FieldGroups:     fieldGroups,
			ExpectedResults: *expectedResults,
		},
	}
}

func (s *NNArchServiceInMemory) GetFieldGroups() []NNFieldGroup {
	return s.FieldGroups
}

func (s *NNArchServiceInMemory) GetExpectedResults() NNExpectedResults {
	return s.ExpectedResults
}
