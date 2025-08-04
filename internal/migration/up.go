package migration

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
)

func Up(ctx context.Context, db *mongo.Database) error {
	log.Println("📁 Running InitUp...")
	if err := InitUp(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Seeding roles...")
	if err := seedRoles(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Seeding profiles...")
	if err := seedProfiles(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Seeding users...")
	if err := seedUsers(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Seeding countries...")
	if err := seedCountries(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Seeding customer types...")
	if err := seedCustomerTypes(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Seeding travel purpose...")
	if err := seedTravelPurpose(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Seeding currencies...")
	if err := seedCurrencies(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Seeding districts...")
	if err := seedDistricts(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Seeding processes...")
	if err := seedProcesses(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Seeding subprocesses...")
	if err := seedSubprocesses(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Seeding departments...")
	if err := seedDepartments(ctx, db); err != nil {
		return err
	}

	log.Println("🌱 Seeding branches...")
	if err := seedBranches(ctx, db); err != nil {
		return err
	}

	log.Println("✅ All migrations applied successfully.")
	return nil
}
