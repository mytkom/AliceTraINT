package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/hash"
	"github.com/mytkom/AliceTraINT/internal/jalien"
)

type QueueHandler struct {
	TrainingMachineRepo    repository.TrainingMachineRepository
	TrainingTaskRepo       repository.TrainingTaskRepository
	TrainingTaskResultRepo repository.TrainingTaskResultRepository
}

func (qh *QueueHandler) getAuthorizedTrainingMachine(r *http.Request, tmId uint) (*models.TrainingMachine, error) {
	secretId := r.Header.Get("Secret-Id")

	trainingMachine, err := qh.TrainingMachineRepo.GetByID(tmId)
	if err != nil {
		return nil, err
	}

	ok, err := hash.VerifyKey(secretId, trainingMachine.SecretKeyHashed)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("authorization failure")
	}

	trainingMachine.LastActivityAt = time.Now()
	qh.TrainingMachineRepo.Update(trainingMachine)

	return trainingMachine, nil
}

func (qh *QueueHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	type Body struct {
		Status models.TrainingTaskStatus
	}

	ttIdStr := r.PathValue("id")
	ttId, err := strconv.ParseUint(ttIdStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid training dataset id", http.StatusUnprocessableEntity)
		return
	}

	tt, err := qh.TrainingTaskRepo.GetByID(uint(ttId))
	if err != nil {
		http.Error(w, "training task does not exist", http.StatusNotFound)
		return
	}

	if tt.TrainingMachineId == nil {
		http.Error(w, "unauthorized machine", http.StatusUnauthorized)
		return
	}

	_, err = qh.getAuthorizedTrainingMachine(r, *tt.TrainingMachineId)
	if err != nil {
		http.Error(w, "unauthorized machine", http.StatusUnauthorized)
		return
	}

	bodyDecoded := Body{}
	err = json.NewDecoder(r.Body).Decode(&bodyDecoded)
	if err != nil {
		http.Error(w, "bad status format or unexisting status", http.StatusUnprocessableEntity)
		return
	}

	tt.Status = bodyDecoded.Status

	err = qh.TrainingTaskRepo.Update(tt)
	if err != nil {
		http.Error(w, "error on status update", http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (qh *QueueHandler) QueryTask(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		ID            uint
		AODFiles      []jalien.AODFile
		Configuration interface{}
	}

	tmIdStr := r.PathValue("id")
	tmId, err := strconv.ParseUint(tmIdStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid training machine id", http.StatusUnprocessableEntity)
		return
	}

	tm, err := qh.getAuthorizedTrainingMachine(r, uint(tmId))
	if err != nil {
		http.Error(w, "unauthorized machine", http.StatusUnauthorized)
		return
	}

	tt, err := qh.TrainingTaskRepo.GetFirstQueued()
	if err != nil {
		http.Error(w, "no task to run", http.StatusNotFound)
		return
	}

	tt.TrainingMachineId = &tm.ID
	tt.Status = models.Training

	err = qh.TrainingTaskRepo.Update(tt)
	if err != nil {
		http.Error(w, "cannot assign task to machine", http.StatusUnprocessableEntity)
		return
	}

	j, err := json.Marshal(Response{
		ID:            tt.ID,
		AODFiles:      tt.TrainingDataset.AODFiles,
		Configuration: tt.Configuration,
	})
	if err != nil {
		http.Error(w, "cannot marshal response", http.StatusUnprocessableEntity)
		return
	}

	w.Write(j)
	w.WriteHeader(http.StatusOK)
}

func (qh *QueueHandler) CreateTrainingTaskResult(w http.ResponseWriter, r *http.Request) {
	ttIdStr := r.PathValue("id")
	ttId, err := strconv.ParseUint(ttIdStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid training dataset id", http.StatusUnprocessableEntity)
		return
	}

	tt, err := qh.TrainingTaskRepo.GetByID(uint(ttId))
	if err != nil {
		http.Error(w, "training task does not exist", http.StatusNotFound)
		return
	}

	if tt.TrainingMachineId == nil {
		http.Error(w, "unauthorized machine", http.StatusUnauthorized)
		return
	}

	_, err = qh.getAuthorizedTrainingMachine(r, *tt.TrainingMachineId)
	if err != nil {
		http.Error(w, "unauthorized machine", http.StatusUnauthorized)
		return
	}

	var ttr models.TrainingTaskResult
	err = json.NewDecoder(r.Body).Decode(&ttr)
	if err != nil {
		http.Error(w, "incorrect data format", http.StatusUnprocessableEntity)
		return
	}

	ttr.TrainingTaskId = tt.ID

	err = qh.TrainingTaskResultRepo.Create(&ttr)
	if err != nil {
		http.Error(w, "error during task result creation", http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func InitQueryRoutes(mux *http.ServeMux, tmRepo repository.TrainingMachineRepository, ttRepo repository.TrainingTaskRepository, ttrRepo repository.TrainingTaskResultRepository) {
	qh := &QueueHandler{
		TrainingMachineRepo:    tmRepo,
		TrainingTaskRepo:       ttRepo,
		TrainingTaskResultRepo: ttrRepo,
	}

	mux.Handle("POST /training_tasks/{id}/status", http.HandlerFunc(qh.UpdateStatus))
	mux.Handle("GET /training_machines/{id}/training_task", http.HandlerFunc(qh.QueryTask))
	mux.Handle("POST /training_tasks/{id}/training_task_results", http.HandlerFunc(qh.CreateTrainingTaskResult))
}
