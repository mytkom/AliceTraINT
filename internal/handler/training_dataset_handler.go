package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/mytkom/AliceTraINT/internal/middleware"
	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/mytkom/AliceTraINT/internal/utils"
)

type TrainingDatasetHandler struct {
	*environment.Env
	Service service.ITrainingDatasetService
}

func NewTrainingDatasetHandler(env *environment.Env, service service.ITrainingDatasetService) *TrainingDatasetHandler {
	return &TrainingDatasetHandler{
		Env:     env,
		Service: service,
	}
}

func (h *TrainingDatasetHandler) Index(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title string
	}

	err := h.ExecuteTemplate(w, "training-datasets_index", TemplateData{
		Title: "Training Datasets",
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
		return
	}
}

func (h *TrainingDatasetHandler) List(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		TrainingDatasets []models.TrainingDataset
	}

	user, ok := middleware.GetLoggedUser(r)
	if !ok || user == nil {
		writeError(w, r, http.StatusUnauthorized, errMsgUserUnauthorized, nil)
		return
	}

	trainingDatasets, err := h.Service.GetAll(user.ID, utils.IsUserScoped(r))
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	err = h.ExecuteTemplate(w, "training-datasets_list", TemplateData{
		TrainingDatasets: trainingDatasets,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
		return
	}
}

func (h *TrainingDatasetHandler) Show(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title           string
		TrainingDataset models.TrainingDataset
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "invalid training dataset id", err)
		return
	}

	trainingDataset, err := h.Service.GetByID(uint(id))
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	err = h.ExecuteTemplate(w, "training-datasets_show", TemplateData{
		Title:           "Training Datasets",
		TrainingDataset: *trainingDataset,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
	}
}

func (h *TrainingDatasetHandler) New(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title string
	}

	err := h.ExecuteTemplate(w, "training-datasets_new", TemplateData{
		Title: "Create New Training Dataset!",
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
		return
	}
}

func (h *TrainingDatasetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var trainingDataset models.TrainingDataset
	err := json.NewDecoder(r.Body).Decode(&trainingDataset)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid request payload", err)
		return
	}

	user, ok := middleware.GetLoggedUser(r)
	if !ok || user == nil {
		writeError(w, r, http.StatusUnauthorized, errMsgUserUnauthorized, nil)
		return
	}
	trainingDataset.UserId = user.ID

	err = h.Service.Create(&trainingDataset)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	utils.HTMXRedirect(w, "/training-datasets")
	w.WriteHeader(http.StatusCreated)
}

func (h *TrainingDatasetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	trainingDatasetId, err := strconv.ParseUint(r.PathValue("id"), 10, 32)

	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid training dataset id", err)
		return
	}

	user, ok := middleware.GetLoggedUser(r)
	if !ok || user == nil {
		writeError(w, r, http.StatusUnauthorized, errMsgUserUnauthorized, nil)
		return
	}

	err = h.Service.Delete(user.ID, uint(trainingDatasetId))
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TrainingDatasetHandler) ExploreDirectory(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Path      string
		Subdirs   []jalien.Dir
		AODFiles  []jalien.AODFile
		ParentDir string
	}

	path := r.URL.Query().Get("path")
	dirContents, parentDir, err := h.Service.ExploreDirectory(path)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	err = h.ExecuteTemplate(w, "training-datasets_tree-browser", TemplateData{
		Path:      path,
		AODFiles:  dirContents.AODFiles,
		Subdirs:   dirContents.Subdirs,
		ParentDir: parentDir,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
		return
	}
}

func (h *TrainingDatasetHandler) FindAods(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		AODFiles []jalien.AODFile
	}

	path := r.URL.Query().Get("path")
	aods, err := h.Service.FindAods(path)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	err = h.ExecuteTemplate(w, "training-datasets_file-list", TemplateData{
		AODFiles: aods,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
		return
	}
}

func (h *TrainingDatasetHandler) SelectRandomAods(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		AODFiles []jalien.AODFile
	}

	path := r.URL.Query().Get("path")

	runCountStr := r.URL.Query().Get("runCount")
	if runCountStr == "" {
		runCountStr = "10"
	}
	runCount, err := strconv.Atoi(runCountStr)
	if err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "invalid runCount", err)
		return
	}

	filesPerRunStr := r.URL.Query().Get("filesPerRun")
	if filesPerRunStr == "" {
		filesPerRunStr = "5"
	}
	filesPerRun, err := strconv.Atoi(filesPerRunStr)
	if err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "invalid filesPerRun", err)
		return
	}

	minSizeMBStr := r.URL.Query().Get("minFileSizeMB")
	if minSizeMBStr == "" {
		minSizeMBStr = "2048"
	}
	minSizeMB, err := strconv.ParseFloat(minSizeMBStr, 64)
	if err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "invalid minFileSizeMB", err)
		return
	}
	minSizeBytes := uint64(minSizeMB * 1024 * 1024)

	aods, err := h.Service.SelectRandomAODSubset(path, runCount, filesPerRun, minSizeBytes)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	err = h.ExecuteTemplate(w, "training-datasets_file-list", TemplateData{
		AODFiles: aods,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
		return
	}
}

func InitTrainingDatasetRoutes(mux *http.ServeMux, env *environment.Env, jalien service.IJAliEnService) {
	prefix := "training-datasets"

	tjh := &TrainingDatasetHandler{
		Env:     env,
		Service: service.NewTrainingDatasetService(env.RepositoryContext, jalien),
	}

	cache := utils.NewCache(time.Duration(tjh.JalienCacheMinutes) * time.Minute)

	authMw := middleware.NewAuthMw(tjh.IAuthService, true)
	cacheMw := middleware.NewCacheMw(cache)
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

	mux.Handle(fmt.Sprintf("DELETE /%s/{id}", prefix), middleware.Chain(
		http.HandlerFunc(tjh.Delete),
		validateHtmxMw,
		authMw,
	))

	mux.Handle(fmt.Sprintf("POST /%s", prefix), middleware.Chain(
		http.HandlerFunc(tjh.Create),
		validateHtmxMw,
		authMw,
	))

	mux.Handle(fmt.Sprintf("GET /%s/explore-directory", prefix), middleware.Chain(
		http.HandlerFunc(tjh.ExploreDirectory),
		cacheMw,
		validateHtmxMw,
		authMw,
	))

	mux.Handle(fmt.Sprintf("GET /%s/find-aods", prefix), middleware.Chain(
		http.HandlerFunc(tjh.FindAods),
		cacheMw,
		validateHtmxMw,
		authMw,
	))

	mux.Handle(fmt.Sprintf("GET /%s/select-random-aods", prefix), middleware.Chain(
		http.HandlerFunc(tjh.SelectRandomAods),
		validateHtmxMw,
		authMw,
	))
}
