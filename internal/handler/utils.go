package handler

import (
	"fmt"
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
)

func getAuthorizedUser(a *auth.Auth, userRepo repository.UserRepository, w http.ResponseWriter, r *http.Request) (*models.User, error) {
	sess := a.GlobalSessions.SessionStart(w, r)
	loggedUserId := sess.Get("loggedUserId")
	if loggedUserId != nil {
		loggedUser, err := userRepo.GetByID(loggedUserId.(uint))
		if err != nil {
			return nil, err
		}
		return loggedUser, nil
	}

	return nil, fmt.Errorf("user not logged in")
}
