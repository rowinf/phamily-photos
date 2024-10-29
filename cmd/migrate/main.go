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

	godotenv.Load()

	arguments := []string{}
	db, err := sql.Open("pgx", os.Getenv("GOOSE_DBSTRING"))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v\n", err)
		}
		fmt.Printf("closing connection...")
		cancel()
	}()

	if err := goose.RunContext(ctx, "up", db, "./cmd/migrate", arguments...); err != nil {
		log.Fatalf("goose %v: %v", "up", err)
		return
	}
}
