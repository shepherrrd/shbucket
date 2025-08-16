package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"shbucket/src/Infrastructure/Migrations"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found or error loading .env file:", err)
	}

	fmt.Println("🚀 SHBucket Migration Tool (GoNtext)")
	fmt.Println("===================================")

	if len(os.Args) < 2 {
		showHelp()
		return
	}

	command := os.Args[1]

	migrationCmd, err := migrations.NewMigrationCommands()
	if err != nil {
		fmt.Printf("❌ Failed to initialize migrations: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "migrations:add":
		if len(os.Args) < 3 {
			fmt.Println("❌ Migration name required")
			fmt.Println("💡 Usage: go run . migrations:add InitialCreate")
			os.Exit(1)
		}
		migrationName := os.Args[2]
		if err := migrationCmd.AddMigration(migrationName); err != nil {
			fmt.Printf("❌ Failed to add migration: %v\n", err)
			os.Exit(1)
		}

	case "migrations:update":
		if err := migrationCmd.Update(); err != nil {
			fmt.Printf("❌ Failed to update database: %v\n", err)
			os.Exit(1)
		}

	case "migrations:status":
		if err := migrationCmd.Status(); err != nil {
			fmt.Printf("❌ Failed to get migration status: %v\n", err)
			os.Exit(1)
		}

	case "migrations:list":
		if err := migrationCmd.Status(); err != nil {
			fmt.Printf("❌ Failed to list migrations: %v\n", err)
			os.Exit(1)
		}

	case "migrations:drop":
		if err := migrationCmd.Drop(); err != nil {
			fmt.Printf("❌ Failed to drop database: %v\n", err)
			os.Exit(1)
		}

	case "migrations:rollback":
		steps := 1 // Default to rollback 1 migration
		if len(os.Args) >= 3 {
			// Parse number of steps if provided
			var err error
			if steps, err = parseSteps(os.Args[2]); err != nil {
				fmt.Printf("❌ Invalid steps parameter: %v\n", err)
				fmt.Println("💡 Usage: go run . migrations:rollback [steps]")
				os.Exit(1)
			}
		}
		
		if err := migrationCmd.Rollback(steps); err != nil {
			fmt.Printf("❌ Failed to rollback migrations: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("❌ Unknown command: %s\n", command)
		showHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println()
	fmt.Println("📖 Available Commands:")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Println("🔄 Migration Commands:")
	fmt.Println("  migrations:add <name>       Create a new migration")
	fmt.Println("  migrations:update           Apply migrations to database")
	fmt.Println("  migrations:status           Show current migration status")
	fmt.Println("  migrations:list             List all migration files")
	fmt.Println("  migrations:rollback [steps] Rollback last N migrations (default: 1)")
	fmt.Println("  migrations:drop             Drop all database tables")
	fmt.Println()
	fmt.Println("📋 Examples:")
	fmt.Println("  ./migrations migrations:add InitialCreate")
	fmt.Println("  ./migrations migrations:update")
	fmt.Println("  ./migrations migrations:status")
	fmt.Println("  ./migrations migrations:rollback")
	fmt.Println("  ./migrations migrations:rollback 2")
	fmt.Println()
	fmt.Println("📋 Or using go run:")
	fmt.Println("  go run ./cmd/migrations migrations:add InitialCreate")
	fmt.Println("  go run ./cmd/migrations migrations:update")
	fmt.Println("  go run ./cmd/migrations migrations:status")
	fmt.Println("  go run ./cmd/migrations migrations:rollback")
	fmt.Println("  go run ./cmd/migrations migrations:rollback 2")
	fmt.Println()
	fmt.Println("⚙️  Environment:")
	fmt.Println("  Set DATABASE_URL environment variable or it will default to:")
	fmt.Println("  postgres://postgres@localhost:5432/shbucket?sslmode=disable")
	fmt.Println()
	fmt.Println("✨ Features:")
	fmt.Println("  • Model snapshots (like EF Core)")
	fmt.Println("  • Change detection")
	fmt.Println("  • LINQ-style queries")
	fmt.Println("  • Automatic migration generation")

	// Show current DATABASE_URL if set
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		// Mask password for security
		maskedURL := maskPassword(dbURL)
		fmt.Printf("\n🔗 Current DATABASE_URL: %s\n", maskedURL)
	}
}

// maskPassword masks the password in a database URL for display
func maskPassword(url string) string {
	// Simple password masking - replace password with ***
	if strings.Contains(url, "://") && strings.Contains(url, ":") && strings.Contains(url, "@") {
		parts := strings.Split(url, "://")
		if len(parts) == 2 {
			scheme := parts[0]
			remaining := parts[1]
			
			atIndex := strings.Index(remaining, "@")
			if atIndex > 0 {
				userPass := remaining[:atIndex]
				hostDb := remaining[atIndex:]
				
				if strings.Contains(userPass, ":") {
					userParts := strings.Split(userPass, ":")
					if len(userParts) >= 2 {
						user := userParts[0]
						return fmt.Sprintf("%s://%s:***%s", scheme, user, hostDb)
					}
				}
			}
		}
	}
	return url
}

// parseSteps parses the number of rollback steps from command line argument
func parseSteps(stepStr string) (int, error) {
	steps, err := strconv.Atoi(stepStr)
	if err != nil {
		return 0, fmt.Errorf("steps must be a valid number")
	}
	if steps < 1 {
		return 0, fmt.Errorf("steps must be at least 1")
	}
	return steps, nil
}