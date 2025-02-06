package env

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

var (
	ErrMissingEnvVars = errors.New("missing environment variables")
	ErrLoadingEnv     = errors.New("could not load environment variables")
)

func LoadEnv() (err error) {
	err = godotenv.Load()
	if err != nil {
		return ErrLoadingEnv
	}

	dbURL := os.Getenv("DB_URL")
	ollamaUrl := os.Getenv("OLLAMA_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPwd := os.Getenv("DB_PWD")
	dbName := os.Getenv("DB_NAME")
	port := os.Getenv("PORT")

	if dbURL == "" || dbUser == "" || dbPwd == "" || dbName == "" || port == "" || ollamaUrl == "" {
		return ErrMissingEnvVars
	}

	return nil
}
