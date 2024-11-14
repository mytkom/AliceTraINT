package router

import (
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/config"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/handler"
	"github.com/mytkom/AliceTraINT/internal/utils"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, cfg *config.Config) *http.ServeMux {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))

	// templates
	baseTemplate := utils.BaseTemplate()

	// repositories
	userRepo := repository.NewUserRepository(db)
	trainingDatasetRepo := repository.NewTrainingDatasetRepository(db)
	trainingTaskRepo := repository.NewTrainingTaskRepository(db)
	trainingMachineRepo := repository.NewTrainingMachineRepository(db)

	auth := auth.NewAuth(userRepo)

	// routes
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("GET /login", auth.LoginHandler)
	mux.HandleFunc("GET /callback", auth.CallbackHandler)

	// handlers' routes
	handler.InitLandingRoutes(mux, baseTemplate, auth)
	handler.InitUserRoutes(mux, baseTemplate, userRepo, auth)
	handler.InitTrainingDatasetRoutes(mux, baseTemplate, trainingDatasetRepo, userRepo, auth, cfg.JalienCacheMinutes)
	handler.InitTrainingTaskRoutes(mux, baseTemplate, trainingTaskRepo, trainingDatasetRepo, userRepo, auth)
	handler.InitTrainingMachineRoutes(mux, baseTemplate, trainingMachineRepo, userRepo, auth)

	return mux
}
