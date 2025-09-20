package main

import (
	"context"
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

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	req.Header.Set("Authorization", "Bearer token")

	ctx := withHTTPRequest(context.Background(), req)
	respObj, err := s.PostSubmit(ctx, restapi.PostSubmitRequestObject{
		Body: &restapi.PostSubmitJSONRequestBody{
			Problem: problem.Name,
			Source:  "#include <bits/stdc++.h>\nint main(){return 0;}",
			Lang:    "cpp",
		},
	})
	if err != nil {
		t.Fatalf("PostSubmit returned error: %v", err)
	}

	resp, ok := respObj.(restapi.PostSubmit200JSONResponse)
	if !ok {
		t.Fatalf("unexpected response type %T", respObj)
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

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	ctx := withHTTPRequest(context.Background(), req)

	respObj, err := s.PostSubmit(ctx, restapi.PostSubmitRequestObject{
		Body: &restapi.PostSubmitJSONRequestBody{
			Problem: problem.Name,
			Source:  "#include <bits/stdc++.h>\nint main(){return 0;}",
			Lang:    "cpp",
		},
	})
	if err != nil {
		t.Fatalf("PostSubmit returned error: %v", err)
	}

	resp, ok := respObj.(restapi.PostSubmit200JSONResponse)
	if !ok {
		t.Fatalf("unexpected response type %T", respObj)
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
