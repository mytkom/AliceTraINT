package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	utils "github.com/mytkom/AliceTraINT/internal/hash"
	"github.com/mytkom/AliceTraINT/internal/middleware"
)

type TrainingMachineHandler struct {
	TrainingMachineRepo repository.TrainingMachineRepository
	UserRepo            repository.UserRepository
	Auth                *auth.Auth
	Template            *template.Template
}

func (h *TrainingMachineHandler) Index(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title string
	}

	err := h.Template.ExecuteTemplate(w, "training-machines_index", TemplateData{
		Title: "Training Machines",
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TrainingMachineHandler) List(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		TrainingMachines []models.TrainingMachine
	}

	var trainingMachines []models.TrainingMachine
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

		trainingMachines, err = h.TrainingMachineRepo.GetAllUser(loggedUserId.(uint))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		trainingMachines, err = h.TrainingMachineRepo.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = h.Template.ExecuteTemplate(w, "training-machines_list", TemplateData{
		TrainingMachines: trainingMachines,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TrainingMachineHandler) Show(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title           string
		TrainingMachine models.TrainingMachine
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid training machine id", http.StatusUnprocessableEntity)
		return
	}

	trainingMachine, err := h.TrainingMachineRepo.GetByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.Template.ExecuteTemplate(w, "training-machines_show", TemplateData{
		Title:           "Training Machine",
		TrainingMachine: *trainingMachine,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TrainingMachineHandler) New(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title      string
		NNArchSpec map[string]NNArchSpec
	}

	err := h.Template.ExecuteTemplate(w, "training-machines_new", TemplateData{
		Title: "Register New Training Machine!",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TrainingMachineHandler) Create(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		ID        uint
		SecretKey string
	}

	var trainingMachine models.TrainingMachine
	err := json.NewDecoder(r.Body).Decode(&trainingMachine)
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

		trainingMachine.UserId = loggedUser.ID
	}

	secretKey, err := utils.GenerateKey(32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	trainingMachine.SecretKeyHashed, err = utils.HashKey(secretKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.TrainingMachineRepo.Create(&trainingMachine)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.Template.ExecuteTemplate(w, "training-machines_show-secret", TemplateData{
		ID:        trainingMachine.ID,
		SecretKey: secretKey,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TrainingMachineHandler) Delete(w http.ResponseWriter, r *http.Request) {
	trainingMachineId, err := strconv.ParseUint(r.PathValue("id"), 10, 32)

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

	err = h.TrainingMachineRepo.Delete(loggedUserId.(uint), uint(trainingMachineId))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func InitTrainingMachineRoutes(mux *http.ServeMux, baseTemplate *template.Template, trainingMachineRepo repository.TrainingMachineRepository, userRepo repository.UserRepository, auth *auth.Auth) {
	prefix := "training-machines"

	tmh := &TrainingMachineHandler{
		TrainingMachineRepo: trainingMachineRepo,
		UserRepo:            userRepo,
		Auth:                auth,
		Template:            baseTemplate,
	}

	authMw := middleware.NewAuthMw(auth)
	validateHtmxMw := middleware.NewValidateHTMXMw()
	blockHtmxMw := middleware.NewBlockHTMXMw()

	mux.Handle(fmt.Sprintf("GET /%s", prefix), middleware.Chain(
		http.HandlerFunc(tmh.Index),
		blockHtmxMw,
		authMw,
	))

	mux.Handle(fmt.Sprintf("GET /%s/{id}", prefix), middleware.Chain(
		http.HandlerFunc(tmh.Show),
		blockHtmxMw,
		authMw,
	))

	mux.Handle(fmt.Sprintf("GET /%s/list", prefix), middleware.Chain(
		http.HandlerFunc(tmh.List),
		validateHtmxMw,
		authMw,
	))

	mux.Handle(fmt.Sprintf("GET /%s/new", prefix), middleware.Chain(
		http.HandlerFunc(tmh.New),
		blockHtmxMw,
		authMw,
	))

	mux.Handle(fmt.Sprintf("POST /%s", prefix), middleware.Chain(
		http.HandlerFunc(tmh.Create),
		validateHtmxMw,
		authMw,
	))

	mux.Handle(fmt.Sprintf("DELETE /%s/{id}", prefix), middleware.Chain(
		http.HandlerFunc(tmh.Delete),
		validateHtmxMw,
		authMw,
	))
}
