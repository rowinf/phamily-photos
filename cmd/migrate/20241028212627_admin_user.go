package main

import (
	"context"
	"database/sql"
	"os"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	goose.AddMigrationContext(upAdminUser, downAdminUser)
}

func upAdminUser(ctx context.Context, tx *sql.Tx) error {
	familyName := os.Getenv("INITIAL_FAMILY_NAME")
	adminUserName := os.Getenv("INITIAL_USER_NAME")
	password := os.Getenv("INITIAL_USER_PASSWORD")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	userId := uuid.New().String()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
	INSERT INTO families (id, created_at, updated_at, name, description) VALUES (1, NOW(), NOW(), $1, $2)`,
		familyName, "Admin Family")
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
	INSERT INTO users (id, created_at, updated_at, name, apikey, password, family_id)
	VALUES ($1, NOW(), NOW(), $2, encode(sha256(random()::text::bytea), 'hex'), $3, 1)`,
		userId, adminUserName, string(hashedPassword))
	if err != nil {
		return err
	}
	return nil
}

func downAdminUser(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
