package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/yosupo06/library-checker-judge/database"
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

func TestGetMonitoring(t *testing.T) {
	db := setupTestDB(t)
	if err := database.RegisterUser(db, "alice", "uid-monitoring"); err != nil {
		t.Fatalf("register user: %v", err)
	}
	if err := database.SaveProblem(db, database.Problem{
		Name:             "aplusb_monitoring",
		Title:            "A + B",
		SourceUrl:        "https://example.com/aplusb",
		Timelimit:        2000,
		TestCasesVersion: "v1",
		Version:          "1",
		OverallVersion:   "1",
	}); err != nil {
		t.Fatalf("save problem: %v", err)
	}
	if _, err := database.SaveSubmission(db, database.Submission{
		ProblemName: "aplusb_monitoring",
		Lang:        "cpp",
		Status:      "AC",
		Source:      "#include <bits/stdc++.h>\nint main(){return 0;}",
		MaxTime:     1,
		MaxMemory:   1,
		UserName:    sql.NullString{String: "alice", Valid: true},
	}); err != nil {
		t.Fatalf("save submission: %v", err)
	}

	r := chi.NewRouter()
	_ = restapi.HandlerFromMux(newRESTHandler(&server{db: db}), r)
	req := httptest.NewRequest(http.MethodGet, "/monitoring", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /monitoring status=%d body=%s", w.Code, w.Body.String())
	}
	var resp restapi.MonitoringResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.TotalUsers != 1 || resp.TotalSubmissions != 1 {
		t.Fatalf("unexpected monitoring response: %+v", resp)
	}
}
