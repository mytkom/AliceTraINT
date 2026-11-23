package service

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/mock"
)

// DataURLPrefix is the URL path under which stored files are served.
// File paths saved in the database are URLs starting with this prefix,
// independently of the directory the files are stored in.
const DataURLPrefix = "/data/"

type IFileService interface {
	SaveFile(file multipart.File, handler *multipart.FileHeader) (*models.File, error)
	OpenFile(filepath string) (io.ReadCloser, func(io.ReadCloser), error)
}

type LocalFileService struct {
	BasePath string
}

func NewLocalFileService(basePath string) *LocalFileService {
	return &LocalFileService{
		BasePath: basePath,
	}
}

func (l *LocalFileService) SaveFile(file multipart.File, handler *multipart.FileHeader) (*models.File, error) {
	//nolint:errcheck
	defer file.Close()

	log.Printf("File name: %+v\n", handler.Filename)
	log.Printf("File size: %+v\n", handler.Size)
	log.Printf("File header: %+v\n", handler.Header)

	today := time.Now().Format("2006-01-02")
	tempFolderPath := filepath.Join(l.BasePath, today)
	if err := os.MkdirAll(tempFolderPath, os.ModePerm); err != nil {
		return nil, err
	}

	ext := filepath.Ext(handler.Filename)
	tempFileName := fmt.Sprintf("upload-%s-*%s", handler.Filename[:len(handler.Filename)-len(ext)], ext)

	tempFile, err := os.CreateTemp(tempFolderPath, tempFileName)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck
	defer tempFile.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	_, err = tempFile.Write(fileBytes)
	if err != nil {
		return nil, err
	}

	filename := filepath.Base(tempFile.Name())
	fileModel := &models.File{
		Name: handler.Filename,
		Path: DataURLPrefix + today + "/" + filename,
		Size: uint64(handler.Size),
	}

	return fileModel, nil
}

// OpenFile opens a file by its stored URL path (see DataURLPrefix).
func (l *LocalFileService) OpenFile(p string) (io.ReadCloser, func(io.ReadCloser), error) {
	// Clean as an absolute path, so that ".." cannot escape BasePath
	rel := filepath.Clean("/" + strings.TrimPrefix(p, DataURLPrefix))
	fileReader, err := os.Open(filepath.Join(l.BasePath, rel))
	if err != nil {
		return nil, func(r io.ReadCloser) {}, err
	}

	return fileReader, func(r io.ReadCloser) {
		//nolint:errcheck
		r.Close()
	}, nil
}

type MockFileService struct {
	mock.Mock
}

func NewMockFileService() *MockFileService {
	return &MockFileService{}
}

func (m *MockFileService) SaveFile(file multipart.File, handler *multipart.FileHeader) (*models.File, error) {
	args := m.Called(file, handler)
	if args.Get(0) != nil {
		return args.Get(0).(*models.File), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockFileService) OpenFile(filepath string) (io.ReadCloser, func(io.ReadCloser), error) {
	args := m.Called(filepath)
	if args.Get(0) != nil {
		return args.Get(0).(io.ReadCloser), args.Get(1).(func(io.ReadCloser)), args.Error(2)
	}
	return nil, nil, args.Error(2)
}
