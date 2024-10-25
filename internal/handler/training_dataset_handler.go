package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/mytkom/AliceTraINT/internal/middleware"
	"github.com/mytkom/AliceTraINT/internal/utils"
)

type TrainingDatasetHandler struct {
	TrainingDatasetRepo repository.TrainingDatasetRepository
	UserRepo            repository.UserRepository
	Auth                *auth.Auth
	Template            *template.Template
}

type IndexTemplateData struct {
	Title            string
	TrainingDatasets []models.TrainingDataset
}

func (h *TrainingDatasetHandler) Index(w http.ResponseWriter, r *http.Request) {
	sess := h.Auth.GlobalSessions.SessionStart(w, r)
	loggedUserId := sess.Get("loggedUserId")
	if loggedUserId != nil {
		_, err := h.UserRepo.GetByID(loggedUserId.(uint))
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}
	}

	trainingDatasets, err := h.TrainingDatasetRepo.GetAllUser(loggedUserId.(uint))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = h.Template.ExecuteTemplate(w, "training-datasets_index", IndexTemplateData{
		Title:            "Training Datasets",
		TrainingDatasets: trainingDatasets,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TrainingDatasetHandler) New(w http.ResponseWriter, r *http.Request) {
	err := h.Template.ExecuteTemplate(w, "training-datasets_new", map[string]interface{}{
		"Title": "Create New Training Dataset!",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TrainingDatasetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var trainingDataset models.TrainingDataset
	err := json.NewDecoder(r.Body).Decode(&trainingDataset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	sess := h.Auth.GlobalSessions.SessionStart(w, r)
	loggedUserId := sess.Get("loggedUserId")
	if loggedUserId != nil {
		loggedUser, err := h.UserRepo.GetByID(loggedUserId.(uint))
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}

		trainingDataset.UserId = loggedUser.ID
	}

	err = h.TrainingDatasetRepo.Create(&trainingDataset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *TrainingDatasetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	trainingDatasetId, err := strconv.ParseUint(r.PathValue("id"), 10, 32)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	sess := h.Auth.GlobalSessions.SessionStart(w, r)
	loggedUserId := sess.Get("loggedUserId")
	if loggedUserId != nil {
		_, err := h.UserRepo.GetByID(loggedUserId.(uint))
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}
	}

	err = h.TrainingDatasetRepo.Delete(loggedUserId.(uint), uint(trainingDatasetId))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

type exploreDirectoryTemplateData struct {
	Path      string
	Subdirs   []jalien.Dir
	AODFiles  []jalien.AODFile
	ParentDir string
}

func (h *TrainingDatasetHandler) ExploreDirectory(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	dirContents, err := jalien.ListAndParseDirectory(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	parentDir := "/"
	if path != "/" {
		parentDir = filepath.Dir(strings.TrimSuffix(path, "/"))
		if parentDir != "/" {
			parentDir += "/"
		}
	}

	data := exploreDirectoryTemplateData{
		Path:      path,
		AODFiles:  dirContents.AODFiles,
		Subdirs:   dirContents.Subdirs,
		ParentDir: parentDir,
	}

	err = h.Template.ExecuteTemplate(w, "training-datasets_tree-browser", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type findAODsTemplateData struct {
	AODFiles []jalien.AODFile
}

func (h *TrainingDatasetHandler) FindAods(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	aods, err := jalien.FindAODFiles(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	data := &findAODsTemplateData{
		AODFiles: aods,
	}

	err = h.Template.ExecuteTemplate(w, "training-datasets_file-list", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func InitTrainingDatasetRoutes(mux *http.ServeMux, baseTemplate *template.Template, trainingDatasetRepo repository.TrainingDatasetRepository, userRepo repository.UserRepository, auth *auth.Auth, jalienCacheMinutes uint) {
	prefix := "training-datasets"

	tjh := &TrainingDatasetHandler{
		TrainingDatasetRepo: trainingDatasetRepo,
		UserRepo:            userRepo,
		Auth:                auth,
		Template:            baseTemplate,
	}

	cache := utils.NewCache(time.Duration(jalienCacheMinutes) * time.Minute)

	authMw := middleware.NewAuthMw(auth)
	cacheMw := middleware.NewCacheMw(cache)

	mux.Handle(fmt.Sprintf("GET /%s", prefix), middleware.Chain(
		http.HandlerFunc(tjh.Index),
		authMw,
	))

	mux.Handle(fmt.Sprintf("GET /%s/new", prefix), middleware.Chain(
		http.HandlerFunc(tjh.New),
		authMw,
	))

	mux.Handle(fmt.Sprintf("DELETE /%s/{id}", prefix), middleware.Chain(
		http.HandlerFunc(tjh.Delete),
		authMw,
	))

	mux.Handle(fmt.Sprintf("POST /%s", prefix), middleware.Chain(
		http.HandlerFunc(tjh.Create),
		authMw,
	))

	mux.Handle(fmt.Sprintf("GET /%s/explore-directory", prefix), middleware.Chain(
		http.HandlerFunc(tjh.ExploreDirectory),
		cacheMw,
		authMw,
	))

	mux.Handle(fmt.Sprintf("GET /%s/find-aods", prefix), middleware.Chain(
		http.HandlerFunc(tjh.FindAods),
		cacheMw,
		authMw,
	))
}
