package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/mytkom/AliceTraINT/internal/hash"
	"github.com/mytkom/AliceTraINT/internal/middleware"
)

type TrainingMachineHandler struct {
	*environment.Env
}

func NewTrainingMachineHandler(env *environment.Env) *TrainingMachineHandler {
	return &TrainingMachineHandler{
		Env: env,
	}
}

func (h *TrainingMachineHandler) Index(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title string
	}

	err := h.ExecuteTemplate(w, "training-machines_index", TemplateData{
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
		loggedUser, err := getAuthorizedUser(h.Auth, h.User, w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		trainingMachines, err = h.TrainingMachine.GetAllUser(loggedUser.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		trainingMachines, err = h.TrainingMachine.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = h.ExecuteTemplate(w, "training-machines_list", TemplateData{
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

	trainingMachine, err := h.TrainingMachine.GetByID(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.ExecuteTemplate(w, "training-machines_show", TemplateData{
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
		NNArchSpec map[string]NNConfigField
	}

	err := h.ExecuteTemplate(w, "training-machines_new", TemplateData{
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

	loggedUser, err := getAuthorizedUser(h.Auth, h.User, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	trainingMachine.UserId = loggedUser.ID

	secretKey, err := hash.GenerateKey(32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	trainingMachine.SecretKeyHashed, err = hash.HashKey(secretKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.TrainingMachine.Create(&trainingMachine)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.ExecuteTemplate(w, "training-machines_show-secret", TemplateData{
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

	loggedUser, err := getAuthorizedUser(h.Auth, h.User, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = h.TrainingMachine.Delete(loggedUser.ID, uint(trainingMachineId))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func InitTrainingMachineRoutes(mux *http.ServeMux, env *environment.Env) {
	prefix := "training-machines"

	tmh := &TrainingMachineHandler{
		Env: env,
	}

	authMw := middleware.NewAuthMw(env.Auth, true)
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
