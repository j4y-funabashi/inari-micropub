package session

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"strings"

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

func (ss SessionStore) Save(sessData app.SessionData) error {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(sessData)
	if err != nil {
		return err
	}

	_, err = ss.db.Exec(
		`UPDATE sessions SET data = $1 WHERE id = $2`,
		buf.String(),
		sessData.Token,
	)
	return err
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

func (ss SessionStore) Fetch(sessionID string) (app.SessionData, error) {
	out := app.SessionData{}
	rows, err := ss.db.Query(
		`SELECT data FROM sessions WHERE id = $1`,
		sessionID,
	)
	if err != nil {
		return out, err
	}
	defer rows.Close()

	var mfJSON string
	for rows.Next() {
		err := rows.Scan(&mfJSON)
		if err != nil {
			return out, err
		}
		err = json.NewDecoder(strings.NewReader(mfJSON)).Decode(&out)
		if err != nil {
			return out, err
		}
	}
	return out, nil
}
