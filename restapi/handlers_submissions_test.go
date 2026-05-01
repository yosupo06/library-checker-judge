package main

import (
	"context"
	"database/sql"
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

func TestPostRejudge_AllowsSubmissionOwner(t *testing.T) {
	db := setupTestDB(t)
	problem := database.Problem{
		Name:             "aplusb-rejudge",
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
	if err := database.RegisterUser(db, "alice", "uid-rejudge"); err != nil {
		t.Fatalf("register user: %v", err)
	}
	id, err := database.SaveSubmission(db, database.Submission{
		ProblemName:      problem.Name,
		Lang:             "cpp",
		Status:           "AC",
		Source:           "#include <bits/stdc++.h>\nint main(){return 0;}",
		TestCasesVersion: "v1",
		MaxTime:          1,
		MaxMemory:        1,
		UserName:         sql.NullString{String: "alice", Valid: true},
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}

	s := &server{db: db, authClient: fakeAuthClient{uid: "uid-rejudge"}}
	req := httptest.NewRequest(http.MethodPost, "/submissions/1/rejudge", nil)
	req.Header.Set("Authorization", "Bearer token")
	ctx := withHTTPRequest(context.Background(), req)
	respObj, err := s.PostRejudge(ctx, restapi.PostRejudgeRequestObject{Id: id})
	if err != nil {
		t.Fatalf("PostRejudge returned error: %v", err)
	}
	if _, ok := respObj.(restapi.PostRejudge200JSONResponse); !ok {
		t.Fatalf("unexpected response type %T", respObj)
	}
}

func TestGetSubmissionInfo_CanRejudgeOwner(t *testing.T) {
	db := setupTestDB(t)
	problem := database.Problem{
		Name:             "aplusb-rejudge-owner",
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
	if err := database.RegisterUser(db, "alice", "uid-owner"); err != nil {
		t.Fatalf("register user: %v", err)
	}
	id, err := database.SaveSubmission(db, database.Submission{
		ProblemName:      problem.Name,
		Lang:             "cpp",
		Status:           "AC",
		Source:           "#include <bits/stdc++.h>\nint main(){return 0;}",
		TestCasesVersion: "v1",
		MaxTime:          1,
		MaxMemory:        1,
		UserName:         sql.NullString{String: "alice", Valid: true},
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}

	s := &server{db: db, authClient: fakeAuthClient{uid: "uid-owner"}}
	req := httptest.NewRequest(http.MethodGet, "/submissions/1", nil)
	req.Header.Set("Authorization", "Bearer token")
	ctx := withHTTPRequest(context.Background(), req)
	respObj, err := s.GetSubmissionInfo(ctx, restapi.GetSubmissionInfoRequestObject{Id: id})
	if err != nil {
		t.Fatalf("GetSubmissionInfo returned error: %v", err)
	}
	resp, ok := respObj.(restapi.GetSubmissionInfo200JSONResponse)
	if !ok {
		t.Fatalf("unexpected response type %T", respObj)
	}
	if !resp.CanRejudge {
		t.Fatalf("expected owner can_rejudge=true")
	}
}

func TestGetSubmissionInfo_CanRejudgeOldAcceptedSubmission(t *testing.T) {
	db := setupTestDB(t)
	problem := database.Problem{
		Name:             "aplusb-rejudge-old-ac",
		Title:            "A + B",
		SourceUrl:        "https://example.com/aplusb",
		Timelimit:        2000,
		TestCasesVersion: "v2",
		Version:          "1",
		OverallVersion:   "1",
	}
	if err := database.SaveProblem(db, problem); err != nil {
		t.Fatalf("save problem: %v", err)
	}
	if err := database.RegisterUser(db, "alice", "uid-alice"); err != nil {
		t.Fatalf("register alice: %v", err)
	}
	if err := database.RegisterUser(db, "bob", "uid-bob"); err != nil {
		t.Fatalf("register bob: %v", err)
	}
	id, err := database.SaveSubmission(db, database.Submission{
		ProblemName:      problem.Name,
		Lang:             "cpp",
		Status:           "AC",
		Source:           "#include <bits/stdc++.h>\nint main(){return 0;}",
		TestCasesVersion: "v1",
		MaxTime:          1,
		MaxMemory:        1,
		UserName:         sql.NullString{String: "alice", Valid: true},
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}

	s := &server{db: db, authClient: fakeAuthClient{uid: "uid-bob"}}
	req := httptest.NewRequest(http.MethodGet, "/submissions/1", nil)
	req.Header.Set("Authorization", "Bearer token")
	ctx := withHTTPRequest(context.Background(), req)
	respObj, err := s.GetSubmissionInfo(ctx, restapi.GetSubmissionInfoRequestObject{Id: id})
	if err != nil {
		t.Fatalf("GetSubmissionInfo returned error: %v", err)
	}
	resp, ok := respObj.(restapi.GetSubmissionInfo200JSONResponse)
	if !ok {
		t.Fatalf("unexpected response type %T", respObj)
	}
	if !resp.CanRejudge {
		t.Fatalf("expected old AC submission can_rejudge=true for non-owner")
	}
}

func TestPostRejudge_RejectsNonOwnerWithoutEligibility(t *testing.T) {
	db := setupTestDB(t)
	problem := database.Problem{
		Name:             "aplusb-rejudge-forbidden",
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
	if err := database.RegisterUser(db, "alice", "uid-alice"); err != nil {
		t.Fatalf("register alice: %v", err)
	}
	if err := database.RegisterUser(db, "bob", "uid-bob"); err != nil {
		t.Fatalf("register bob: %v", err)
	}
	id, err := database.SaveSubmission(db, database.Submission{
		ProblemName:      problem.Name,
		Lang:             "cpp",
		Status:           "AC",
		Source:           "#include <bits/stdc++.h>\nint main(){return 0;}",
		TestCasesVersion: "v1",
		MaxTime:          1,
		MaxMemory:        1,
		UserName:         sql.NullString{String: "alice", Valid: true},
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}

	s := &server{db: db, authClient: fakeAuthClient{uid: "uid-bob"}}
	req := httptest.NewRequest(http.MethodPost, "/submissions/1/rejudge", nil)
	req.Header.Set("Authorization", "Bearer token")
	ctx := withHTTPRequest(context.Background(), req)
	_, err = s.PostRejudge(ctx, restapi.PostRejudgeRequestObject{Id: id})
	if err == nil {
		t.Fatalf("expected forbidden error")
	}
	if httpErr, ok := getHTTPError(err); !ok || httpErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403, got %v", err)
	}
}

func TestPostRejudge_RejectsAnonymous(t *testing.T) {
	db := setupTestDB(t)
	s := &server{db: db}
	req := httptest.NewRequest(http.MethodPost, "/submissions/1/rejudge", nil)
	ctx := withHTTPRequest(context.Background(), req)
	_, err := s.PostRejudge(ctx, restapi.PostRejudgeRequestObject{Id: 1})
	if err == nil {
		t.Fatalf("expected unauthorized error")
	}
	if httpErr, ok := getHTTPError(err); !ok || httpErr.Status != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %v", err)
	}
}
