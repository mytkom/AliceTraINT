package handler

import (
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/mytkom/AliceTraINT/internal/middleware"
	_ "github.com/thomasdarimont/go-kc-example/session_memory"
)

type UserHandler struct {
	*environment.Env
}

func NewUserHandler(env *environment.Env) *UserHandler {
	return &UserHandler{
		Env: env,
	}
}

func (h *UserHandler) Index(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Users      []models.User
		LoggedUser *models.User
		Title      string
	}

	users, err := h.User.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := TemplateData{
		Users: users,
		Title: "Users List",
	}

	sess := h.GlobalSessions.SessionStart(w, r)
	loggedUserId := sess.Get("loggedUserId")
	if loggedUserId != nil {
		loggedUser, _ := h.User.GetByID(loggedUserId.(uint))
		data.LoggedUser = loggedUser
	}

	err = h.ExecuteTemplate(w, "users_index", data)
	if err != nil {
		http.Error(w, "Cannot render template", http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Cannot parse form", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		CernPersonId: r.FormValue("cern-person-id"),
		Username:     r.FormValue("username"),
		FirstName:    r.FormValue("first-name"),
		FamilyName:   r.FormValue("family-name"),
		Email:        r.FormValue("email"),
	}

	if err := h.User.Create(user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.ExecuteTemplate(w, "users_user", user)
	if err != nil {
		http.Error(w, "Cannot render template", http.StatusInternalServerError)
		return
	}
}

func InitUserRoutes(mux *http.ServeMux, env *environment.Env) {
	uh := NewUserHandler(env)

	authMw := middleware.NewAuthMw(uh.Auth, true)

	mux.Handle("GET /users", middleware.Chain(
		http.HandlerFunc(uh.Index),
		authMw,
	))

	mux.Handle("POST /users", middleware.Chain(
		http.HandlerFunc(uh.CreateUser),
		authMw,
	))
}
