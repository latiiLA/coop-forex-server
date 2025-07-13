package migration

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

func InitUp(ctx context.Context, db *mongo.Database) error {
	err := db.CreateCollection(ctx, "roles")
	if err != nil {
		return err
	}

	err = db.CreateCollection(ctx, "users")
	if err != nil {
		return err
	}

	err = db.CreateCollection(ctx, "profiles")
	if err != nil {
		return err
	}

	err = db.CreateCollection(ctx, "countries")
	if err != nil {
		return err
	}

	err = db.CreateCollection(ctx, "districts")
	if err != nil {
		return err
	}

	err = db.CreateCollection(ctx, "branches")
	if err != nil {
		return err
	}

	err = db.CreateCollection(ctx, "processes")
	if err != nil {
		return err
	}

	err = db.CreateCollection(ctx, "subprocesses")
	if err != nil {
		return err
	}

	err = db.CreateCollection(ctx, "departments")
	if err != nil {
		return err
	}
	return nil
}
