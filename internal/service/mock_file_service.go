package service

import (
	"mime/multipart"
	"os"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/stretchr/testify/mock"
)

type MockFileService struct {
	mock.Mock
}

func (m *MockFileService) SaveFile(file multipart.File, handler *multipart.FileHeader) (*models.File, error) {
	args := m.Called(file, handler)
	if args.Get(0) != nil {
		return args.Get(0).(*models.File), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockFileService) OpenFile(filepath string) (*os.File, func(*os.File), error) {
	args := m.Called(filepath)
	if args.Get(0) != nil {
		return args.Get(0).(*os.File), args.Get(1).(func(*os.File)), args.Error(2)
	}
	return nil, nil, args.Error(2)
}
