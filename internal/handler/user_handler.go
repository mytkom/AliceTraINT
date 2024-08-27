package handler

import (
	"html/template"
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/thomasdarimont/go-kc-example/session"
	_ "github.com/thomasdarimont/go-kc-example/session_memory"
)

type UserHandler struct {
	UserRepo       repository.UserRepository
	Templates      *template.Template
	GlobalSessions *session.Manager
}

func NewUserHandler(userRepo repository.UserRepository, templates *template.Template, globalSessions *session.Manager) *UserHandler {
	return &UserHandler{UserRepo: userRepo, Templates: templates, GlobalSessions: globalSessions}
}

func (h *UserHandler) Index(w http.ResponseWriter, r *http.Request) {
	users, err := h.UserRepo.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Users      []models.User
		LoggedUser *models.User
	}{Users: users}

	sess := h.GlobalSessions.SessionStart(w, r)
	loggedUserId := sess.Get("loggedUserId")
	if loggedUserId != nil {
		loggedUser, _ := h.UserRepo.GetUserByID(loggedUserId.(int))
		data.LoggedUser = loggedUser
	}

	err = h.Templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, "Cannot render template", http.StatusInternalServerError)
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Cannot parse form", http.StatusInternalServerError)
	}

	user := &models.User{
		CernPersonId: r.FormValue("cern-person-id"),
		Username:     r.FormValue("username"),
		FirstName:    r.FormValue("first-name"),
		FamilyName:   r.FormValue("family-name"),
		Email:        r.FormValue("email"),
	}

	if err := h.UserRepo.CreateUser(user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.Templates.ExecuteTemplate(w, "users.html", user)
	if err != nil {
		http.Error(w, "Cannot render template", http.StatusInternalServerError)
	}
}
