package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"sort"
	"strconv"

	"github.com/mytkom/AliceTraINT/internal/ccdb"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/mytkom/AliceTraINT/internal/middleware"
	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/mytkom/AliceTraINT/internal/utils"
)

type NNExpectedResults struct {
	Onnx map[string]string `json:"onnx"`
}

type NNConfigField struct {
	FullName     string      `json:"full_name"`
	Type         string      `json:"type"`
	DefaultValue interface{} `json:"default_value"`
	Min          interface{} `json:"min"`
	Max          interface{} `json:"max"`
	Step         interface{} `json:"step"`
	Description  string      `json:"description"`
}

type NNArchSpec struct {
	FieldConfigs    map[string]NNConfigField `json:"field_configs"`
	ExpectedResults NNExpectedResults        `json:"expected_results"`
}

func loadNNArchSpec(filename string) (*NNArchSpec, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	var arch NNArchSpec
	bytes, err := io.ReadAll(file)

	if err != nil {
		return nil, err
	}

	err = file.Close()

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &arch)

	if err != nil {
		return nil, err
	}

	return &arch, nil
}

type TrainingTaskHandler struct {
	*environment.Env
	CCDBApi     *ccdb.CCDBApi
	NNArchSpec  *NNArchSpec
	FileService service.IFileService
}

func (h *TrainingTaskHandler) Index(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title string
	}

	err := h.ExecuteTemplate(w, "training-tasks_index", TemplateData{
		Title: "Training Tasks",
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TrainingTaskHandler) List(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		TrainingTasks []models.TrainingTask
	}

	var trainingTasks []models.TrainingTask
	var err error

	if r.URL.Query().Get("userScoped") == "on" {
		loggedUser, err := getAuthorizedUser(h.Auth, h.User, w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		trainingTasks, err = h.TrainingTask.GetAllUser(loggedUser.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		trainingTasks, err = h.TrainingTask.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = h.ExecuteTemplate(w, "training-tasks_list", TemplateData{
		TrainingTasks: trainingTasks,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TrainingTaskHandler) UploadToCCDB(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid training task id", http.StatusUnprocessableEntity)
		return
	}

	trainingTask, err := h.TrainingTask.GetByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	runs := []uint64{}
	for _, aod := range trainingTask.TrainingDataset.AODFiles {
		if !slices.Contains(runs, aod.RunNumber) {
			runs = append(runs, aod.RunNumber)
		}
	}

	if len(runs) == 0 {
		http.Error(w, "unexpected behaviour: empty training dataset", http.StatusInternalServerError)
		return
	}

	sort.Slice(runs, func(i, j int) bool {
		return runs[i] < runs[j]
	})

	firstRunInfo, err := h.CCDBApi.GetRunInformation(runs[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lastRunInfo, err := h.CCDBApi.GetRunInformation(runs[len(runs)-1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("From run %d, SOR %d", firstRunInfo.RunNumber, firstRunInfo.SOR)
	fmt.Printf("to run %d, EOR %d", lastRunInfo.RunNumber, lastRunInfo.SOR)

	onnxFiles, err := h.TrainingTaskResult.GetByType(trainingTask.ID, models.Onnx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, onnxFile := range onnxFiles {
		file, close, err := h.FileService.OpenFile(onnxFile.File.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer close(file)

		if upload_filename, ok := h.NNArchSpec.ExpectedResults.Onnx[onnxFile.Name]; ok {
			err = ccdb.UploadFile(h.Config, firstRunInfo.SOR, lastRunInfo.EOR, upload_filename, file)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			fmt.Printf("not expected file: %s", onnxFile.Name)
			continue
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TrainingTaskHandler) Show(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title        string
		TrainingTask models.TrainingTask
		ImageFiles   []models.TrainingTaskResult
		OnnxFiles    []models.TrainingTaskResult
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid training task id", http.StatusUnprocessableEntity)
		return
	}

	trainingTask, err := h.TrainingTask.GetByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	imageFiles, err := h.TrainingTaskResult.GetByType(trainingTask.ID, models.Image)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	onnxFiles, err := h.TrainingTaskResult.GetByType(trainingTask.ID, models.Onnx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.ExecuteTemplate(w, "training-tasks_show", TemplateData{
		Title:        "Training Tasks",
		TrainingTask: *trainingTask,
		ImageFiles:   imageFiles,
		OnnxFiles:    onnxFiles,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TrainingTaskHandler) New(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title            string
		TrainingDatasets []models.TrainingDataset
		NNArchSpec       *NNArchSpec
	}

	loggedUser, err := getAuthorizedUser(h.Auth, h.User, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	trainingDatasets, err := h.TrainingDataset.GetAllUser(loggedUser.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.ExecuteTemplate(w, "training-tasks_new", TemplateData{
		Title:            "Create New Training Task!",
		TrainingDatasets: trainingDatasets,
		NNArchSpec:       h.NNArchSpec,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TrainingTaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var trainingTask models.TrainingTask
	err := json.NewDecoder(r.Body).Decode(&trainingTask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	loggedUser, err := getAuthorizedUser(h.Auth, h.User, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	trainingTask.UserId = loggedUser.ID

	err = h.TrainingTask.Create(&trainingTask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.HTMXRedirect(w, "/training-tasks")
	w.WriteHeader(http.StatusCreated)
}

func InitTrainingTaskRoutes(mux *http.ServeMux, env *environment.Env, ccdbApi *ccdb.CCDBApi, fileService service.IFileService) {
	prefix := "training-tasks"

	nnArchSpec, err := loadNNArchSpec("web/nn_architectures/proposed.json")
	if err != nil {
		log.Fatal("cannot load architecture configuration specification file")
	}

	tjh := &TrainingTaskHandler{
		Env:         env,
		CCDBApi:     ccdbApi,
		NNArchSpec:  nnArchSpec,
		FileService: fileService,
	}

	authMw := middleware.NewAuthMw(env.Auth, true)
	validateHtmxMw := middleware.NewValidateHTMXMw()
	blockHtmxMw := middleware.NewBlockHTMXMw()

	mux.Handle(fmt.Sprintf("GET /%s", prefix), middleware.Chain(
		http.HandlerFunc(tjh.Index),
		blockHtmxMw,
		authMw,
	))

	mux.Handle(fmt.Sprintf("GET /%s/{id}", prefix), middleware.Chain(
		http.HandlerFunc(tjh.Show),
		blockHtmxMw,
		authMw,
	))

	mux.Handle(fmt.Sprintf("GET /%s/list", prefix), middleware.Chain(
		http.HandlerFunc(tjh.List),
		validateHtmxMw,
		authMw,
	))

	mux.Handle(fmt.Sprintf("GET /%s/new", prefix), middleware.Chain(
		http.HandlerFunc(tjh.New),
		blockHtmxMw,
		authMw,
	))

	mux.Handle(fmt.Sprintf("POST /%s", prefix), middleware.Chain(
		http.HandlerFunc(tjh.Create),
		validateHtmxMw,
		authMw,
	))

	mux.Handle(fmt.Sprintf("POST /%s/{id}/upload-to-ccdb", prefix), middleware.Chain(
		http.HandlerFunc(tjh.UploadToCCDB),
		validateHtmxMw,
		authMw,
	))
}
