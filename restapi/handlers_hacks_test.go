package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yosupo06/library-checker-judge/database"
	restapi "github.com/yosupo06/library-checker-judge/restapi/internal/api"
	"gorm.io/gorm"
)

func createTestSubmission(t *testing.T, db *gorm.DB, problemName string) int32 {
	t.Helper()
	if err := db.Create(&database.Problem{
		Name:             problemName,
		Title:            "A + B",
		SourceUrl:        "https://example.com",
		Timelimit:        2000,
		TestCasesVersion: "v1",
		Version:          "v1",
		OverallVersion:   "v1",
	}).Error; err != nil {
		t.Fatalf("create problem: %v", err)
	}
	id, err := database.SaveSubmission(db, database.Submission{
		SubmissionTime: time.Now(),
		ProblemName:    problemName,
		Lang:           "cpp",
		Status:         "WJ",
		Source:         "int main() { return 0; }",
		MaxTime:        -1,
		MaxMemory:      -1,
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}
	return id
}

func TestPostHackAndFetch(t *testing.T) {
	db := setupTestDB(t)
	submissionID := createTestSubmission(t, db, "aplusb-hack1")
	if err := database.RegisterUser(db, "alice", "uid-123"); err != nil {
		t.Fatalf("register user: %v", err)
	}

	s := &server{db: db, authClient: fakeAuthClient{uid: "uid-123"}}
	r := chi.NewRouter()
	_ = restapi.HandlerFromMux(newRESTHandler(s), r)

	payload := map[string]any{
		"submission":    submissionID,
		"test_case_txt": base64.StdEncoding.EncodeToString([]byte("42")),
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/hacks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("POST /hacks status=%d body=%s", w.Code, w.Body.String())
	}

	var hackResp restapi.HackResponse
	if err := json.Unmarshal(w.Body.Bytes(), &hackResp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if hackResp.Id < 0 {
		t.Fatalf("invalid hack id: %d", hackResp.Id)
	}

	hack, err := database.FetchHack(db, hackResp.Id)
	if err != nil {
		t.Fatalf("fetch hack: %v", err)
	}
	if !hack.UserName.Valid || hack.UserName.String != "alice" {
		t.Fatalf("expected user alice, got %+v", hack.UserName)
	}

	infoReq := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/hacks/%d", hackResp.Id), nil)
	infoW := httptest.NewRecorder()
	r.ServeHTTP(infoW, infoReq)
	if infoW.Code != http.StatusOK {
		t.Fatalf("GET /hacks/{id} status=%d body=%s", infoW.Code, infoW.Body.String())
	}

	var info restapi.HackInfoResponse
	if err := json.Unmarshal(infoW.Body.Bytes(), &info); err != nil {
		t.Fatalf("decode hack info: %v", err)
	}
	if info.Overview.Id != hackResp.Id {
		t.Fatalf("unexpected overview id: %d", info.Overview.Id)
	}
	if info.Overview.UserName == nil || *info.Overview.UserName != "alice" {
		t.Fatalf("expected overview user alice, got %+v", info.Overview.UserName)
	}
	if info.TestCaseTxt == nil || string(*info.TestCaseTxt) != "42" {
		t.Fatalf("unexpected test case txt: %v", info.TestCaseTxt)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/hacks", nil)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("GET /hacks status=%d body=%s", listW.Code, listW.Body.String())
	}
	var list restapi.HackListResponse
	if err := json.Unmarshal(listW.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list.Hacks) != 1 {
		t.Fatalf("expected 1 hack, got %d", len(list.Hacks))
	}
}

func TestPostHack_Unauthorized(t *testing.T) {
	db := setupTestDB(t)
	submissionID := createTestSubmission(t, db, "aplusb-hack2")
	s := &server{db: db, authClient: fakeAuthClient{uid: "uid-123"}}
	r := chi.NewRouter()
	_ = restapi.HandlerFromMux(newRESTHandler(s), r)

	payload := map[string]any{
		"submission":    submissionID,
		"test_case_txt": base64.StdEncoding.EncodeToString([]byte("42")),
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/hacks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
