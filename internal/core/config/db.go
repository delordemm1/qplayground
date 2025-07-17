package config

import (
	"context"
	"log"

	"github.com/delordemm1/qplayground/internal/platform"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InitDatabase initializes and returns a PostgreSQL connection pool
func InitDatabase() *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), platform.ENV_DATABASE_URL)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	return pool
}
