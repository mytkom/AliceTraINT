package router

import (
	"html/template"
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/handler"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *http.ServeMux {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))

	// templates
	baseTemplate := template.Must(template.ParseFiles("web/templates/base.html"))

	// repositories
	userRepo := repository.NewUserRepository(db)

	auth := auth.NewAuth(userRepo)

	// routes
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("GET /login", auth.LoginHandler)
	mux.HandleFunc("GET /callback", auth.CallbackHandler)

	// handlers' routes
	handler.InitLandingRoutes(mux, baseTemplate, auth)
	handler.InitUserRoutes(mux, baseTemplate, userRepo, auth)
	handler.InitTrainJobRoutes(mux, baseTemplate, auth)

	return mux
}
