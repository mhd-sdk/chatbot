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

	apiKey := os.Getenv("API_KEY")
	secretKey := os.Getenv("SECRET_KEY")

	marketDataUrl := os.Getenv("MARKET_DATA_URL")
	routingUrl := os.Getenv("ROUTING_URL")

	if apiKey == "" || secretKey == "" || marketDataUrl == "" || routingUrl == "" {
		return ErrMissingEnvVars
	}

	return nil
}
