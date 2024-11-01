package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/middleware"
	"github.com/mytkom/AliceTraINT/internal/utils"
)

type NNArchSpec struct {
	FullName     string      `json:"full_name"`
	Type         string      `json:"type"`
	DefaultValue interface{} `json:"default_value"`
	Min          interface{} `json:"min"`
	Max          interface{} `json:"max"`
	Step         interface{} `json:"step"`
	Description  string      `json:"description"`
}

func loadNNArchSpec(filename string) (map[string]NNArchSpec, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	var config map[string]NNArchSpec
	bytes, err := io.ReadAll(file)

	if err != nil {
		return nil, err
	}

	err = file.Close()

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &config)

	if err != nil {
		return nil, err
	}

	return config, nil
}

type TrainingTaskHandler struct {
	TrainingTaskRepo    repository.TrainingTaskRepository
	TrainingDatasetRepo repository.TrainingDatasetRepository
	UserRepo            repository.UserRepository
	Auth                *auth.Auth
	Template            *template.Template
	NNArchSpec          map[string]NNArchSpec
}

func (h *TrainingTaskHandler) Index(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title string
	}

	err := h.Template.ExecuteTemplate(w, "training-tasks_index", TemplateData{
		Title: "Training Tasks",
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TrainingTaskHandler) List(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		TrainingTasks []models.TrainingTask
	}

	var trainingTasks []models.TrainingTask
	var err error

	if r.URL.Query().Get("userScoped") == "on" {
		sess := h.Auth.GlobalSessions.SessionStart(w, r)
		loggedUserId := sess.Get("loggedUserId")
		if loggedUserId != nil {
			_, err := h.UserRepo.GetByID(loggedUserId.(uint))
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			}
		}

		trainingTasks, err = h.TrainingTaskRepo.GetAllUser(loggedUserId.(uint))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		trainingTasks, err = h.TrainingTaskRepo.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	err = h.Template.ExecuteTemplate(w, "training-tasks_list", TemplateData{
		TrainingTasks: trainingTasks,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TrainingTaskHandler) New(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title            string
		TrainingDatasets []models.TrainingDataset
		NNArchSpec       map[string]NNArchSpec
	}

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

	err = h.Template.ExecuteTemplate(w, "training-tasks_new", TemplateData{
		Title:            "Create New Training Task!",
		TrainingDatasets: trainingDatasets,
		NNArchSpec:       h.NNArchSpec,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TrainingTaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var trainingTask models.TrainingTask
	err := json.NewDecoder(r.Body).Decode(&trainingTask)
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

		trainingTask.UserId = loggedUser.ID
	}

	err = h.TrainingTaskRepo.Create(&trainingTask)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	utils.HTMXRedirect(w, "/training-tasks")
	w.WriteHeader(http.StatusCreated)
}

func InitTrainingTasksRoutes(mux *http.ServeMux, baseTemplate *template.Template, trainingTaskRepo repository.TrainingTaskRepository, trainingDatasetRepo repository.TrainingDatasetRepository, userRepo repository.UserRepository, auth *auth.Auth) {
	prefix := "training-tasks"

	nnArchSpec, err := loadNNArchSpec("internal/nn_architectures/proposed.json")
	if err != nil {
		log.Fatal("cannot load architecture configuration specification file")
	}

	tjh := &TrainingTaskHandler{
		TrainingTaskRepo:    trainingTaskRepo,
		TrainingDatasetRepo: trainingDatasetRepo,
		UserRepo:            userRepo,
		Auth:                auth,
		Template:            baseTemplate,
		NNArchSpec:          nnArchSpec,
	}

	authMw := middleware.NewAuthMw(auth)
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

	mux.Handle(fmt.Sprintf("POST /%s", prefix), middleware.Chain(
		http.HandlerFunc(tjh.Create),
		validateHtmxMw,
		authMw,
	))
}
