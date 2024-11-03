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

func (h *TrainingDatasetHandler) Index(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title string
	}

	err := h.Template.ExecuteTemplate(w, "training-datasets_index", TemplateData{
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
		sess := h.Auth.GlobalSessions.SessionStart(w, r)
		loggedUserId := sess.Get("loggedUserId")
		if loggedUserId != nil {
			_, err := h.UserRepo.GetByID(loggedUserId.(uint))
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
		}

		trainingDatasets, err = h.TrainingDatasetRepo.GetAllUser(loggedUserId.(uint))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		trainingDatasets, err = h.TrainingDatasetRepo.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = h.Template.ExecuteTemplate(w, "training-datasets_list", TemplateData{
		TrainingDatasets: trainingDatasets,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TrainingDatasetHandler) New(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title string
	}

	err := h.Template.ExecuteTemplate(w, "training-datasets_new", TemplateData{
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

	sess := h.Auth.GlobalSessions.SessionStart(w, r)
	loggedUserId := sess.Get("loggedUserId")
	if loggedUserId != nil {
		loggedUser, err := h.UserRepo.GetByID(loggedUserId.(uint))
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		trainingDataset.UserId = loggedUser.ID
	}

	err = h.TrainingDatasetRepo.Create(&trainingDataset)
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

	sess := h.Auth.GlobalSessions.SessionStart(w, r)
	loggedUserId := sess.Get("loggedUserId")
	if loggedUserId != nil {
		_, err := h.UserRepo.GetByID(loggedUserId.(uint))
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	err = h.TrainingDatasetRepo.Delete(loggedUserId.(uint), uint(trainingDatasetId))
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

	err = h.Template.ExecuteTemplate(w, "training-datasets_tree-browser", TemplateData{
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

	err = h.Template.ExecuteTemplate(w, "training-datasets_file-list", TemplateData{
		AODFiles: aods,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
	validateHtmxMw := middleware.NewValidateHTMXMw()
	blockHtmxMw := middleware.NewBlockHTMXMw()

	mux.Handle(fmt.Sprintf("GET /%s", prefix), middleware.Chain(
		http.HandlerFunc(tjh.Index),
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
