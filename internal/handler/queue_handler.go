package handler

import (
	"fmt"
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/hash"
)

type QueryHandler struct {
	TrainingMachineRepo repository.TrainingMachineRepository
}

func (qh *QueryHandler) getAuthorizedTrainingMachine(r *http.Request, tmId uint) (*models.TrainingMachine, error) {
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

	return trainingMachine, nil
}
