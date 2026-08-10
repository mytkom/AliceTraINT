package service

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"

	"github.com/mytkom/AliceTraINT/internal/applog"
	"github.com/mytkom/AliceTraINT/internal/ccdb"
	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/mytkom/AliceTraINT/internal/gridtls"
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
	rootCAs      *x509.CertPool
}

func NewCCDBService(env *environment.Env) *CCDBService {
	cert, err := tls.LoadX509KeyPair(env.CertPath, env.KeyPath)
	if err != nil {
		log.Fatalf("cannot create CCDBService: %s", err.Error())
	}

	rootCAs, err := gridtls.LoadCertPoolWithOptionalFile(env.CCDBCACertPath)
	if err != nil {
		log.Printf("CCDB: using system roots only (%v)", err)
	}

	return &CCDBService{
		baseURL:      env.CCDBBaseURL,
		uploadSubdir: env.CCDBUploadSubdir,
		cert:         cert,
		rootCAs:      rootCAs,
	}
}

func (s *CCDBService) tlsConfig() *tls.Config {
	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{s.cert},
		RootCAs:      s.rootCAs,
	}
}

func (s *CCDBService) GetRunInformation(runNumber uint64) (*ccdb.RunInformation, error) {
	return applog.ExternalCallResult("CCDB", "GetRunInformation", fmt.Sprintf("run=%d", runNumber), func() (*ccdb.RunInformation, error) {
		return ccdb.GetRunInformation(s.baseURL, runNumber, s.tlsConfig())
	})
}

func (s *CCDBService) UploadFile(sor, eor uint64, filename string, file io.Reader) error {
	return applog.ExternalCall("CCDB", "UploadFile", fmt.Sprintf("file=%s sor=%d eor=%d", filename, sor, eor), func() error {
		return ccdb.UploadFile(
			fmt.Sprintf("%s/%s", s.baseURL, s.uploadSubdir),
			s.tlsConfig(),
			sor,
			eor,
			filename,
			file,
		)
	})
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
