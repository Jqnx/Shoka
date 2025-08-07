package database

import (
	"Shoka/internal/config"
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

func NewConn(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) *pgx.Conn {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.ErrorContext(ctx, "error establishing db connection")
	}
	return conn.Conn()
}

func NewPool(ctx context.Context, c *config.Config, log *slog.Logger) (*pgxpool.Pool, error) {
	switch {
	case c.Database.DBHost == "":
		return nil, config.ErrNoDBHost
	case c.Database.DBPort == "":
		return nil, config.ErrNoDBPort
	case c.Database.DBDatabase == "":
		return nil, config.ErrNoDBDatabase
	case c.Database.DBUser == "":
		return nil, config.ErrNoDBUser
	case c.Database.DBPassword == "":
		return nil, config.ErrNoDBPassword
	}

	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable TimeZone=%s", c.Database.DBUser, c.Database.DBPassword, c.Database.DBHost, c.Database.DBPort, c.Database.DBDatabase, c.TimeZone)
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.ErrorContext(ctx, "error creating new db pool")
	}

	return pool, nil
}
