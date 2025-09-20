package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
)

// Unit test for the REST router focusing on /langs which is DB-independent.
func TestGetLangList_Unit(t *testing.T) {
	// Build chi router with our server implementation
	r := chi.NewRouter()
	_ = restapi.HandlerFromMux(newRESTHandler(&server{db: nil}), r)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/langs", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /langs status=%d", w.Code)
	}
	var resp restapi.LangListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(resp.Langs) == 0 {
		t.Fatalf("empty langs in response")
	}
	// Basic field sanity check
	if resp.Langs[0].Id == "" || resp.Langs[0].Name == "" || resp.Langs[0].Version == "" {
		t.Fatalf("invalid first lang: %+v", resp.Langs[0])
	}
}
