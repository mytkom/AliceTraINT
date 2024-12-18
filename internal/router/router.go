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
	repoContext := repository.NewRepositoryContext(
		repository.NewUserRepository(db),
		repository.NewTrainingMachineRepository(db),
		repository.NewTrainingDatasetRepository(db),
		repository.NewTrainingTaskRepository(db),
		repository.NewTrainingTaskResultRepository(db),
	)

	// services
	hasher := service.NewArgon2Hasher()
	fileService := service.NewLocalFileService(cfg.DataDirPath)
	queueService := service.NewQueueService(fileService, repoContext, hasher)
	auth := auth.NewAuth(repoContext.User)
	ccdbApi := ccdb.NewCCDBApi(cfg.CCDBBaseURL)

	// routes
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))
	mux.Handle("GET /data/", middleware.Chain(http.StripPrefix("/data/", fsData), middleware.NewAuthMw(auth, false)))
	mux.HandleFunc("GET /login", auth.LoginHandler)
	mux.HandleFunc("GET /callback", auth.CallbackHandler)

	// handlers' routes
	handler.InitLandingRoutes(mux, baseTemplate, auth)
	handler.InitUserRoutes(mux, baseTemplate, repoContext.User, auth)
	handler.InitTrainingDatasetRoutes(mux, baseTemplate, repoContext.TrainingDataset, repoContext.User, auth, cfg.JalienCacheMinutes)
	handler.InitTrainingTaskRoutes(mux, baseTemplate, repoContext.TrainingTask, repoContext.TrainingDataset, repoContext.TrainingTaskResult, repoContext.User, auth, ccdbApi, cfg, fileService)
	handler.InitTrainingMachineRoutes(mux, baseTemplate, repoContext.TrainingMachine, repoContext.User, auth)
	handler.InitQueueRoutes(mux, repoContext, queueService)

	return mux
}
