package database

import (
	"Shoka/internal/config"
	"Shoka/internal/logger"
	"context"
	"log"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	l    logger.Logger
	host string
	port string
	db   = config.Database{
		DBDatabase: "database",
		DBUser:     "user",
		DBPassword: "password",
	}
	conf = config.Config{
		Database: db,
	}
)

func mustStartPostgresContainer() (func(context.Context, ...testcontainers.TerminateOption) error, error) {
	dbContainer, err := postgres.Run(
		context.Background(),
		"postgres:latest",
		postgres.WithDatabase(db.DBDatabase),
		postgres.WithUsername(db.DBUser),
		postgres.WithPassword(db.DBPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}

	dbHost, err := dbContainer.Host(context.Background())
	if err != nil {
		return dbContainer.Terminate, err
	}

	dbPort, err := dbContainer.MappedPort(context.Background(), "5432/tcp")
	if err != nil {
		return dbContainer.Terminate, err
	}

	host = dbHost
	port = dbPort.Port()

	return dbContainer.Terminate, err
}

func TestMain(m *testing.M) {
	teardown, err := mustStartPostgresContainer()
	if err != nil {
		log.Fatalf("could not start postgres container: %v", err)
	}

	logger := logger.NewLogger("")
	l = logger
	m.Run()

	if teardown != nil && teardown(context.Background()) != nil {
		log.Fatalf("could not teardown postgres container: %v", err)
	}
}

func TestNew(t *testing.T) {
	ctx := context.Background()
	conf.Database.DBHost = host
	conf.Database.DBPort = port

	_, err := NewPool(ctx, &conf, l)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Log("connected to database.")
}
