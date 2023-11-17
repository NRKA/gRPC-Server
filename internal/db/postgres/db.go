package postgres

import (
	"context"
	"database/sql"
	"github.com/NRKA/gRPC-Server/internal/db"
	"github.com/NRKA/gRPC-Server/internal/repository"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"log"
	"os"
	"testing"
)

const (
	dbHost     = "DB_HOST"
	dbPort     = "DB_PORT"
	dbUser     = "DB_USER"
	dbPassword = "DB_PASSWORD"
	dbName     = "DB_NAME"
)

type TDB struct {
	DB     repository.DataBaseInterface
	Config db.DatabaseConfig
}

func NewFromEnv() *TDB {
	dbConfig := db.DatabaseConfig{
		Host:     os.Getenv(dbHost),
		Port:     os.Getenv(dbPort),
		User:     os.Getenv(dbUser),
		Password: os.Getenv(dbPassword),
		DBName:   os.Getenv(dbName),
	}
	db, err := db.NewDB(context.Background(), dbConfig)
	if err != nil {
		log.Fatalf("failed to connect to database")
	}
	return &TDB{DB: db, Config: dbConfig}
}

func (d *TDB) SetUp(t *testing.T) {
	db, err := sql.Open("postgres", db.GenerateDsn(d.Config))

	if err != nil {
		log.Fatalf("failed to open database")
	}
	defer db.Close()

	if err = goose.Up(db, "../../db/migrations"); err != nil {
		log.Fatalf("Error setting up the database migrations: %v", err)
	}
}

func (d *TDB) TearDown() {
	db, err := sql.Open("postgres", db.GenerateDsn(d.Config))

	if err != nil {
		log.Fatalf("failed to open database")
	}
	defer db.Close()

	if err = goose.Down(db, "../../db/migrations"); err != nil {
		log.Fatalf("Error setting up the database migrations: %v", err)
	}
}
