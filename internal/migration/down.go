package migration

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
)

func Down(ctx context.Context, db *mongo.Database) error {

	log.Println("🌱 Unseeding roles...")
	if err := unseedRoles(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Unseeding profiles...")
	if err := unseedProfiles(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Unseeding users...")
	if err := unseedUsers(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Unseeding countries...")
	if err := unseedCountries(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Unseeding customer types...")
	if err := unseedCustomerTypes(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Unseeding travel purposes...")
	if err := unseedTravelPurpose(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Unseeding currencies...")
	if err := unseedCurrencies(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Unseeding districts...")
	if err := unseedDistricts(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Unseeding processes...")
	if err := unseedProcesses(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Unseeding subprocesses...")
	if err := unseedSubprocesses(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Unseeding departments...")
	if err := unseedDepartments(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Unseeding branches...")
	if err := unseedBranches(ctx, db); err != nil {
		return err
	}

	log.Println("📁 Running InitDown...")
	if err := InitUp(ctx, db); err != nil {
		return err
	}

	log.Println("✅ All migrations applied successfully.")
	return nil
}
