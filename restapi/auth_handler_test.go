package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bytes"
	"github.com/go-chi/chi/v5"
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
