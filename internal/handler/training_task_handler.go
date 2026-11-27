package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
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

type taskListOption struct {
	Value string
	Label string
}

var taskListSortOptions = []taskListOption{
	{"newest", "Newest first"},
	{"oldest", "Oldest first"},
	{"updated", "Last updated"},
	{"name-asc", "Name (A–Z)"},
	{"name-desc", "Name (Z–A)"},
}

func taskListFilters(r *http.Request) (repository.TrainingTaskListQuery, url.Values, []taskListOption) {
	values := r.URL.Query()
	params := url.Values{}
	for _, key := range []string{"name", "author", "status", "from", "to", "sort", "userScoped"} {
		if value := strings.TrimSpace(values.Get(key)); value != "" {
			params.Set(key, value)
		}
	}

	query := repository.TrainingTaskListQuery{
		Name: params.Get("name"), Author: params.Get("author"), Page: 1,
	}
	if page, err := strconv.Atoi(values.Get("page")); err == nil && page > 1 && page <= int(^uint(0)>>1)/repository.TrainingTaskPageSize {
		query.Page = page
		params.Set("page", strconv.Itoa(page))
	}

	for _, option := range taskListSortOptions {
		if option.Value == params.Get("sort") {
			query.Sort = option.Value
			break
		}
	}
	if query.Sort == "" {
		query.Sort = "newest"
		params.Del("sort")
	}

	var statuses []taskListOption
	for status := models.Failed; status <= models.Uploaded; status++ {
		value := strconv.FormatUint(uint64(status), 10)
		statuses = append(statuses, taskListOption{value, status.String()})
		if value == params.Get("status") {
			selected := status
			query.Status = &selected
		}
	}
	if query.Status == nil {
		params.Del("status")
	}
	if !utils.IsUserScoped(r) {
		params.Del("userScoped")
	}
	if from, err := time.Parse(time.DateOnly, params.Get("from")); err == nil {
		query.CreatedFrom = from
	} else {
		params.Del("from")
	}
	if to, err := time.Parse(time.DateOnly, params.Get("to")); err == nil {
		query.CreatedBefore = to.AddDate(0, 0, 1)
	} else {
		params.Del("to")
	}
	return query, params, statuses
}

func (h *TrainingTaskHandler) Index(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title         string
		Params        url.Values
		ListURL       string
		Sort          string
		SortOptions   []taskListOption
		StatusOptions []taskListOption
	}

	query, params, statuses := taskListFilters(r)
	listURL := url.URL{Path: "/training-tasks/list", RawQuery: params.Encode()}
	err := h.ExecuteTemplate(w, "training-tasks_index", TemplateData{
		Title:         "Training Tasks",
		Params:        params,
		ListURL:       listURL.String(),
		Sort:          query.Sort,
		SortOptions:   taskListSortOptions,
		StatusOptions: statuses,
	})

	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
		return
	}
}

func (h *TrainingTaskHandler) List(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		TrainingTasks         []models.TrainingTask
		Page                  int
		First, Previous, Next string
	}

	user, ok := middleware.GetLoggedUser(r)
	if !ok || user == nil {
		writeError(w, r, http.StatusUnauthorized, errMsgUserUnauthorized, nil)
		return
	}

	query, params, _ := taskListFilters(r)
	if utils.IsUserScoped(r) {
		query.OwnerID = &user.ID
	}
	page, err := h.Service.List(query)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	pageURL := func(number int) string {
		if number == 1 {
			params.Del("page")
		} else {
			params.Set("page", strconv.Itoa(number))
		}
		link := url.URL{Path: "/training-tasks", RawQuery: params.Encode()}
		return link.String()
	}
	data := TemplateData{TrainingTasks: page.Tasks, Page: query.Page}
	if query.Page > 1 {
		data.First = pageURL(1)
		if len(page.Tasks) > 0 {
			data.Previous = pageURL(query.Page - 1)
		}
	}
	if page.HasNext {
		data.Next = pageURL(query.Page + 1)
	}

	err = h.ExecuteTemplate(w, "training-tasks_list", data)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
		return
	}
}

func (h *TrainingTaskHandler) UploadToCCDB(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "invalid training task id", err)
		return
	}

	err = h.Service.UploadOnnxResults(uint(id))
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	utils.HTMXRefresh(w)
	w.WriteHeader(http.StatusOK)
}

