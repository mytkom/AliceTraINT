package router

import (
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/ccdb"
	"github.com/mytkom/AliceTraINT/internal/config"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/environment"
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
	repoContext := repository.NewRepositoryContext(db)

	// services
	hasher := service.NewArgon2Hasher()
	fileService := service.NewLocalFileService(cfg.DataDirPath)
	queueService := service.NewQueueService(fileService, repoContext, hasher)
	auth := auth.NewAuth(repoContext.User)
	ccdbApi := ccdb.NewCCDBApi(cfg.CCDBBaseURL)

	env := environment.NewEnv(repoContext, auth, baseTemplate, cfg)

	// routes
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))
	mux.Handle("GET /data/", middleware.Chain(http.StripPrefix("/data/", fsData), middleware.NewAuthMw(auth, false)))
	mux.HandleFunc("GET /login", auth.LoginHandler)
	mux.HandleFunc("GET /callback", auth.CallbackHandler)

	// handlers' routes
	handler.InitLandingRoutes(mux, env)
	handler.InitUserRoutes(mux, env)
	handler.InitTrainingDatasetRoutes(mux, env)
	handler.InitTrainingTaskRoutes(mux, env, ccdbApi, fileService)
	handler.InitTrainingMachineRoutes(mux, env)
	handler.InitQueueRoutes(mux, env, queueService)

	return mux
}
