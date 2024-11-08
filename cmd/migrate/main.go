package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

func main() {

	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	arguments := []string{}
	db, err := sql.Open("pgx", os.Getenv("GOOSE_DBSTRING"))

	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v\n", err)
		}
		fmt.Printf("Database connection closed.")
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}

	if err := goose.RunContext(ctx, "up", db, os.Getenv("GOOSE_MIGRATION_DIR"), arguments...); err != nil {
		log.Fatalf("goose %v: %v", "up", err)
	}
}
