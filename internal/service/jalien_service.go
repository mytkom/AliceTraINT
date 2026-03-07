package service

import (
	"log"

	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/stretchr/testify/mock"
)

type IJAliEnService interface {
	FindAODFiles(path string) ([]jalien.AODFile, error)
	ListAndParseDirectory(path string) (*jalien.DirectoryContents, error)
}

type JAliEnService struct {
	client *jalien.Client
}

func NewJAliEnService(env *environment.Env) *JAliEnService {
	cfg := env.Config

	client, err := jalien.NewClient(cfg.JalienHost, cfg.JalienPort, cfg.CertPath, cfg.KeyPath, cfg.JalienCertCADir, cfg.JalienTimeoutSeconds)
	if err != nil {
		log.Fatalf("cannot create JAliEnService: %s", err.Error())
	}

	return &JAliEnService{
		client: client,
	}
}

func (s *JAliEnService) FindAODFiles(path string) ([]jalien.AODFile, error) {
	return s.client.FindAODFiles(path)
}

func (s *JAliEnService) ListAndParseDirectory(path string) (*jalien.DirectoryContents, error) {
	return s.client.ListAndParseDirectory(path)
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
