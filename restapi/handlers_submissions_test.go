package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yosupo06/library-checker-judge/database"
	restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
)

func TestPostSubmit_AssignsUserName(t *testing.T) {
	db := setupTestDB(t)
	problem := database.Problem{
		Name:             "aplusb-submit",
		Title:            "A + B",
		SourceUrl:        "https://example.com/aplusb",
		Timelimit:        2000,
		TestCasesVersion: "v1",
		Version:          "1",
		OverallVersion:   "1",
	}
	if err := database.SaveProblem(db, problem); err != nil {
		t.Fatalf("save problem: %v", err)
	}
	if err := database.RegisterUser(db, "alice", "uid-submit"); err != nil {
		t.Fatalf("register user: %v", err)
	}

	s := &server{db: db, authClient: fakeAuthClient{uid: "uid-submit"}}

	body := map[string]any{
		"problem": problem.Name,
		"source":  "#include <bits/stdc++.h>\nint main(){return 0;}",
		"lang":    "cpp",
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()

	s.PostSubmit(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("POST /submit status=%d body=%s", w.Code, w.Body.String())
	}

	var resp restapi.SubmitResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Id <= 0 {
		t.Fatalf("invalid submission id: %d", resp.Id)
	}

	sub, err := database.FetchSubmission(db, resp.Id)
	if err != nil {
		t.Fatalf("fetch submission: %v", err)
	}
	if !sub.UserName.Valid || sub.UserName.String != "alice" {
		t.Fatalf("expected user alice, got %+v", sub.UserName)
	}
}

func TestPostSubmit_AnonymousAllowed(t *testing.T) {
	db := setupTestDB(t)
	problem := database.Problem{
		Name:             "aplusb-unauth",
		Title:            "A + B",
		SourceUrl:        "https://example.com/aplusb",
		Timelimit:        2000,
		TestCasesVersion: "v1",
		Version:          "1",
		OverallVersion:   "1",
	}
	if err := database.SaveProblem(db, problem); err != nil {
		t.Fatalf("save problem: %v", err)
	}

	s := &server{db: db, authClient: fakeAuthClient{uid: "uid-submit"}}

	body := map[string]any{
		"problem": problem.Name,
		"source":  "#include <bits/stdc++.h>\nint main(){return 0;}",
		"lang":    "cpp",
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/submit", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.PostSubmit(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("POST /submit status=%d body=%s", w.Code, w.Body.String())
	}

	var resp restapi.SubmitResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Id <= 0 {
		t.Fatalf("invalid submission id: %d", resp.Id)
	}

	sub, err := database.FetchSubmission(db, resp.Id)
	if err != nil {
		t.Fatalf("fetch submission: %v", err)
	}
	if sub.UserName.Valid {
		t.Fatalf("expected anonymous submission, got %+v", sub.UserName)
	}
}
