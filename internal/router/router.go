package router

import (
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/ccdb"
	"github.com/mytkom/AliceTraINT/internal/config"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/handler"
	"github.com/mytkom/AliceTraINT/internal/middleware"
	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/mytkom/AliceTraINT/internal/utils"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, cfg *config.Config) *http.ServeMux {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))
	fsData := http.FileServer(http.Dir("data"))

	// templates
	baseTemplate := utils.BaseTemplate()

	// repositories
	userRepo := repository.NewUserRepository(db)
	trainingDatasetRepo := repository.NewTrainingDatasetRepository(db)
	trainingTaskRepo := repository.NewTrainingTaskRepository(db)
	trainingMachineRepo := repository.NewTrainingMachineRepository(db)
	trainingTaskResultRepo := repository.NewTrainingTaskResultRepository(db)

	// services
	fileService := service.NewLocalFileService(cfg.DataDirPath)
	auth := auth.NewAuth(userRepo)
	ccdbApi := ccdb.NewCCDBApi(cfg.CCDBBaseURL)

	// routes
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))
	mux.Handle("GET /data/", middleware.Chain(http.StripPrefix("/data/", fsData), middleware.NewAuthMw(auth, false)))
	mux.HandleFunc("GET /login", auth.LoginHandler)
	mux.HandleFunc("GET /callback", auth.CallbackHandler)

	// handlers' routes
	handler.InitLandingRoutes(mux, baseTemplate, auth)
	handler.InitUserRoutes(mux, baseTemplate, userRepo, auth)
	handler.InitTrainingDatasetRoutes(mux, baseTemplate, trainingDatasetRepo, userRepo, auth, cfg.JalienCacheMinutes)
	handler.InitTrainingTaskRoutes(mux, baseTemplate, trainingTaskRepo, trainingDatasetRepo, trainingTaskResultRepo, userRepo, auth, ccdbApi, cfg, fileService)
	handler.InitTrainingMachineRoutes(mux, baseTemplate, trainingMachineRepo, userRepo, auth)
	handler.InitQueryRoutes(mux, trainingMachineRepo, trainingTaskRepo, trainingTaskResultRepo, fileService)

	return mux
}
