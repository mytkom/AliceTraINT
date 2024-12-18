package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/mytkom/AliceTraINT/internal/middleware"
	"github.com/mytkom/AliceTraINT/internal/utils"
)

type TrainingDatasetHandler struct {
	*environment.Env
}

func (h *TrainingDatasetHandler) Index(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title string
	}

	err := h.ExecuteTemplate(w, "training-datasets_index", TemplateData{
		Title: "Training Datasets",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TrainingDatasetHandler) List(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		TrainingDatasets []models.TrainingDataset
	}

	var trainingDatasets []models.TrainingDataset
	var err error

	if r.URL.Query().Get("userScoped") == "on" {
		loggedUser, err := h.GetAuthorizedUser(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		trainingDatasets, err = h.TrainingDataset.GetAllUser(loggedUser.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		trainingDatasets, err = h.TrainingDataset.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = h.ExecuteTemplate(w, "training-datasets_list", TemplateData{
		TrainingDatasets: trainingDatasets,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "invalid training dataset id", http.StatusUnprocessableEntity)
		return
	}

	trainingDataset, err := h.TrainingDataset.GetByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.ExecuteTemplate(w, "training-datasets_show", TemplateData{
		Title:           "Training Datasets",
		TrainingDataset: *trainingDataset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TrainingDatasetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var trainingDataset models.TrainingDataset
	err := json.NewDecoder(r.Body).Decode(&trainingDataset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	loggedUser, err := h.GetAuthorizedUser(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	trainingDataset.UserId = loggedUser.ID

	err = h.TrainingDataset.Create(&trainingDataset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.HTMXRedirect(w, "/training-datasets")
	w.WriteHeader(http.StatusCreated)
}

func (h *TrainingDatasetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	trainingDatasetId, err := strconv.ParseUint(r.PathValue("id"), 10, 32)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	loggedUser, err := h.GetAuthorizedUser(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = h.TrainingDataset.Delete(loggedUser.ID, uint(trainingDatasetId))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	if path == "" {
		path = "/"
	}

	dirContents, err := jalien.ListAndParseDirectory(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	parentDir := "/"
	if path != "/" {
		parentDir = filepath.Dir(strings.TrimSuffix(path, "/"))
		if parentDir != "/" {
			parentDir += "/"
		}
	}

	err = h.ExecuteTemplate(w, "training-datasets_tree-browser", TemplateData{
		Path:      path,
		AODFiles:  dirContents.AODFiles,
		Subdirs:   dirContents.Subdirs,
		ParentDir: parentDir,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TrainingDatasetHandler) FindAods(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		AODFiles []jalien.AODFile
	}

	path := r.URL.Query().Get("path")

	aods, err := jalien.FindAODFiles(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.ExecuteTemplate(w, "training-datasets_file-list", TemplateData{
		AODFiles: aods,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func InitTrainingDatasetRoutes(mux *http.ServeMux, env *environment.Env) {
	prefix := "training-datasets"

	tjh := &TrainingDatasetHandler{
		Env: env,
	}

	cache := utils.NewCache(time.Duration(tjh.JalienCacheMinutes) * time.Minute)

	authMw := middleware.NewAuthMw(tjh.Auth, true)
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
}
