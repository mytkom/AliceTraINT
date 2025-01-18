package service

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"

	"github.com/mytkom/AliceTraINT/internal/ccdb"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"gorm.io/gorm"
)

type TrainingTaskWithResults struct {
	TrainingTask *models.TrainingTask
	ImageFiles   []models.TrainingTaskResult
	OnnxFiles    []models.TrainingTaskResult
	LogFiles     []models.TrainingTaskResult
}

type TrainingTaskHelpers struct {
	TrainingDatasets []models.TrainingDataset
	FieldConfigs     NNFieldConfigs
}

type ITrainingTaskService interface {
	Create(tm *models.TrainingTask) error
	GetAll(loggedUserId uint, userScoped bool) ([]models.TrainingTask, error)
	GetHelpers(loggedUserId uint) (*TrainingTaskHelpers, error)
	GetByID(id uint) (*TrainingTaskWithResults, error)
	UploadOnnxResults(id uint) error
}

type TrainingTaskService struct {
	*repository.RepositoryContext
	CCDBService   ICCDBService
	JAliEnService IJAliEnService
	FileService   IFileService
	NNArch        INNArchService
	PeriodRegex   *regexp.Regexp
}

func NewTrainingTaskService(repo *repository.RepositoryContext, ccdbService ICCDBService, jalienService IJAliEnService, fileService IFileService, nnArch INNArchService) *TrainingTaskService {
	return &TrainingTaskService{
		RepositoryContext: repo,
		CCDBService:       ccdbService,
		JAliEnService:     jalienService,
		FileService:       fileService,
		NNArch:            nnArch,
		PeriodRegex:       regexp.MustCompile(`(/alice/sim/\d{4}/LHC[a-z0-9A-Z\_].+(/\d+)?)/\d+/AOD/\d+`),
	}
}

var errTaskNotFound = NewErrHandlerNotFound("TrainingTask")

func (s *TrainingTaskService) Create(tt *models.TrainingTask) error {
	// Status must start with Queued
	tt.Status = models.Queued

	err := s.TrainingTask.Create(tt)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return &ErrHandlerValidation{
				Field: "Name",
				Msg:   errMsgNotUnique,
			}
		} else {
			return errInternalServerError
		}
	}

	return nil
}

func (s *TrainingTaskService) GetAll(loggedUserId uint, userScoped bool) ([]models.TrainingTask, error) {
	var trainingTasks []models.TrainingTask
	var err error

	if userScoped {
		trainingTasks, err = s.TrainingTask.GetAllUser(loggedUserId)
		if err != nil {
			return nil, errInternalServerError
		}
	} else {
		trainingTasks, err = s.TrainingTask.GetAll()
		if err != nil {
			return nil, errInternalServerError
		}
	}

	return trainingTasks, nil
}

func (s *TrainingTaskService) GetHelpers(loggedUserId uint) (*TrainingTaskHelpers, error) {
	trainingDatasets, err := s.TrainingDataset.GetAllUser(loggedUserId)
	if err != nil {
		return nil, errInternalServerError
	}

	return &TrainingTaskHelpers{
		TrainingDatasets: trainingDatasets,
		FieldConfigs:     s.NNArch.GetFieldConfigs(),
	}, nil
}

func (s *TrainingTaskService) GetByID(id uint) (*TrainingTaskWithResults, error) {
	trainingTask, err := s.TrainingTask.GetByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errTaskNotFound
		} else {
			return nil, errInternalServerError
		}
	}

	var imageFiles []models.TrainingTaskResult
	if trainingTask.Status >= models.Training {
		imageFiles, err = s.TrainingTaskResult.GetByType(trainingTask.ID, models.Image)
		if err != nil {
			return nil, errInternalServerError
		}
	}

	var onnxFiles []models.TrainingTaskResult
	if trainingTask.Status >= models.Benchmarking {
		onnxFiles, err = s.TrainingTaskResult.GetByType(trainingTask.ID, models.Onnx)
		if err != nil {
			return nil, errInternalServerError
		}
	}

	var logFiles []models.TrainingTaskResult
	if trainingTask.Status != models.Queued {
		logFiles, err = s.TrainingTaskResult.GetByType(trainingTask.ID, models.Log)
		if err != nil {
			return nil, errInternalServerError
		}
	}

	return &TrainingTaskWithResults{
		TrainingTask: trainingTask,
		ImageFiles:   imageFiles,
		OnnxFiles:    onnxFiles,
		LogFiles:     logFiles,
	}, nil
}

