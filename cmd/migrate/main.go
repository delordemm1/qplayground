package main

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"strings"

	// "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	// _ "github.com/lib/pq"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/db/main.go [create <migration-name> | up | down | status | version]")
		os.Exit(1)
	}

	command := os.Args[1]

	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Configure goose
	goose.SetDialect("pgx")
	goose.SetBaseFS(embedMigrations)
	// conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
	// 	os.Exit(1)
	// }
	// defer conn.Close(context.Background())

	switch command {
	case "create":
		if len(os.Args) < 3 {
			log.Fatal("Please provide a migration name, e.g., `go run cmd/migrate/main.go create add_users_table`")
		}
		migrationName := os.Args[2]
		migrationName = strings.ToLower(migrationName)
		migrationName = strings.ReplaceAll(migrationName, "-", "_")
		// Ensure migration name is valid (must be at least 1 character, no spaces, and only alphanumeric + underscores)
		if len(migrationName) < 1 {
			log.Fatal("Migration name must be at least 1 character")
		}
		if migrationName[0] == '_' {
			log.Fatal("Migration name cannot start with an underscore")
		}
		for _, char := range migrationName {
			if !(('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || ('0' <= char && char <= '9')) && char != '_' {
				log.Fatalf("Migration name contains an invalid character: '%c'. Only alphanumeric characters and underscores are allowed.", char)
			}
		}
		log.Printf("MigrationName: %s", migrationName)

		if err := goose.Create(db, "cmd/migrate/migrations", migrationName, "sql"); err != nil {
			log.Fatal(err)
		}
	case "up":
		if err := goose.Up(db, "migrations"); err != nil {
			log.Fatal(err)
		}
	case "down":
		if err := goose.Down(db, "migrations"); err != nil {
			log.Fatal(err)
		}
	case "status":
		if err := goose.Status(db, "migrations"); err != nil {
			log.Fatal(err)
		}
	case "version":
		if err := goose.Version(db, "migrations"); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Println("Unknown command. Use up, down, status, version")
		os.Exit(1)
	}
}
