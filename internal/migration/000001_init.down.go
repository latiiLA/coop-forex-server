package migration

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
)

func InitDown(ctx context.Context, db *mongo.Database) error {
	collections := []string{"users", "roles", "profiles", "countries", "currencies", "districts", "branches", "processes", "subprocesses", "departments"}

	for _, col := range collections {
		if err := db.Collection(col).Drop(ctx); err != nil {
			log.Printf("⚠️ Failed to drop collection '%s': %v", col, err)
		} else {
			log.Printf("🗑️ Dropped collection: %s", col)
		}
	}

	log.Println("✅ Down migration completed.")
	return nil
}
