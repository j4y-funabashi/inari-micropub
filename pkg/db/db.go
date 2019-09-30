package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
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

CREATE TABLE IF NOT EXISTS "media_published" (
	"id" TEXT PRIMARY KEY
);

`
}

func OpenDB() (*sql.DB, error) {

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, fmt.Errorf("failed opening database: %s", err.Error())
	}

	// ensure DB is provisioned
	_, err = db.Exec(createDB())
	if err != nil {
		return nil, fmt.Errorf("setting up database: %v", err)
	}

	return db, nil
}
