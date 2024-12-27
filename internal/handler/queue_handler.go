package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/mytkom/AliceTraINT/internal/service"
)

type QueueHandler struct {
	*environment.Env
	QueueService service.IQueueService
}

func (qh *QueueHandler) parseId(r *http.Request) (uint, error) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("error parsing id: %s", err.Error())
	}

	return uint(id), nil
}

func (qh *QueueHandler) trainingMachineFromPath(r *http.Request) (*models.TrainingMachine, *models.TrainingTask, error) {
	ttId, err := qh.parseId(r)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid training task id: %s", err.Error())
	}

	tt, err := qh.TrainingTask.GetByID(uint(ttId))
	if err != nil {
		return nil, nil, fmt.Errorf("training task does not exist: %s", err.Error())
	}

	if tt.TrainingMachineId == nil {
		return nil, nil, fmt.Errorf("unauthorized machine")
	}

	tm, err := qh.QueueService.AuthorizeTrainingMachine(r.Header.Get("Secret-Id"), *tt.TrainingMachineId)
	if err != nil {
		return nil, nil, fmt.Errorf("unauthorized machine")
	}

	return tm, tt, nil
}

func (qh *QueueHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	type Body struct {
		Status models.TrainingTaskStatus
	}

	_, tt, err := qh.trainingMachineFromPath(r)
	if err != nil {
		http.Error(w, "unauthorized machine", http.StatusUnauthorized)
		return
	}

	var bodyDecoded Body
	if err := json.NewDecoder(r.Body).Decode(&bodyDecoded); err != nil {
		http.Error(w, "bad status format", http.StatusUnprocessableEntity)
		return
	}

	if err := qh.QueueService.UpdateTrainingTaskStatus(tt.ID, bodyDecoded.Status); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (qh *QueueHandler) QueryTask(w http.ResponseWriter, r *http.Request) {
	tmId, err := qh.parseId(r)
	if err != nil {
		http.Error(w, "bad training machine id", http.StatusBadRequest)
		return
	}

	tm, err := qh.QueueService.AuthorizeTrainingMachine(r.Header.Get("Secret-Id"), uint(tmId))
	if err != nil {
		http.Error(w, "unauthorized machine", http.StatusUnauthorized)
		return
	}

	tt, err := qh.QueueService.AssignTaskToMachine(tm.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := struct {
		ID            uint
		AODFiles      []jalien.AODFile
		Configuration interface{}
	}{
		ID:            tt.ID,
		AODFiles:      tt.TrainingDataset.AODFiles,
		Configuration: tt.Configuration,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "cannot encode response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (qh *QueueHandler) CreateTrainingTaskResult(w http.ResponseWriter, r *http.Request) {
	_, tt, err := qh.trainingMachineFromPath(r)
	if err != nil {
		http.Error(w, "unauthorized machine", http.StatusUnauthorized)
		return
	}

	r.ParseMultipartForm(20 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "error reading file", http.StatusUnprocessableEntity)
		return
	}
	defer file.Close()

	ttr, err := qh.QueueService.CreateTrainingTaskResult(
		tt.ID,
		file,
		handler,
		r.Form.Get("name"),
		r.Form.Get("description"),
		r.Form.Get("file-type"),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ttr)
}

func InitQueueRoutes(mux *http.ServeMux, env *environment.Env, fileService service.IFileService, hasher service.Hasher) {
	qh := &QueueHandler{
		Env:          env,
		QueueService: service.NewQueueService(fileService, env.RepositoryContext, hasher),
	}

	mux.Handle("POST /training-tasks/{id}/status", http.HandlerFunc(qh.UpdateStatus))
	mux.Handle("GET /training-machines/{id}/training-task", http.HandlerFunc(qh.QueryTask))
	mux.Handle("POST /training-tasks/{id}/training-task-results", http.HandlerFunc(qh.CreateTrainingTaskResult))
}
