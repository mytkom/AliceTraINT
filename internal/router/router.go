package router

import (
	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/handler"
	"gorm.io/gorm"
	"html/template"
	"net/http"
)

func NewRouter(db *gorm.DB) *http.ServeMux {
	mux := http.NewServeMux()

	templates := template.Must(template.ParseGlob("web/templates/*.html"))

	userRepo := repository.NewUserRepository(db)
	auth := auth.NewAuth(userRepo)

	userHandler := handler.NewUserHandler(userRepo, templates, auth.GlobalSessions)

	mux.HandleFunc("GET /", userHandler.Index)
	mux.HandleFunc("POST /users", userHandler.CreateUser)
	mux.HandleFunc("GET /login", auth.LoginHandler)
	mux.HandleFunc("GET /callback", auth.CallbackHandler)

	return mux
}
