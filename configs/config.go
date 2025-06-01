package configs

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var (
	JwtSecret string
	MongoURL  string
	Timeout   time.Duration
)

func LoadConfig() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found or couldn't load it, relying on environment variables")
	}

	MongoURL = os.Getenv("MONGO_URL")
	if MongoURL == "" {
		log.Fatal("Mongo url is required but not set")
	}

	JwtSecret = os.Getenv("JWT_SECRET")
	if JwtSecret == "" {
		log.Fatal("JWT_SECRET is required but not set")
	}

	timeoutStr := os.Getenv("APP_TIMEOUT")
	if timeoutStr == "" {
		log.Fatalf("APP_TIMEOUT is required but not set")
	}

	Timeout, err = time.ParseDuration(timeoutStr)
	if err != nil {
		log.Fatalf("Invalid APP_TIMEOUT format: %v", err)
	}
}
