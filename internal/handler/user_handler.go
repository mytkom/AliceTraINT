package handler

import (
	"html/template"
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/middleware"
	_ "github.com/thomasdarimont/go-kc-example/session_memory"
)

type UserHandler struct {
	UserRepo repository.UserRepository
	Template *template.Template
	Auth     *auth.Auth
}

func NewUserHandler(baseTemplate *template.Template, userRepo repository.UserRepository, auth *auth.Auth) *UserHandler {
	return &UserHandler{
		UserRepo: userRepo,
		Template: baseTemplate,
		Auth:     auth,
	}
}

func (h *UserHandler) Index(w http.ResponseWriter, r *http.Request) {
	users, err := h.UserRepo.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Users      []models.User
		LoggedUser *models.User
		Title      string
	}{Users: users, Title: "Users List"}

	sess := h.Auth.GlobalSessions.SessionStart(w, r)
	loggedUserId := sess.Get("loggedUserId")
	if loggedUserId != nil {
		loggedUser, _ := h.UserRepo.GetByID(loggedUserId.(uint))
		data.LoggedUser = loggedUser
	}

	err = h.Template.ExecuteTemplate(w, "users_index", data)
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

	if err := h.UserRepo.Create(user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.Template.ExecuteTemplate(w, "users_user", user)
	if err != nil {
		http.Error(w, "Cannot render template", http.StatusInternalServerError)
	}
}

func InitUserRoutes(mux *http.ServeMux, baseTemplate *template.Template, userRepo repository.UserRepository, auth *auth.Auth) {
	uh := NewUserHandler(baseTemplate, userRepo, auth)

	authMw := middleware.NewAuthMw(auth)

	mux.Handle("GET /users", middleware.Chain(
		http.HandlerFunc(uh.Index),
		authMw,
	))

	mux.Handle("POST /users", middleware.Chain(
		http.HandlerFunc(uh.CreateUser),
		authMw,
	))
}
