package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mytkom/AliceTraINT/internal"
	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/config"
	"github.com/mytkom/AliceTraINT/internal/db/migrate"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize GORM with PostgreSQL driver
	dsn := cfg.Database.ConnectionString()
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Info),
		TranslateError: true,
	})
	if err != nil {
		log.Fatalf("Failed to initialize GORM: %v", err)
	}

	// Run database migrations
	migrate.MigrateDB(gormDB)

	// Setup and start the HTTP server
	repoContext := repository.NewRepositoryContext(gormDB)
	authService := auth.NewAuthService(repoContext.User)
	r := internal.NewRouter(cfg, repoContext, authService)

	// Add logging middleware
	logMw := middleware.NewLogMw()
	loggedR := logMw(r)

	portString := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s\n", portString)
	log.Fatal(http.ListenAndServe(portString, loggedR))
}
