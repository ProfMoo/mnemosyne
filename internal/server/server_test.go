package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/profmoo/mnemosyne/db"
	"github.com/profmoo/mnemosyne/gen/api"
	"github.com/profmoo/mnemosyne/internal/server"
)

func TestHello(t *testing.T) {
	database, err := db.Open(t.Context(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	recorder := httptest.NewRecorder()
	server.NewHandler(database).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/hello", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	var response api.HelloResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Message != "Hello, World!" {
		t.Errorf("message = %q, want Hello, World!", response.Message)
	}
}

func TestRouting(t *testing.T) {
	database, err := db.Open(t.Context(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	handler := server.NewHandler(database)
	for _, test := range []struct {
		name   string
		method string
		path   string
		status int
	}{
		{"unknown path", http.MethodGet, "/missing", http.StatusNotFound},
		{"unsupported method", http.MethodPost, "/hello", http.StatusMethodNotAllowed},
	} {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(test.method, test.path, nil))
			if recorder.Code != test.status {
				t.Errorf("status = %d, want %d", recorder.Code, test.status)
			}
		})
	}
}

func TestHelloUsesStoredGreeting(t *testing.T) {
	database, err := db.Open(t.Context(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if _, err := database.ExecContext(t.Context(), "UPDATE greetings SET message = ? WHERE id = 1", "Hello, SQLite!"); err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	server.NewHandler(database).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/hello", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var response api.HelloResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Message != "Hello, SQLite!" {
		t.Errorf("message = %q, want stored greeting", response.Message)
	}
}
