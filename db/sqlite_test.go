package db_test

import (
	"path/filepath"
	"testing"

	"github.com/profmoo/mnemosyne/db"
	queries "github.com/profmoo/mnemosyne/gen/db"
)

func TestOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "example.sqlite")
	database, err := db.Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	message, err := queries.New(database).GetGreeting(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if message != "Hello, World!" {
		t.Fatalf("greeting = %q, want Hello, World!", message)
	}
	if _, err := database.ExecContext(t.Context(), "UPDATE greetings SET message = ? WHERE id = 1", "Hello, SQLite!"); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := db.Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { reopened.Close() })
	message, err = queries.New(reopened).GetGreeting(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if message != "Hello, SQLite!" {
		t.Errorf("persisted greeting = %q, want Hello, SQLite!", message)
	}
}
