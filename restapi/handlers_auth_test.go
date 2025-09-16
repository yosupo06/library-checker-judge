package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"github.com/yosupo06/library-checker-judge/database"
	restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
)

func TestParseBearerToken(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer abc.def.ghi")
	tok := parseBearerToken(r)
	if tok != "abc.def.ghi" {
		t.Fatalf("expected token 'abc.def.ghi', got %q", tok)
	}
}

func TestGetCurrentUserInfo_Anonymous(t *testing.T) {
	r := chi.NewRouter()
	_ = restapi.HandlerFromMux(&server{db: nil}, r)
	req := httptest.NewRequest(http.MethodGet, "/auth/current_user", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var m map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := m["user"]; ok {
		t.Fatalf("expected no 'user' field for anonymous response; got %v", m)
	}
}

func TestPostRegister_Unauthorized(t *testing.T) {
	r := chi.NewRouter()
	_ = restapi.HandlerFromMux(&server{db: nil}, r)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader([]byte(`{"name":"alice"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestPatchCurrentUserInfo_Succeeds(t *testing.T) {
	db := setupTestDB(t)
	if err := database.RegisterUser(db, "alice", "uid-123"); err != nil {
		t.Fatalf("register user: %v", err)
	}
	var captured database.User
	called := false
	s := &server{
		db:         db,
		authClient: fakeAuthClient{uid: "uid-123"},
		updateUserFn: func(_ *gorm.DB, u database.User) error {
			called = true
			captured = u
			return nil
		},
	}

	body := `{"user":{"name":"alice","library_url":"https://example.com","is_developer":true}}`
	req := httptest.NewRequest(http.MethodPatch, "/auth/current_user", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()

	s.PatchCurrentUserInfo(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("PATCH /auth/current_user status=%d body=%s", w.Code, w.Body.String())
	}
	if !called {
		t.Fatalf("updateUser was not invoked")
	}
	if captured.UID != "uid-123" {
		t.Fatalf("expected UID uid-123, got %q", captured.UID)
	}
	if captured.LibraryURL != "https://example.com" {
		t.Fatalf("library url not propagated: %q", captured.LibraryURL)
	}
	if !captured.IsDeveloper {
		t.Fatalf("is_developer not propagated")
	}
}

type fakeAuthClient struct {
	uid string
}

func (f fakeAuthClient) parseUID(_ context.Context, token string) string {
	if token == "" {
		return ""
	}
	return f.uid
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dialector := sqlite.Open("file:restapi_test.db?mode=memory&cache=shared&_busy_timeout=5000&_journal_mode=WAL")
	db, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}