func (s *TrainingTaskService) UploadOnnxResults(id uint) error {
	trainingTask, err := s.TrainingTask.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errTaskNotFound
		} else {
			return errInternalServerError
		}
	}

	if trainingTask.Status < models.Completed {
		return &ErrHandlerValidation{
			Field: "Status",
			Msg:   "must be completed or uploaded",
		}
	}

	lhcPeriods, err := s.getLHCPeriods(trainingTask)
	if err != nil {
		return err
	}

	var minSOR, maxEOR uint64
	initialized := false

	for i, period := range lhcPeriods {
		log.Printf("%d: Name=\"%s\" DirPath=\"%s\"", i, period.Name, period.DirPath)

		dirContents, err := s.JAliEnService.ListAndParseDirectory(period.DirPath)
		if err != nil {
			return err
		}

		smallestRun, greatestRun, err := s.findRunNumberRange(dirContents.Subdirs)
		if err != nil {
			return err
		}

		firstRunInfo, lastRunInfo, err := s.getRunInfoRange(smallestRun, greatestRun)
		if err != nil {
			return err
		}

		if !initialized || firstRunInfo.SOR < minSOR {
			minSOR = firstRunInfo.SOR
		}

		if !initialized || lastRunInfo.EOR > maxEOR {
			maxEOR = lastRunInfo.EOR
		}

		initialized = true
	}

	mappedOnnxFiles, err := s.filterOnnxFiles(trainingTask.ID)
	if err != nil {
		return err
	}

	for uploadName, file := range mappedOnnxFiles {
		if err := s.uploadOnnxFile(minSOR, maxEOR, file, uploadName); err != nil {
			return err
		}
	}

	trainingTask.Status = models.Uploaded
	if err := s.TrainingTask.Update(trainingTask); err != nil {
		return errInternalServerError
	}

	return nil
}

type lhcPeriod struct {
	Name    string
	DirPath string
}

func (s *TrainingTaskService) periodPathFromAODPath(aodPath string) (string, error) {
	matches := s.PeriodRegex.FindStringSubmatch(aodPath)

	if len(matches) != 3 {
		return "", errors.New("unexpected AOD path format, cannot correctly match")
	}

	return matches[1], nil
}

func (s *TrainingTaskService) getLHCPeriods(task *models.TrainingTask) ([]lhcPeriod, error) {
	var periods []lhcPeriod
	initialized := false

	for _, aod := range task.TrainingDataset.AODFiles {
		if !slices.ContainsFunc(periods, func(p lhcPeriod) bool {
			return p.Name == aod.LHCPeriod
		}) {
			periodPath, err := s.periodPathFromAODPath(aod.Path)
			if err != nil {
				return nil, err
			}

			periods = append(periods, lhcPeriod{
				Name:    aod.LHCPeriod,
				DirPath: periodPath,
			})
		}
		initialized = true
	}

	if !initialized {
		return nil, errors.New("unexpected behaviour: empty training dataset")
	}

	return periods, nil
}

func (s *TrainingTaskService) findRunNumberRange(subdirs []jalien.Dir) (uint64, uint64, error) {
	var smallestRun, greatestRun uint64
	initialized := false

	for _, dir := range subdirs {
		runNumber, err := strconv.ParseUint(dir.Name, 10, 64)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		if !initialized || runNumber < smallestRun {
			smallestRun = runNumber
		}
		if !initialized || runNumber > greatestRun {
			greatestRun = runNumber
		}
		initialized = true
	}

	if !initialized {
		return 0, 0, errors.New("unexpected behaviour: empty training dataset")
	}

	return smallestRun, greatestRun, nil
}

func (s *TrainingTaskService) getRunInfoRange(smallestRun, greatestRun uint64) (*ccdb.RunInformation, *ccdb.RunInformation, error) {
	firstRunInfo, err := s.CCDBService.GetRunInformation(smallestRun)
	if err != nil {
		return nil, nil, handleCCDBError(err)
	}

	lastRunInfo, err := s.CCDBService.GetRunInformation(greatestRun)
	if err != nil {
		return nil, nil, handleCCDBError(err)
	}

	log.Printf("From run %d, SOR %d", firstRunInfo.RunNumber, firstRunInfo.SOR)
	log.Printf("to run %d, EOR %d", lastRunInfo.RunNumber, lastRunInfo.EOR)

	return firstRunInfo, lastRunInfo, nil
}

func (s *TrainingTaskService) filterOnnxFiles(ttId uint) (map[string]*models.TrainingTaskResult, error) {
	expectedOnnxFilenames := s.NNArch.GetExpectedResults().Onnx
	mappedResults := make(map[string]*models.TrainingTaskResult, len(expectedOnnxFilenames))
	onnxFiles, err := s.TrainingTaskResult.GetByType(ttId, models.Onnx)
	if err != nil {
		return nil, err
	}

	for localName, expectedName := range expectedOnnxFilenames {
		found := false
		for _, file := range onnxFiles {
			if file.Name == localName {
				mappedResults[expectedName] = &file
				found = true
				break
			}
		}

		if !found {
			log.Printf("expected file not present: %s", localName)
			return nil, NewErrHandlerNotFound(fmt.Sprintf("TrainingTask's result file: %s", localName))
		}
	}

	return mappedResults, nil
}

func (s *TrainingTaskService) uploadOnnxFile(sor, eor uint64, onnxFile *models.TrainingTaskResult, uploadFilename string) error {
	f, closeFile, err := s.FileService.OpenFile(onnxFile.File.Path)
	if err != nil {
		return err
	}
	defer closeFile(f)

	if err := s.CCDBService.UploadFile(sor, eor, uploadFilename, f); err != nil {
		return handleCCDBError(err)
	}

	return nil
}
