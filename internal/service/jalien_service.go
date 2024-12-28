package service

import (
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/stretchr/testify/mock"
)

type IJAliEnService interface {
	FindAODFiles(path string) ([]jalien.AODFile, error)
	ListAndParseDirectory(path string) (*jalien.DirectoryContents, error)
}

type JAliEnService struct{}

func NewJAliEnService() *JAliEnService {
	return &JAliEnService{}
}

func (s *JAliEnService) FindAODFiles(path string) ([]jalien.AODFile, error) {
	return jalien.FindAODFiles(path)
}

func (s *JAliEnService) ListAndParseDirectory(path string) (*jalien.DirectoryContents, error) {
	return jalien.ListAndParseDirectory(path)
}

type MockJAliEnService struct {
	mock.Mock
}

func NewMockJAliEnService() *MockJAliEnService {
	return &MockJAliEnService{}
}

func (s *MockJAliEnService) FindAODFiles(path string) ([]jalien.AODFile, error) {
	args := s.Called(path)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]jalien.AODFile), args.Error(1)
}

func (s *MockJAliEnService) ListAndParseDirectory(path string) (*jalien.DirectoryContents, error) {
	args := s.Called(path)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*jalien.DirectoryContents), args.Error(1)
}
