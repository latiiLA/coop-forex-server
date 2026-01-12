package main

import (
	"context"
	"io"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	configs "github.com/latiiLA/coop-forex-server/configs"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/router"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func setupLogger() {
	// Create log directory if not exists
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		err := os.Mkdir("logs", os.ModePerm)
		if err != nil {
			log.Fatal("❌ Could not open log file:", err)
		}
	}

	// Open log file
	file, err := os.OpenFile("logs/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("❌ Could not open log file:", err)
	}

	// Output to both file and stdout
	multiWriter := io.MultiWriter(file, os.Stdout)
	log.SetOutput(multiWriter)
	log.SetFormatter(&log.JSONFormatter{})
	logLevel, err := log.ParseLevel(configs.LogLevel)
	if err != nil {
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)
}

func main() {
	configs.LoadConfig()
	setupLogger()

	ctx := context.TODO()

	clientOptions := options.Client().ApplyURI(configs.MongoURL)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Mongo connection error:", err)
	}

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Warning: failed to disconnect from DB: %v", err)

		} else {
			log.Printf("DB connection closed")
		}
	}()

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal("Ping error:", err)
	}

	log.Info("✅ Connected to DB!")

	db_name := "forex_db"
	if configs.DBName != "" {
		db_name = configs.DBName
	}

	db := client.Database(db_name)
	timeout := configs.Timeout

	// Start the routes
	r := gin.Default()

	// Cors policy
	allowed := configs.AllowedOrigins
	if len(allowed) == 0 {
		logrus.Warn("AllowedOrigins is empty. Falling back to '*'")
		allowed = []string{"*"}
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowed,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Content-Disposition"},
		AllowCredentials: true,
	}))
	// Only trust localhost
	r.SetTrustedProxies(configs.AllowedOrigins)
	r.SetTrustedProxies([]string{"127.0.0.1"})
	r.ForwardedByClientIP = true

	// add logger middleware
	r.Use(middleware.RequestLogger())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RateLimitMiddleware())

	// create an API group
	api := r.Group("/api")

	api.Static("/uploads", "./uploads") // allow upload access

	router.RouterSetup(api, timeout, db)

	if err := r.RunTLS(":8080", configs.CertFile, configs.KeyFile); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
