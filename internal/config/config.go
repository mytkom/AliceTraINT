package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Database           DatabaseConfig
	Port               string
	JalienCacheMinutes uint
}

type DatabaseConfig struct {
	Host     string
	Port     uint
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsUint("DB_PORT", 5432),
			User:     getEnv("DB_USER", "user"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "AliceTraINT_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Port:               getEnv("ALICETRAINT_PORT", "8088"),
		JalienCacheMinutes: getEnvAsUint("ALICETRAINT_JALIEN_CACHE_MINUTES", 60),
	}
}

func (dbConfig *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsUint(name string, defaultVal uint) uint {
	if valueStr, exists := os.LookupEnv(name); exists {
		if value, err := strconv.ParseUint(valueStr, 10, 32); err == nil {
			return uint(value)
		}
	}
	return defaultVal
}
