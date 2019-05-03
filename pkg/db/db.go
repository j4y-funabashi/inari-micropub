package db

import (
	"database/sql"
	"fmt"

	// register the sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
)

func createDB() string {
	return `
CREATE TABLE IF NOT EXISTS "posts" (
	"id" TEXT PRIMARY KEY,
	"year" INTEGER NOT NULL,
	"sort_key" TEXT NOT NULL,
	"month" INTEGER NOT NULL,
	"data" TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS "idx_posts_year" ON "posts"("year");
CREATE INDEX IF NOT EXISTS "idx_posts_month" ON "posts"("month");

CREATE TABLE IF NOT EXISTS "media" (
	"id" TEXT PRIMARY KEY,
	"year" INTEGER NOT NULL,
	"month" INTEGER NOT NULL,
	"sort_key" TEXT NOT NULL,
	"data" TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS "idx_media_year" ON "media"("year");
CREATE INDEX IF NOT EXISTS "idx_media_month" ON "media"("month");

CREATE TABLE IF NOT EXISTS "events" (
	"id" TEXT PRIMARY KEY,
	"version" TEXT NOT NULL,
	"data" TEXT NOT NULL
);

`
}

func OpenDB() (*sql.DB, error) {
	var db *sql.DB
	var err error
	defer func() {
		if err != nil && db != nil {
			db.Close()
		}
	}()

	dbPath := ""

	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %v", err)
	}

	// ensure DB is provisioned
	_, err = db.Exec(createDB())
	if err != nil {
		return nil, fmt.Errorf("setting up database: %v", err)
	}

	return db, nil
}
