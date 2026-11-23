package service_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/stretchr/testify/assert"
)

type memFile struct {
	*bytes.Reader
}

func (memFile) Close() error { return nil }

func saveTestFile(t *testing.T, fs *service.LocalFileService, name string, content []byte) string {
	t.Helper()

	file := memFile{bytes.NewReader(content)}
	header := &multipart.FileHeader{Filename: name, Size: int64(len(content))}

	fileModel, err := fs.SaveFile(file, header)
	assert.NoError(t, err)

	return fileModel.Path
}

func TestLocalFileService_SaveFile_StoresURLPath(t *testing.T) {
	for _, tc := range []struct {
		name     string
		basePath func(t *testing.T) string
	}{
		{name: "absolute base path", basePath: func(t *testing.T) string { return t.TempDir() }},
		{name: "relative base path", basePath: func(t *testing.T) string {
			t.Chdir(t.TempDir())
			return "data"
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			basePath := tc.basePath(t)
			fs := service.NewLocalFileService(basePath)
			content := []byte("loss: 0.1")

			// Act
			path := saveTestFile(t, fs, "train.log", content)

			// Assert
			assert.Regexp(t, `^/data/\d{4}-\d{2}-\d{2}/upload-train-\d+\.log$`, path)

			stored, err := os.ReadFile(filepath.Join(basePath, strings.TrimPrefix(path, service.DataURLPrefix)))
			assert.NoError(t, err)
			assert.Equal(t, content, stored)
		})
	}
}

func TestLocalFileService_OpenFile_ReadsSavedFile(t *testing.T) {
	// Arrange
	fs := service.NewLocalFileService(t.TempDir())
	content := []byte("onnx model bytes")
	path := saveTestFile(t, fs, "model.onnx", content)

	// Act
	reader, closeFile, err := fs.OpenFile(path)

	// Assert
	assert.NoError(t, err)
	defer closeFile(reader)
	read, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, content, read)
}

func TestLocalFileService_OpenFile_CannotEscapeBasePath(t *testing.T) {
	// Arrange
	root := t.TempDir()
	basePath := filepath.Join(root, "data")
	assert.NoError(t, os.MkdirAll(basePath, os.ModePerm))
	assert.NoError(t, os.WriteFile(filepath.Join(root, "secret.txt"), []byte("secret"), 0o600))
	fs := service.NewLocalFileService(basePath)

	// Act
	_, _, err := fs.OpenFile(service.DataURLPrefix + "../secret.txt")

	// Assert
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestLocalFileService_SavedPathIsServedFromBasePath(t *testing.T) {
	// Arrange
	basePath := t.TempDir()
	fs := service.NewLocalFileService(basePath)
	content := []byte("image bytes")
	path := saveTestFile(t, fs, "plot.png", content)
	// same handler setup as the data route in internal/router.go
	handler := http.StripPrefix(service.DataURLPrefix, http.FileServer(http.Dir(basePath)))

	// Act
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, content, rr.Body.Bytes())
}
