package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/yosupo06/library-checker-judge/database"
	restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
	"gorm.io/gorm"
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
	req := httptest.NewRequest(http.MethodPatch, "/auth/current_user", bytes.NewReader([]byte(body)))
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

func TestGetUserInfo_SolvedMap(t *testing.T) {
	db := setupTestDB(t)

	problems := []database.Problem{
		{
			Name:             "aplusb",
			Title:            "A + B",
			SourceUrl:        "https://example.com/aplusb",
			Timelimit:        2000,
			TestCasesVersion: "v2",
			Version:          "1",
		},
		{
			Name:             "aplusb_old",
			Title:            "A + B Old",
			SourceUrl:        "https://example.com/aplusb_old",
			Timelimit:        2000,
			TestCasesVersion: "v5",
			Version:          "1",
		},
	}
	for _, p := range problems {
		if err := database.SaveProblem(db, p); err != nil {
			t.Fatalf("save problem %s: %v", p.Name, err)
		}
	}

	if err := database.RegisterUser(db, "alice", "uid-123"); err != nil {
		t.Fatalf("register user: %v", err)
	}

	subs := []database.Submission{
		{
			ProblemName:      "aplusb",
			UserName:         sql.NullString{String: "alice", Valid: true},
			Status:           "AC",
			TestCasesVersion: "v1",
			Source:           "#include <bits/stdc++.h>\nint main(){}",
		},
		{
			ProblemName:      "aplusb",
			UserName:         sql.NullString{String: "alice", Valid: true},
			Status:           "AC",
			TestCasesVersion: "v2",
			Source:           "#include <bits/stdc++.h>\nint main(){}",
		},
		{
			ProblemName:      "aplusb_old",
			UserName:         sql.NullString{String: "alice", Valid: true},
			Status:           "AC",
			TestCasesVersion: "legacy",
			Source:           "#include <bits/stdc++.h>\nint main(){}",
		},
	}
	for i, sub := range subs {
		if _, err := database.SaveSubmission(db, sub); err != nil {
			t.Fatalf("save submission %d: %v", i, err)
		}
	}

	r := chi.NewRouter()
	_ = restapi.HandlerFromMux(&server{db: db}, r)
	req := httptest.NewRequest(http.MethodGet, "/users/alice", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}

	var resp restapi.UserInfoResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if got := resp.SolvedMap["aplusb"]; got != "LATEST_AC" {
		t.Fatalf("expected LATEST_AC for aplusb, got %q", got)
	}
	if got := resp.SolvedMap["aplusb_old"]; got != "AC" {
		t.Fatalf("expected AC for aplusb_old, got %q", got)
	}
	if _, ok := resp.SolvedMap["missing"]; ok {
		t.Fatalf("unexpected entry for missing problem")
	}
}
