package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"workflow-code-test/api/pkg/db"
)

func main() {
	var (
		migrationsPath = flag.String("migrations-path", "./migrations", "Path to migrations directory")
		databaseURL    = flag.String("database-url", "", "Database URL (can also use DATABASE_URL env var)")
		command        = flag.String("command", "up", "Migration command: up, down, version")
		seedData       = flag.Bool("seed", false, "Seed test data after running migrations")
	)
	flag.Parse()

	// Use environment variable if database URL not provided
	if *databaseURL == "" {
		*databaseURL = os.Getenv("DATABASE_URL")
	}

	if *databaseURL == "" {
		log.Fatal("Database URL must be provided via -database-url flag or DATABASE_URL environment variable")
	}

	// Connect to database
	dbConfig := db.DefaultConfig()
	dbConfig.URI = *databaseURL

	if err := db.Connect(dbConfig); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Disconnect()

	switch *command {
	case "up":
		if err := db.RunMigrations(*databaseURL, *migrationsPath); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		fmt.Println("Migrations completed successfully")

		// Seed test data if requested
		if *seedData {
			fmt.Println("Seeding test data...")
			if err := SeedTestData(); err != nil {
				log.Fatalf("Failed to seed test data: %v", err)
			}
			fmt.Println("Test data seeded successfully")
		}

	case "version":
		version, dirty, err := db.GetMigrationVersion(*databaseURL, *migrationsPath)
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		if dirty {
			fmt.Printf("Current migration version: %d (dirty)\n", version)
		} else {
			fmt.Printf("Current migration version: %d\n", version)
		}

	default:
		log.Fatalf("Unknown command: %s. Available commands: up, version", *command)
	}
}
