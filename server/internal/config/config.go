package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	App     AppConfig
	Storage StorageConfig
}

type AppConfig struct {
	GrpcHost string
	GrpcPort int
}
type StorageConfig struct {
	FilesPath string
}

func Init() *Config {

	grpcPort, err := strconv.Atoi(getVar("GRPC_PORT", "50050"))

	if err != nil {
		log.Fatalf("GRPC_PORT doesn't look like an integer: %s", err)
	}

	filesPath := getVar("FILE_STORAGE", "files")

	if err = os.MkdirAll(filesPath, 0777); err != nil {
		log.Fatalf("File storage dont create: %s", err)
	}

	return &Config{
		App: AppConfig{
			GrpcHost: getVar("GRPC_HOST", "0.0.0.0"),
			GrpcPort: grpcPort,
		},

		Storage: StorageConfig{
			FilesPath: filesPath,
		},
	}
}

func getVar(key string, fallback string) string {

	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