func (h *TrainingTaskHandler) Restart(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetLoggedUser(r)
	if !ok || user == nil {
		writeError(w, r, http.StatusUnauthorized, errMsgUserUnauthorized, nil)
		return
	}

	id, err := strconv.ParseUint(r.PathValue("id"), 10, 32)
	if err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "invalid training task id", err)
		return
	}

	if err := h.Service.Restart(user.ID, uint(id)); err != nil {
		handleServiceError(w, r, err)
		return
	}

	utils.HTMXRefresh(w)
	w.WriteHeader(http.StatusOK)
}

func (h *TrainingTaskHandler) Show(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title            string
		TrainingTask     models.TrainingTask
		ImageFiles       []models.TrainingTaskResult
		OnnxFiles        []models.TrainingTaskResult
		LogFiles         []models.TrainingTaskResult
		HasConfiguration bool
		CanRestart       bool
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "invalid training task id", err)
		return
	}

	tt, err := h.Service.GetByID(uint(id))
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	user, _ := middleware.GetLoggedUser(r)
	err = h.ExecuteTemplate(w, "training-tasks_show", TemplateData{
		Title:            "Training Tasks",
		TrainingTask:     *tt.TrainingTask,
		ImageFiles:       tt.ImageFiles,
		OnnxFiles:        tt.OnnxFiles,
		LogFiles:         tt.LogFiles,
		HasConfiguration: utils.HasConfiguration(tt.TrainingTask.Configuration),
		CanRestart:       user != nil && tt.TrainingTask.UserId == user.ID && tt.TrainingTask.Status == models.Failed,
	})

	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
	}
}

func (h *TrainingTaskHandler) New(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title                      string
		TrainingDatasets           []models.TrainingDataset
		FieldGroups                []service.NNFieldGroup
		PrefillConfig              template.JS
		PrefillDatasetId           uint
		PrefillNameSuggest         string
		PrefillSubsampleEventCount uint
		PrefillEnableSubsample     bool
	}

	ttHelpers, err := h.Service.GetHelpers()
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	data := TemplateData{
		Title:                      "Create New Training Task!",
		TrainingDatasets:           ttHelpers.TrainingDatasets,
		FieldGroups:                ttHelpers.FieldGroups,
		PrefillConfig:              "null",
		PrefillSubsampleEventCount: models.DefaultSubsampleEventCount,
		PrefillEnableSubsample:     false,
	}

	if fromTask := r.URL.Query().Get("fromTask"); fromTask != "" {
		fromID, parseErr := strconv.ParseUint(fromTask, 10, 32)
		if parseErr != nil {
			writeError(w, r, http.StatusUnprocessableEntity, "invalid fromTask id", parseErr)
			return
		}
		source, getErr := h.Service.GetByID(uint(fromID))
		if getErr != nil {
			handleServiceError(w, r, getErr)
			return
		}
		data.PrefillConfig = utils.ConfigJSON(source.TrainingTask.Configuration)
		data.PrefillDatasetId = source.TrainingTask.TrainingDatasetId
		data.PrefillNameSuggest = source.TrainingTask.Name + " (copy)"
		if source.TrainingTask.SubsampleEventCount > 0 {
			data.PrefillSubsampleEventCount = source.TrainingTask.SubsampleEventCount
			data.PrefillEnableSubsample = true
		}
	}

	err = h.ExecuteTemplate(w, "training-tasks_new", data)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
		return
	}
}

func (h *TrainingTaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var trainingTask models.TrainingTask
	err := json.NewDecoder(r.Body).Decode(&trainingTask)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid request payload", err)
		return
	}

	user, ok := middleware.GetLoggedUser(r)
	if !ok || user == nil {
		writeError(w, r, http.StatusUnauthorized, errMsgUserUnauthorized, nil)
		return
	}
	trainingTask.UserId = user.ID

	err = h.Service.Create(&trainingTask)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	utils.HTMXRedirect(w, "/training-tasks")
	w.WriteHeader(http.StatusCreated)
}

func InitTrainingTaskRoutes(mux *http.ServeMux, env *environment.Env, ccdbService service.ICCDBService, monalisaService service.IMonalisaService, fileService service.IFileService, nnArch service.INNArchService) {
	prefix := "training-tasks"

	ttService := service.NewTrainingTaskService(env.RepositoryContext, ccdbService, monalisaService, fileService, nnArch)
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

	mux.Handle(fmt.Sprintf("POST /%s/{id}/restart", prefix), middleware.Chain(
		http.HandlerFunc(tjh.Restart),
		validateHtmxMw,
		authMw,
	))
}
