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
	JalienHost         string
	JalienPort         string
	JalienCertCADir    string
	CCDBBaseURL        string
	CCDBUploadSubdir   string
	CertPath           string
	KeyPath            string
	DataDirPath        string
	NNArchPath         string
	DocsDirPath        string
}

type DatabaseConfig struct {
	Host            string
	Port            uint
	User            string
	Password        string
	DBName          string
	SSLMode         string
	SSLRootCertPath string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	const defaultJalienHost = "alice-jcentral.cern.ch"
	const defaultJalienPort = "8097"

	return &Config{
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsUint("DB_PORT", 5432),
			User:            getEnv("DB_USER", "user"),
			Password:        getEnv("DB_PASSWORD", "password"),
			DBName:          getEnv("DB_NAME", "AliceTraINT_db"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			SSLRootCertPath: getEnv("DB_SSL_CERT_PATH", ""),
		},
		Port:               getEnv("ALICETRAINT_PORT", "8088"),
		JalienCacheMinutes: getEnvAsUint("ALICETRAINT_JALIEN_CACHE_MINUTES", 60),
		JalienHost:         getEnv("JALIEN_HOST", defaultJalienHost),
		JalienPort:         getEnv("JALIEN_WSPORT", defaultJalienPort),
		JalienCertCADir:    getEnv("JALIEN_CERT_CA_DIR", ""),
		CCDBBaseURL:        getEnv("CCDB_URL", "http://ccdb-test.cern.ch:8080"),
		CCDBUploadSubdir:   getEnv("CCDB_UPLOAD_SUBDIR", "/Users/m/mmytkows"),
		CertPath:           getEnv("GRID_CERT_PATH", ""),
		KeyPath:            getEnv("GRID_KEY_PATH", ""),
		DataDirPath:        getEnv("ALICETRAINT_DATA_DIR_PATH", "data"),
		NNArchPath:         getEnv("ALICETRAINT_NN_ARCH_DIR", "web/nn_architectures/proposed.json"),
		DocsDirPath:        getEnv("ALICETRAINT_DOCS_DIR_PATH", "docs"),
	}
}

func (dbConfig *DatabaseConfig) ConnectionString() string {
	if dbConfig.SSLMode == "disabled" {
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.SSLMode,
		)
	} else {
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s sslrootcert=%s",
			dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.SSLMode, dbConfig.SSLRootCertPath,
		)
	}
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
