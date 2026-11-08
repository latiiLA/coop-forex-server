package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"

	configs "github.com/latiiLA/coop-forex-server/configs"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/router"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/middleware"
	"github.com/latiiLA/coop-forex-server/internal/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/sirupsen/logrus"
)

func setupLogger() {
	// Create log directory if not exists
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		err := os.Mkdir("logs", os.ModePerm)
		if err != nil {
			logrus.Fatal("❌ Could not open log file:", err)
		}
	}

	// Lumberjack handles rotation
	logrus.SetOutput(&lumberjack.Logger{
		Filename:   "logs/server.log",
		MaxSize:    100,
		MaxBackups: 100,
		MaxAge:     100,
		Compress:   true,
	})

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logLevel, err := logrus.ParseLevel(configs.LogLevel)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logrus.SetLevel(logLevel)
}

func main() {
	configs.LoadConfig()
	setupLogger()

	ctx := context.TODO()

	clientOptions := options.Client().ApplyURI(configs.MongoURL)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logrus.Fatal("Mongo connection error:", err)
	}

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logrus.Printf("Warning: failed to disconnect from DB: %v", err)
		} else {
			logrus.Printf("DB connection closed")
		}
	}()

	if err = client.Ping(ctx, nil); err != nil {
		logrus.Fatal("Ping error:", err)
	}

	logrus.Info("✅ Connected to DB!")

	db_name := "forex_db"
	if configs.DBName != "" {
		db_name = configs.DBName
	}

	db := client.Database(db_name)
	timeout := configs.Timeout

	if mode := os.Getenv("GIN_MODE"); mode == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Start the routes
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     configs.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Content-Disposition"},
		ExposeHeaders:    []string{"Content-Disposition", "X-Receipt-Number"},
		AllowCredentials: true,
	}))

	if err := r.SetTrustedProxies(configs.TrustedProxies); err != nil {
		logrus.Fatalf("Invalid TRUSTED_PROXIES: %v", err)
	}
	r.ForwardedByClientIP = true

	// add logger middleware
	r.Use(middleware.RequestLogger())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RateLimitMiddleware())

	// create an API group
	api := r.Group("/api")

	fileRepo := repository.NewFileRepository(db)
	requestRepo := repository.NewRequestRepository(db)
	userRepo := repository.NewUserRepository(db)

	uploads := api.Group("/uploads")
	uploads.Use(middleware.JwtAuthMiddleware(configs.JwtSecret))
	uploads.GET("/*filepath", middleware.NewServeProtectedUpload(fileRepo, requestRepo, userRepo))

	router.RouterSetup(api, timeout, db)

	// Health check endpoint
	api.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := client.Ping(ctx, nil); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "unhealthy",
				"message": "database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "service is running",
		})
	})

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		logrus.Info("Server starting on :8080")
		if err := srv.ListenAndServeTLS(configs.CertFile, configs.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logrus.Fatalf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server gracefully stopped")
}
