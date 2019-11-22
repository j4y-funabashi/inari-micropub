package session_test

import (
	"os"
	"testing"

	"github.com/j4y_funabashi/inari-micropub/pkg/db"
	"github.com/j4y_funabashi/inari-micropub/pkg/session"
)

func TestCreateSession(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.SkipNow()
	}

	sqlClient, err := db.OpenDB()
	if err != nil {
		t.Fatalf("failed to open DB: %s", err.Error())
	}
	sut := session.NewSessionStore(sqlClient)

	// act
	sessData, err := sut.Create()
	if err != nil {
		t.Fatalf("failed to create session: %s", err.Error())
	}

	t.Errorf("%+v", sessData)
}
