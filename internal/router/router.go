package router

import (
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/auth"
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

	// environment
	auth := auth.NewAuth(repoContext.User)
	env := environment.NewEnv(repoContext, auth, baseTemplate, cfg)

	// services
	hasher := service.NewArgon2Hasher()
	ccdbService := service.NewCCDBService(env)
	jalienService := service.NewJAliEnService()
	nnArch := service.NewNNArchService(cfg.NNArchPath)
	fileService := service.NewLocalFileService(cfg.DataDirPath)

	// routes
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))
	mux.Handle("GET /data/", middleware.Chain(http.StripPrefix("/data/", fsData), middleware.NewAuthMw(auth, false)))
	mux.HandleFunc("GET /login", auth.LoginHandler)
	mux.HandleFunc("GET /callback", auth.CallbackHandler)

	// handlers' routes
	handler.InitLandingRoutes(mux, env)
	handler.InitTrainingDatasetRoutes(mux, env, jalienService)
	handler.InitTrainingTaskRoutes(mux, env, ccdbService, fileService, nnArch)
	handler.InitTrainingMachineRoutes(mux, env, hasher)
	handler.InitQueueRoutes(mux, env, fileService, hasher)

	return mux
}
