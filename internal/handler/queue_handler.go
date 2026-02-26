package handler

import (
	"encoding/json"
	"errors"
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

const (
	errMsgUnauthorizedMachine string = "unauthorized machine"
)

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
		return nil, nil, errors.New(errMsgUnauthorizedMachine)
	}

	tm, err := qh.QueueService.AuthorizeTrainingMachine(r.Header.Get("Secret-Id"), *tt.TrainingMachineId)
	if err != nil {
		return nil, nil, errors.New(errMsgUnauthorizedMachine)
	}

	return tm, tt, nil
}

func (qh *QueueHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	type Body struct {
		Status models.TrainingTaskStatus
	}

	_, tt, err := qh.trainingMachineFromPath(r)
	if err != nil {
		// Unauthorized machine or invalid task – surface message but log details.
		writeError(w, r, http.StatusUnauthorized, err.Error(), err)
		return
	}

	var bodyDecoded Body
	if err := json.NewDecoder(r.Body).Decode(&bodyDecoded); err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "bad status format", err)
		return
	}

	if err := qh.QueueService.UpdateTrainingTaskStatus(tt.ID, bodyDecoded.Status); err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "cannot update training task status", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (qh *QueueHandler) QueryTask(w http.ResponseWriter, r *http.Request) {
	tmId, err := qh.parseId(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "bad training machine id", err)
		return
	}

	tm, err := qh.QueueService.AuthorizeTrainingMachine(r.Header.Get("Secret-Id"), uint(tmId))
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, errMsgUnauthorizedMachine, err)
		return
	}

	tt, err := qh.QueueService.AssignTaskToMachine(tm.ID)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "no training task available", err)
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

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		writeError(w, r, http.StatusInternalServerError, "cannot encode response", err)
		return
	}
}

func (qh *QueueHandler) CreateTrainingTaskResult(w http.ResponseWriter, r *http.Request) {
	_, tt, err := qh.trainingMachineFromPath(r)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, err.Error(), err)
		return
	}

	err = r.ParseMultipartForm(20 << 20)
	if err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "error reading multipart input", err)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "error reading file", err)
		return
	}
	//nolint:errcheck
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
		writeError(w, r, http.StatusUnprocessableEntity, "cannot create training task result", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ttr)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
		return
	}
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
