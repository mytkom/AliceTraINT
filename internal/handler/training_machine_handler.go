package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/mytkom/AliceTraINT/internal/middleware"
	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/mytkom/AliceTraINT/internal/utils"
)

type TrainingMachineHandler struct {
	*environment.Env
	Service service.ITrainingMachineService
}

func NewTrainingMachineHandler(env *environment.Env, tmService service.ITrainingMachineService) *TrainingMachineHandler {
	return &TrainingMachineHandler{
		Env:     env,
		Service: tmService,
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

	user, ok := middleware.GetLoggedUser(r)
	if !ok || user == nil {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	trainingMachines, err := h.Service.GetAll(user.ID, utils.IsUserScoped(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	trainingMachine, err := h.Service.GetByID(uint(id))
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
		Title string
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

	user, ok := middleware.GetLoggedUser(r)
	if !ok || user == nil {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}
	trainingMachine.UserId = user.ID

	secretKey, err := h.Service.Create(&trainingMachine)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

	user, ok := middleware.GetLoggedUser(r)
	if !ok || user == nil {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	err = h.Service.Delete(user.ID, uint(trainingMachineId))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func InitTrainingMachineRoutes(mux *http.ServeMux, env *environment.Env, hasher service.Hasher) {
	prefix := "training-machines"

	tmService := service.NewTrainingMachineService(env.RepositoryContext, hasher)
	tmh := NewTrainingMachineHandler(env, tmService)
	authMw := middleware.NewAuthMw(env.IAuthService, true)
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
