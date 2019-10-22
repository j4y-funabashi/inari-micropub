package session

import (
	"bytes"
	"database/sql"
	"encoding/json"

	"github.com/j4y_funabashi/inari-micropub/pkg/app"
	uuid "github.com/satori/go.uuid"
)

type SessionStore struct {
	db *sql.DB
}

func NewSessionStore(db *sql.DB) SessionStore {
	return SessionStore{
		db: db,
	}
}

func (ss SessionStore) Create() (app.SessionData, error) {
	uid := uuid.NewV4()
	sessData := app.SessionData{
		Token: uid.String(),
	}

	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(sessData)
	if err != nil {
		return sessData, err
	}

	_, err = ss.db.Exec(
		`INSERT INTO sessions
			(id, data)
			VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		uid.String(),
		buf.String(),
	)
	if err != nil {
		return sessData, err
	}

	return sessData, nil
}
