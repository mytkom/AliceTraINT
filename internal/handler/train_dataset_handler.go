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

type TrainDatasetHandler struct {
	TrainDatasetRepo repository.TrainDatasetRepository
	UserRepo         repository.UserRepository
	Auth             *auth.Auth
	Template         *template.Template
}

type IndexTemplateData struct {
	Title         string
	TrainDatasets []models.TrainDataset
}

func (h *TrainDatasetHandler) Index(w http.ResponseWriter, r *http.Request) {
	sess := h.Auth.GlobalSessions.SessionStart(w, r)
	loggedUserId := sess.Get("loggedUserId")
	if loggedUserId != nil {
		_, err := h.UserRepo.GetByID(loggedUserId.(uint))
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}
	}

	trainDatasets, err := h.TrainDatasetRepo.GetAllUser(loggedUserId.(uint))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = h.Template.ExecuteTemplate(w, "train-datasets_index", IndexTemplateData{
		Title:         "Train Datasets",
		TrainDatasets: trainDatasets,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TrainDatasetHandler) New(w http.ResponseWriter, r *http.Request) {
	err := h.Template.ExecuteTemplate(w, "train-datasets_new", map[string]interface{}{
		"Title": "Create New Train Job!",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TrainDatasetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var trainDataset models.TrainDataset
	err := json.NewDecoder(r.Body).Decode(&trainDataset)
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

		trainDataset.UserId = loggedUser.ID
	}

	err = h.TrainDatasetRepo.Create(&trainDataset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *TrainDatasetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	trainDatasetId, err := strconv.ParseUint(r.PathValue("id"), 10, 32)

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

	err = h.TrainDatasetRepo.Delete(loggedUserId.(uint), uint(trainDatasetId))
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

func (h *TrainDatasetHandler) ExploreDirectory(w http.ResponseWriter, r *http.Request) {
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

	err = h.Template.ExecuteTemplate(w, "train-datasets_tree-browser", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type findAODsTemplateData struct {
	AODFiles []jalien.AODFile
}

func (h *TrainDatasetHandler) FindAods(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	aods, err := jalien.FindAODFiles(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	data := &findAODsTemplateData{
		AODFiles: aods,
	}

	err = h.Template.ExecuteTemplate(w, "train-datasets_file-list", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func InitTrainDatasetRoutes(mux *http.ServeMux, baseTemplate *template.Template, trainDatasetRepo repository.TrainDatasetRepository, userRepo repository.UserRepository, auth *auth.Auth) {
	prefix := "train-datasets"

	tjh := &TrainDatasetHandler{
		TrainDatasetRepo: trainDatasetRepo,
		UserRepo:         userRepo,
		Auth:             auth,
		Template:         baseTemplate,
	}

	cache := utils.NewCache(60 * time.Minute)

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
