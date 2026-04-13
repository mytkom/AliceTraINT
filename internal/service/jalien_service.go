package service

import (
	"encoding/json"
	"log"
	"path"
	"strings"

	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/mytkom/AliceTraINT/internal/utils"
	"github.com/stretchr/testify/mock"
)

type IJAliEnService interface {
	FindAODFiles(path string) ([]jalien.AODFile, error)
	ListAndParseDirectory(path string) (*jalien.DirectoryContents, error)
}

type JAliEnService struct {
	client   *jalien.Client
	aodCache *utils.Cache
}

func NewJAliEnService(env *environment.Env, aodCache *utils.Cache) *JAliEnService {
	cfg := env.Config

	client, err := jalien.NewClient(cfg.JalienHost, cfg.JalienPort, cfg.CertPath, cfg.KeyPath, cfg.JalienCertCADir, cfg.JalienTimeoutSeconds)
	if err != nil {
		log.Fatalf("cannot create JAliEnService: %s", err.Error())
	}

	return &JAliEnService{
		client:   client,
		aodCache: aodCache,
	}
}
func findAODFilesCacheKey(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	return "aodfiles:" + path.Clean(p)
}

func (s *JAliEnService) FindAODFiles(p string) ([]jalien.AODFile, error) {
	cacheKey := findAODFilesCacheKey(p)
	if cacheKey == "" || s.aodCache == nil {
		return s.client.FindAODFiles(p)
	}

	if raw, ok := s.aodCache.Get(cacheKey); ok {
		var files []jalien.AODFile
		if err := json.Unmarshal([]byte(raw), &files); err == nil {
			return files, nil
		}
	}
	
	files, err := s.client.FindAODFiles(p)
	if err != nil {
		return nil, err
	}

	if b, err := json.Marshal(files); err == nil {
		s.aodCache.Set(cacheKey, string(b))
	}
	return files, nil
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
