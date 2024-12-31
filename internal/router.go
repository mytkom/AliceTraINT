package internal

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
)

func NewRouter(cfg *config.Config, repoContext *repository.RepositoryContext, authService auth.IAuthService) *http.ServeMux {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))

	// templates
	baseTemplate := utils.BaseTemplate()

	env := environment.NewEnv(repoContext, authService, baseTemplate, cfg)

	// services
	hasher := service.NewArgon2Hasher()
	ccdbService := service.NewCCDBService(env)
	jalienService := service.NewJAliEnService()
	nnArch := service.NewNNArchService(cfg.NNArchPath)
	// local file storage
	fileService := service.NewLocalFileService(cfg.DataDirPath)
	fsData := http.FileServer(http.Dir("data"))

	// routes
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))
	mux.Handle("GET /data/", middleware.Chain(http.StripPrefix("/data/", fsData), middleware.NewAuthMw(authService, false)))
	mux.HandleFunc("GET /login", authService.LoginHandler)
	mux.HandleFunc("GET /callback", authService.CallbackHandler)

	// handlers' routes
	handler.InitLandingRoutes(mux, env)
	handler.InitTrainingDatasetRoutes(mux, env, jalienService)
	handler.InitTrainingTaskRoutes(mux, env, ccdbService, jalienService, fileService, nnArch)
	handler.InitTrainingMachineRoutes(mux, env, hasher)
	handler.InitQueueRoutes(mux, env, fileService, hasher)

	return mux
}
