package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/config"
	"github.com/mytkom/AliceTraINT/internal/db/migrate"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/mytkom/AliceTraINT/internal/handler"
	"github.com/mytkom/AliceTraINT/internal/middleware"
	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/mytkom/AliceTraINT/internal/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if len(os.Args) != 3 || (os.Args[2] != "e2e" && os.Args[2] != "baseline" && os.Args[2] != "serve") {
		log.Fatal("start this fixture server with test/browser/run.sh")
	}
	if err := serve(os.Args[1], os.Args[2] == "baseline"); err != nil {
		log.Fatal(err)
	}
}

func serve(root string, baseline bool) error {
	// Load configuration
	absRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	if filepath.Dir(absRoot) != "/tmp" || !strings.HasPrefix(filepath.Base(absRoot), "alicetraint-browser.") {
		return fmt.Errorf("expected a private browser fixture directory, got %q", absRoot)
	}
	dataDir := filepath.Join(absRoot, "fixtures")
	resolvedDataDir, err := filepath.EvalSymlinks(dataDir)
	if err != nil {
		return err
	}
	if resolvedDataDir != dataDir {
		return fmt.Errorf("fixture directory resolves outside its launcher path: %s", resolvedDataDir)
	}

	cfg := &config.Config{
		DataDirPath: dataDir,
		DocsDirPath: "docs",
		NNArchPath:  "web/nn_architectures/proposed.json",
	}

	// Initialize GORM with PostgreSQL driver
	dsn := fmt.Sprintf("host=%s port=5432 user=browser_fixture dbname=alicetraint_browser_fixture sslmode=disable", filepath.Join(absRoot, "socket"))
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		return fmt.Errorf("failed to initialize GORM: %w", err)
	}
	var cluster string
	if err := gormDB.Raw("SHOW data_directory").Scan(&cluster).Error; err != nil {
		return err
	}
	if cluster != filepath.Join(absRoot, "pg") {
		return fmt.Errorf("refusing to migrate unexpected PostgreSQL cluster %q", cluster)
	}
	log.Printf("Using fixture PostgreSQL cluster %s and file storage %s", cluster, cfg.DataDirPath)

	// Run database migrations
	migrate.MigrateDB(gormDB)
	if err := migrate.SeedDB(gormDB); err != nil {
		return err
	}
	if baseline {
		if err := seedGalleries(gormDB, dataDir); err != nil {
			return err
		}
	}

	// Setup and start the HTTP server
	repoContext := repository.NewRepositoryContext(gormDB)
	authService := auth.MockAuthService(repoContext.User)
	r, err := fixtureRouter(cfg, repoContext, authService)
	if err != nil {
		return err
	}

	// Add logging middleware
	logMw := middleware.NewLogMw()
	loggedR := logMw(r)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	url := "http://" + listener.Addr().String()
	if err := os.WriteFile(filepath.Join(absRoot, "url.pending"), []byte(url), 0600); err != nil {
		return errors.Join(err, listener.Close())
	}
	if err := os.Rename(filepath.Join(absRoot, "url.pending"), filepath.Join(absRoot, "url")); err != nil {
		return errors.Join(err, listener.Close())
	}
	log.Printf("Browser fixture ready at %s", url)
	return http.Serve(listener, loggedR)
}

func fixtureRouter(cfg *config.Config, repos *repository.RepositoryContext, authService auth.IAuthService) (*http.ServeMux, error) {
	templates := utils.BaseTemplate()
	// Serve the same pinned browser libraries locally to keep the browser offline.
	for _, path := range []string{"web/templates/static/head.html", "web/templates/training-tasks/show.html"} {
		source, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		localize := strings.NewReplacer(
			"https://unpkg.com/htmx.org@1.9.12/dist/", "/test-assets/htmx/",
			"https://unpkg.com/htmx.org@1.9.12", "/test-assets/htmx/htmx.min.js",
			"https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.11.1/", "/test-assets/highlight/",
		)
		_, err = templates.New(filepath.Base(path)).Parse(localize.Replace(string(source)))
		if err != nil {
			return nil, err
		}
	}

	env := environment.NewEnv(repos, authService, templates, cfg)
	files := service.NewLocalFileService(cfg.DataDirPath)
	jalienService, ccdbService, monalisaService := fixtureServices()
	mux := http.NewServeMux()
	serveDirectory := func(prefix, directory string) http.Handler {
		return http.StripPrefix(prefix, http.FileServer(http.Dir(directory)))
	}
	mux.Handle("GET /static/", serveDirectory("/static/", "static"))
	mux.Handle("GET /docs/static/", serveDirectory("/docs/static/", cfg.DocsDirPath))
	mux.Handle("GET /test-assets/htmx/", serveDirectory("/test-assets/htmx/", "node_modules/htmx.org/dist"))
	mux.Handle("GET /test-assets/highlight/", serveDirectory("/test-assets/highlight/", "node_modules/@highlightjs/cdn-assets"))
	mux.Handle("GET "+service.DataURLPrefix, middleware.Chain(
		serveDirectory(service.DataURLPrefix, cfg.DataDirPath), middleware.NewAuthMw(authService, false)))
	mux.HandleFunc("GET /login", authService.LoginHandler)
	mux.HandleFunc("GET /callback", authService.CallbackHandler)

	handler.InitLandingRoutes(mux, env)
	handler.InitDocsRoutes(mux, env)
	handler.InitTrainingDatasetRoutes(mux, env, jalienService, ccdbService, monalisaService, utils.NewCache(30*time.Minute))
	handler.InitTrainingTaskRoutes(mux, env, ccdbService, monalisaService, files, service.NewNNArchService(cfg.NNArchPath))
	hasher := service.NewArgon2Hasher()
	handler.InitTrainingMachineRoutes(mux, env, hasher)
	handler.InitQueueRoutes(mux, env, files, hasher)
	return mux, nil
}
