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

type TrainingTaskHandler struct {
	*environment.Env
	Service service.ITrainingTaskService
}

func NewTrainingTaskHandler(env *environment.Env, ttService service.ITrainingTaskService) *TrainingTaskHandler {
	return &TrainingTaskHandler{
		Env:     env,
		Service: ttService,
	}
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

	user, ok := middleware.GetLoggedUser(r)
	if !ok || user == nil {
		http.Error(w, errMsgUserUnauthorized, http.StatusUnauthorized)
		return
	}

	trainingTasks, err := h.Service.GetAll(user.ID, utils.IsUserScoped(r))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	err = h.ExecuteTemplate(w, "training-tasks_list", TemplateData{
		TrainingTasks: trainingTasks,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TrainingTaskHandler) UploadToCCDB(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid training task id", http.StatusUnprocessableEntity)
		return
	}

	err = h.Service.UploadOnnxResults(uint(id))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	utils.HTMXRefresh(w)
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

	tt, err := h.Service.GetByID(uint(id))
	if err != nil {
		handleServiceError(w, err)
		return
	}

	err = h.ExecuteTemplate(w, "training-tasks_show", TemplateData{
		Title:        "Training Tasks",
		TrainingTask: *tt.TrainingTask,
		ImageFiles:   tt.ImageFiles,
		OnnxFiles:    tt.OnnxFiles,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TrainingTaskHandler) New(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title            string
		TrainingDatasets []models.TrainingDataset
		FieldConfigs     service.NNFieldConfigs
	}

	user, ok := middleware.GetLoggedUser(r)
	if !ok || user == nil {
		http.Error(w, errMsgUserUnauthorized, http.StatusUnauthorized)
		return
	}

	ttHelpers, err := h.Service.GetHelpers(user.ID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	err = h.ExecuteTemplate(w, "training-tasks_new", TemplateData{
		Title:            "Create New Training Task!",
		TrainingDatasets: ttHelpers.TrainingDatasets,
		FieldConfigs:     ttHelpers.FieldConfigs,
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

	user, ok := middleware.GetLoggedUser(r)
	if !ok || user == nil {
		http.Error(w, errMsgUserUnauthorized, http.StatusUnauthorized)
		return
	}
	trainingTask.UserId = user.ID

	err = h.Service.Create(&trainingTask)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	utils.HTMXRedirect(w, "/training-tasks")
	w.WriteHeader(http.StatusCreated)
}

func InitTrainingTaskRoutes(mux *http.ServeMux, env *environment.Env, ccdbService service.ICCDBService, jalienService service.IJAliEnService, fileService service.IFileService, nnArch service.INNArchService) {
	prefix := "training-tasks"

	ttService := service.NewTrainingTaskService(env.RepositoryContext, ccdbService, jalienService, fileService, nnArch)
	tjh := NewTrainingTaskHandler(env, ttService)

	authMw := middleware.NewAuthMw(env.IAuthService, true)
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
