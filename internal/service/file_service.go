package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/mytkom/AliceTraINT/internal/db/models"
)

type IFileService interface {
	SaveFile(file multipart.File, handler *multipart.FileHeader) (*models.File, error)
	OpenFile(filepath string) (*os.File, func(*os.File), error)
}

type LocalFileService struct {
	BasePath string
}

func (l *LocalFileService) SaveFile(file multipart.File, handler *multipart.FileHeader) (*models.File, error) {
	defer file.Close()

	fmt.Printf("File name: %+v\n", handler.Filename)
	fmt.Printf("File size: %+v\n", handler.Size)
	fmt.Printf("File header: %+v\n", handler.Header)

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
		Path: fmt.Sprintf("/%s/%s", tempFolderPath, filename),
		Size: uint64(handler.Size),
	}

	return fileModel, nil
}

func (l *LocalFileService) OpenFile(filepath string) (*os.File, func(*os.File), error) {
	fileReader, err := os.Open(fmt.Sprintf(".%s", filepath))
	if err != nil {
		return nil, func(r *os.File) {}, err
	}

	return fileReader, func(r *os.File) {
		r.Close()
	}, nil
}

func NewLocalFileService(basePath string) *LocalFileService {
	return &LocalFileService{
		BasePath: basePath,
	}
}
