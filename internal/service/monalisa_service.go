package service

import (
	"sync"

	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/mytkom/AliceTraINT/internal/monalisa"
	"github.com/stretchr/testify/mock"
)

type IMonalisaService interface {
	GetRunList() (*monalisa.HyperloopRunList, error)
}

type MonalisaService struct {
	client        *monalisa.MonalisaClient
	mu            sync.RWMutex
	cachedRunList *monalisa.HyperloopRunList
}

func NewMonalisaService(env *environment.Env) *MonalisaService {
	return &MonalisaService{
		client: monalisa.NewMonalisaClient(env.KeyPath, env.CertPath, env.MonalisaBaseURL),
	}
}

func (s *MonalisaService) GetRunList() (*monalisa.HyperloopRunList, error) {
	s.mu.RLock()
	if s.cachedRunList != nil {
		runList := s.cachedRunList
		s.mu.RUnlock()
		return runList, nil
	}
	s.mu.RUnlock()

	runList, err := s.client.GetRunList()
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.cachedRunList = runList
	s.mu.Unlock()

	return runList, nil
}

type MockMonalisaService struct {
	mock.Mock
}

func NewMockMonalisaService() *MockMonalisaService {
	return &MockMonalisaService{}
}

func (s *MockMonalisaService) GetRunList() (*monalisa.HyperloopRunList, error) {
	args := s.Called()

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*monalisa.HyperloopRunList), args.Error(1)
}
