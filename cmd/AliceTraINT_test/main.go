package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

func createDatabase(cfg *config.Config) func() {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable database=postgres", cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password)
	DB, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	createDatabaseCommand := fmt.Sprintf(`CREATE DATABASE "%s"`, cfg.Database.DBName)
	deleteDatabaseCommand := fmt.Sprintf(`DROP DATABASE "%s" WITH (FORCE)`, cfg.Database.DBName)

	err := DB.Exec(createDatabaseCommand).Error
	if err != nil {
		if err := DB.Exec(deleteDatabaseCommand).Error; err != nil {
			log.Panic(err)
		}
		if err := DB.Exec(createDatabaseCommand).Error; err != nil {
			log.Panic(err)
		}
	}

	return func() {
		if err := DB.Exec(deleteDatabaseCommand).Error; err != nil {
			log.Panic(err)
		}
	}
}

func gracefulShutdown(cleanup func()) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit
	fmt.Println("interrupted:", s)
	cleanup()
	os.Exit(0)
}

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Suffix DB Name with test (it will be created and deleted )
	cfg.Database.DBName = fmt.Sprintf("%s_test", cfg.Database.DBName)
	deleteDatabase := createDatabase(cfg)
	defer deleteDatabase()
	go gracefulShutdown(deleteDatabase)

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
	err = migrate.SeedDB(gormDB)
	if err != nil {
		log.Println(err.Error())
		return
	}

	// Setup and start the HTTP server
	repoContext := repository.NewRepositoryContext(gormDB)
	authService := auth.MockAuthService(repoContext.User)
	r := internal.NewRouter(cfg, repoContext, authService)

	// Add logging middleware
	logMw := middleware.NewLogMw()
	loggedR := logMw(r)

	portString := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting server on %s\n", portString)
	log.Fatal(http.ListenAndServe(portString, loggedR))
}
