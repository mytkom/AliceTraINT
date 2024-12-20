package service

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"

	"github.com/mytkom/AliceTraINT/internal/ccdb"
	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/stretchr/testify/mock"
)

type ICCDBService interface {
	GetRunInformation(runNumber uint64) (*ccdb.RunInformation, error)
	UploadFile(sor, eor uint64, filename string, file io.Reader) error
}

type CCDBService struct {
	baseURL      string
	uploadSubdir string
	cert         tls.Certificate
}

func NewCCDBService(env *environment.Env) *CCDBService {
	cert, err := tls.LoadX509KeyPair(env.CCDBCertPath, env.CCDBKeyPath)
	if err != nil {
		log.Fatalf("cannot create CCDBService: %s", err.Error())
	}

	return &CCDBService{
		baseURL:      env.CCDBBaseURL,
		uploadSubdir: env.CCDBUploadSubdir,
		cert:         cert,
	}
}

func (s *CCDBService) GetRunInformation(runNumber uint64) (*ccdb.RunInformation, error) {
	return ccdb.GetRunInformation(s.baseURL, runNumber)
}

func (s *CCDBService) UploadFile(sor, eor uint64, filename string, file io.Reader) error {
	return ccdb.UploadFile(
		fmt.Sprintf("%s/%s", s.baseURL, s.uploadSubdir),
		&s.cert,
		sor,
		eor,
		filename,
		file,
	)
}

type MockCCDBService struct {
	mock.Mock
}

func NewMockCCDBService() *MockCCDBService {
	return &MockCCDBService{}
}

func (s *MockCCDBService) GetRunInformation(runNumber uint64) (*ccdb.RunInformation, error) {
	args := s.Called(runNumber)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*ccdb.RunInformation), args.Error(1)
}

func (s *MockCCDBService) UploadFile(sor, eor uint64, filename string, file io.Reader) error {
	args := s.Called(sor, eor, filename, file)
	return args.Error(0)
}
