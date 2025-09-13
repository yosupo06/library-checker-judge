package main

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/go-chi/chi/v5"
    "github.com/yosupo06/library-checker-judge/database"
    apitypes "github.com/yosupo06/library-checker-judge/restapi/internal/api"
)

var dummyProblem = database.Problem{
    Name:             "aplusb",
    Title:            "A + B",
    Timelimit:        2000,
    TestCasesVersion: "dummy-testcase-version",
    Version:          "dummy-version",
    SourceUrl:        "https://github.com/yosupo06/library-checker-problems/tree/master/sample/aplusb",
}

func TestProblemInfo(t *testing.T) {
    db := database.CreateTestDB(t)
    if err := database.SaveProblem(db, dummyProblem); err != nil {
      t.Fatal("failed to save problem:", err)
    }

    r := chi.NewRouter()
    _ = apitypes.HandlerFromMux(&server{db: db}, r)
    srv := httptest.NewServer(r)
    defer srv.Close()

    resp, err := http.Get(srv.URL + "/problems/" + dummyProblem.Name)
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("unexpected status: %d", resp.StatusCode)
    }

    var out apitypes.ProblemInfoResponse
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        t.Fatal(err)
    }
    if out.Title != dummyProblem.Title {
        t.Fatalf("title mismatch: got %q", out.Title)
    }
    if out.SourceUrl != dummyProblem.SourceUrl {
        t.Fatalf("source_url mismatch: got %q", out.SourceUrl)
    }
    if out.Version != dummyProblem.Version {
        t.Fatalf("version mismatch: got %q", out.Version)
    }
    if out.TestcasesVersion != dummyProblem.TestCasesVersion {
        t.Fatalf("testcases_version mismatch: got %q", out.TestcasesVersion)
    }
    if out.TimeLimit < 1.9 || out.TimeLimit > 2.1 { // 2000ms -> 2.0s
        t.Fatalf("time_limit mismatch: got %v", out.TimeLimit)
    }
}

func TestProblemInfo_NotFound(t *testing.T) {
    db := database.CreateTestDB(t)
    r := chi.NewRouter()
    _ = apitypes.HandlerFromMux(&server{db: db}, r)
    srv := httptest.NewServer(r)
    defer srv.Close()

    resp, err := http.Get(srv.URL + "/problems/does-not-exist")
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusNotFound {
        t.Fatalf("expected 404, got %d", resp.StatusCode)
    }
}

func TestGetProblems(t *testing.T) {
    db := database.CreateTestDB(t)
    if err := database.SaveProblem(db, dummyProblem); err != nil {
        t.Fatal("failed to save problem:", err)
    }

    r := chi.NewRouter()
    _ = apitypes.HandlerFromMux(&server{db: db}, r)
    srv := httptest.NewServer(r)
    defer srv.Close()

    resp, err := http.Get(srv.URL + "/problems")
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("unexpected status: %d", resp.StatusCode)
    }

    var out apitypes.ProblemListResponse
    if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
        t.Fatal(err)
    }
    if len(out.Problems) == 0 {
        t.Fatalf("expected at least one problem")
    }
}
