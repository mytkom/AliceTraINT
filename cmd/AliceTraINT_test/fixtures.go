package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/mytkom/AliceTraINT/internal/ccdb"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/mytkom/AliceTraINT/internal/monalisa"
	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func fixtureServices() (*service.MockJAliEnService, *service.MockCCDBService, *service.MockMonalisaService) {
	grid := service.NewMockJAliEnService()
	grid.On("FindAODFiles", "/alice/sim/2024/LHC24f3/0/523397").Return(aods(8, 8, 523397, "2024", "LHC24f3"), nil)
	grid.On("FindAODFiles", "/alice/sim/2024/LHC24f3/0").Return(aods(10581, 55, 523397, "2024", "LHC24f3"), nil)
	// The first six runs each have 1,000 files. All remaining runs have
	// fewer than 1,000, so the 6 × 1,000 subset always picks the same files.
	large := aods(6000, 1000, 600000, "2025", "LHC25h3c")
	large = append(large, aods(7355, 900, 600006, "2025", "LHC25h3c")...)
	grid.On("FindAODFiles", "/alice/sim/2025/LHC25h3c/0").Return(large, nil)
	grid.On("FindAODFiles", mock.Anything).Return([]jalien.AODFile(nil), errors.New("no fixture for path"))
	grid.On("ListAndParseDirectory", "/alice/sim/").Return(&jalien.DirectoryContents{Subdirs: []jalien.Dir{
		{Name: "Small fixture run", Path: "/alice/sim/2024/LHC24f3/0/523397"},
		{Name: "10,581 AODs", Path: "/alice/sim/2024/LHC24f3/0"},
		{Name: "13,355 AODs", Path: "/alice/sim/2025/LHC25h3c/0"},
	}}, nil)
	grid.On("ListAndParseDirectory", mock.Anything).Return(&jalien.DirectoryContents{}, nil)

	monitor := service.NewMockMonalisaService()
	monitor.On("GetRunList").Return(&monalisa.HyperloopRunList{
		TagToRuns: map[string][]uint64{"LHC24f3": {523397}, "LHC25h3c": {600000}},
		TagToDesc: map[string]string{"LHC24f3": "Synthetic MC", "LHC25h3c": "Synthetic MC"},
	}, nil)

	conditions := service.NewMockCCDBService()
	for _, run := range []uint64{523397, 600000} {
		conditions.On("GetRunInformation", run).Return(&ccdb.RunInformation{
			RunNumber: run, SOR: 1_730_000_000_000, EOR: 1_730_001_000_000,
		}, nil)
	}
	conditions.On("GetRunInformation", mock.Anything).Return(nil, errors.New("run information unavailable in browser fixtures"))
	conditions.On("UploadFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("uploads disabled in browser fixtures"))
	return grid, conditions, monitor
}

func aods(count, perRun int, firstRun uint64, year, period string) []jalien.AODFile {
	files := make([]jalien.AODFile, count)
	for i := range files {
		run, number := firstRun+uint64(i/perRun), uint64(i%perRun)
		prefix := fmt.Sprintf("/alice/sim/%s/%s/0/%d", year, period, run)
		files[i] = jalien.AODFile{
			Name: "AO2D.root", Path: fmt.Sprintf("%s/AOD/%04d/AO2D.root", prefix, number),
			Size: 3_000_000_000, LHCPeriod: period, RunNumber: run, AODNumber: number,
		}
	}
	return files
}

func seedGalleries(db *gorm.DB, dataDir string) error {
	// Encode once, then serve distinct URLs of equal size to exercise every
	// image request. The fixed indexed noise approximates a 25 KB plot PNG.
	palette := color.Palette{color.White, color.RGBA{R: 60, G: 95, B: 160, A: 255}}
	img := image.NewPaletted(image.Rect(0, 0, 640, 480), palette)
	var state uint32 = 1
	for y := 60; y < 420; y++ {
		for x := 70; x < 570; x++ {
			state = state*1664525 + 1013904223
			img.SetColorIndex(x, y, uint8(state>>31))
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		return err
	}

	for _, count := range []int{180, 738} {
		task := models.TrainingTask{
			Name: fmt.Sprintf("Fixture gallery %d", count), Status: models.Completed,
			UserId: 1, TrainingDatasetId: 1,
		}
		if err := db.Create(&task).Error; err != nil {
			return err
		}
		results := make([]models.TrainingTaskResult, count)
		for i := range results {
			name := fmt.Sprintf("plot-%d-%04d.png", count, i+1)
			if err := os.WriteFile(filepath.Join(dataDir, name), encoded.Bytes(), 0600); err != nil {
				return err
			}
			results[i] = models.TrainingTaskResult{
				Name: name, Type: models.Image, TrainingTaskId: task.ID,
				File: models.File{Name: name, Path: service.DataURLPrefix + name, Size: uint64(encoded.Len())},
			}
		}
		if err := db.CreateInBatches(&results, 100).Error; err != nil {
			return err
		}
	}
	return nil
}
