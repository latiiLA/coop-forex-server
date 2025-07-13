package main

import (
	"context"
	"log"
	"os"

	"github.com/latiiLA/coop-forex-server/configs"
	"github.com/latiiLA/coop-forex-server/internal/migration"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func RunMigrations() error {
	ctx := context.Background()
	configs.LoadConfig()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(configs.MongoURL))
	if err != nil {
		log.Fatal("❌ DB connect error:", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database("coop_forex_db")

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "up":
		if err := migration.Up(ctx, db); err != nil {
			log.Fatal("❌ Migration up failed:", err)
		}
		return err
	case "down":
		if err := migration.Down(ctx, db); err != nil {
			log.Fatal("❌ Migration down failed:", err)
		}
		return err
	default:
		log.Fatal("❌ Unknown command. Use 'up' or 'down'")
		return err
	}
}
