package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	var (
		direction = flag.String("direction", "up", "Migration direction: up or down")
		dbURL     = flag.String("db", "", "Database URL (or use DATABASE_URL env var)")
	)
	flag.Parse()

	// Get database URL from flag or environment
	databaseURL := *dbURL
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if databaseURL == "" {
		log.Fatal("Database URL not provided. Use -db flag or DATABASE_URL environment variable")
	}

	// Parse command
	command := "up"
	if flag.NArg() > 0 {
		command = flag.Arg(0)
	}
	if *direction == "down" {
		command = "down"
	}

	// Connect to database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// Get migrations directory
	migrationsDir := "migrations"
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		// Check if running from a different directory
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			// Try parent directory
			migrationsDir = "../migrations"
			if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
				// Try from root
				migrationsDir = "./migrations"
			}
		}
	}

	// Run migrations
	switch command {
	case "up":
		if err := migrateUp(db, migrationsDir); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Println("Migrations completed successfully")
	case "down":
		if err := migrateDown(db, migrationsDir); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("Rollback completed successfully")
	case "status":
		if err := showStatus(db, migrationsDir); err != nil {
			log.Fatalf("Failed to show status: %v", err)
		}
	default:
		log.Fatalf("Unknown command: %s. Use 'up', 'down', or 'status'", command)
	}
}

func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.Exec(query)
	return err
}

func migrateUp(db *sql.DB, migrationsDir string) error {
	// Get all migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	sort.Strings(files)

	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply pending migrations
	for _, file := range files {
		version := filepath.Base(file)

		// Skip if already applied
		if applied[version] {
			log.Printf("Skipping %s (already applied)", version)
			continue
		}

		// Read migration file
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// Begin transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Execute migration
		log.Printf("Applying %s...", version)
		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to apply migration %s: %w", version, err)
		}

		// Record migration
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", version, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", version, err)
		}

		log.Printf("✓ Applied %s", version)
	}

	return nil
}

func migrateDown(db *sql.DB, migrationsDir string) error {
	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		log.Println("No migrations to rollback")
		return nil
	}

	// Get latest migration
	var latest string
	for version := range applied {
		if version > latest {
			latest = version
		}
	}

	log.Printf("Rolling back %s...", latest)

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Remove migration record
	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = $1", latest); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %w", err)
	}

	log.Printf("✓ Rolled back %s", latest)
	log.Println("Note: This only removes the migration record. Manual SQL may be needed to undo schema changes.")

	return nil
}

func showStatus(db *sql.DB, migrationsDir string) error {
	// Get all migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	sort.Strings(files)

	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Show status
	fmt.Println("\nMigration Status:")
	fmt.Println("==================")
	for _, file := range files {
		version := filepath.Base(file)
		status := "[ ]"
		if applied[version] {
			status = "[✓]"
		}
		fmt.Printf("%s %s\n", status, version)
	}
	fmt.Printf("\nTotal: %d, Applied: %d, Pending: %d\n", len(files), len(applied), len(files)-len(applied))

	return nil
}

func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}
