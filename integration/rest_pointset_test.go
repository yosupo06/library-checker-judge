package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	_ "embed"
	"github.com/google/uuid"
	"github.com/yosupo06/library-checker-judge/database"
)

//go:embed testdata/point_set_range_sort_range_composite_correct.cpp
var pointSetCorrectSource string

//go:embed testdata/point_set_range_sort_range_composite_example.in
var pointSetHackInput string

func TestREST_PointSetRangeSortRangeComposite_HackFlow(t *testing.T) {
	if err := waitForREST(t, "http://localhost:12381/health", 2*time.Minute); err != nil {
		t.Fatalf("REST /health not ready: %v", err)
	}

	t.Setenv("PGHOST", "localhost")
	t.Setenv("PGPORT", "5432")
	t.Setenv("PGDATABASE", "librarychecker")
	t.Setenv("PGUSER", "postgres")
	t.Setenv("PGPASSWORD", "lcdummypassword")

	db := database.Connect(database.GetDSNFromEnv(), false)
	if err := waitForProblem(db, "point_set_range_sort_range_composite", 3*time.Minute); err != nil {
		t.Fatalf("problem not found: %v", err)
	}

	idToken := getEmulatorIDToken(t)
	userName := fmt.Sprintf("psrsrc-%s", uuid.NewString()[:8])

	client := &http.Client{Timeout: 5 * time.Second}

	registerPayload := map[string]any{"name": userName}
	rb, _ := json.Marshal(registerPayload)
	regReq, _ := http.NewRequest(http.MethodPost, "http://localhost:12381/auth/register", bytes.NewReader(rb))
	regReq.Header.Set("Content-Type", "application/json")
	regReq.Header.Set("Authorization", "Bearer "+idToken)
	regResp, err := client.Do(regReq)
	if err != nil {
		t.Fatalf("POST /auth/register failed: %v", err)
	}
	if regResp.Body != nil {
		defer func() { _ = regResp.Body.Close() }()
	}
	if regResp.StatusCode != http.StatusOK {
		bb, _ := io.ReadAll(regResp.Body)
		t.Fatalf("register status=%d body=%s", regResp.StatusCode, string(bb))
	}

	submitPayload := map[string]any{
		"problem":      "point_set_range_sort_range_composite",
		"source":       pointSetCorrectSource,
		"lang":         "cpp",
		"tle_knockout": false,
	}
	sb, _ := json.Marshal(submitPayload)
	submitResp, err := client.Post("http://localhost:12381/submit", "application/json", bytes.NewReader(sb))
	if err != nil {
		t.Fatalf("POST /submit failed: %v", err)
	}
	defer func() { _ = submitResp.Body.Close() }()
	if submitResp.StatusCode != http.StatusOK {
		bb, _ := io.ReadAll(submitResp.Body)
		t.Fatalf("submit status=%d body=%s", submitResp.StatusCode, string(bb))
	}
	var submitOut struct {
		Id int32 `json:"id"`
	}
	if err := json.NewDecoder(submitResp.Body).Decode(&submitOut); err != nil {
		t.Fatalf("decode submit response: %v", err)
	}

	deadline := time.Now().Add(10 * time.Minute)
	for time.Now().Before(deadline) {
		url := fmt.Sprintf("http://localhost:12381/submissions/%d", submitOut.Id)
		subResp, err := client.Get(url)
		if err != nil {
			t.Fatalf("GET %s failed: %v", url, err)
		}
		var subInfo struct {
			Overview struct {
				Status string `json:"status"`
			} `json:"overview"`
		}
		if err := json.NewDecoder(subResp.Body).Decode(&subInfo); err != nil {
			_ = subResp.Body.Close()
			t.Fatalf("decode submission info failed: %v", err)
		}
		_ = subResp.Body.Close()
		switch subInfo.Overview.Status {
		case "AC":
			goto SUBMISSION_READY
		case "WA", "TLE", "MLE", "RE", "CE", "IE", "ICE":
			t.Fatalf("unexpected submission status %q", subInfo.Overview.Status)
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("timeout waiting for submission %d", submitOut.Id)

SUBMISSION_READY:
	hackPayload := struct {
		Submission  int32  `json:"submission"`
		TestCaseTxt []byte `json:"test_case_txt"`
	}{
		Submission:  submitOut.Id,
		TestCaseTxt: []byte(pointSetHackInput),
	}
	hb, _ := json.Marshal(hackPayload)
	hackReq, _ := http.NewRequest(http.MethodPost, "http://localhost:12381/hacks", bytes.NewReader(hb))
	hackReq.Header.Set("Content-Type", "application/json")
	hackReq.Header.Set("Authorization", "Bearer "+idToken)
	hackResp, err := client.Do(hackReq)
	if err != nil {
		t.Fatalf("POST /hacks failed: %v", err)
	}
	defer func() { _ = hackResp.Body.Close() }()
	if hackResp.StatusCode != http.StatusOK {
		bb, _ := io.ReadAll(hackResp.Body)
		t.Fatalf("POST /hacks status=%d body=%s", hackResp.StatusCode, string(bb))
	}
	var hackOut struct {
		Id int32 `json:"id"`
	}
	if err := json.NewDecoder(hackResp.Body).Decode(&hackOut); err != nil {
		t.Fatalf("decode hack response: %v", err)
	}

	deadline = time.Now().Add(10 * time.Minute)
	for time.Now().Before(deadline) {
		url := fmt.Sprintf("http://localhost:12381/hacks/%d", hackOut.Id)
		hr, err := client.Get(url)
		if err != nil {
			t.Fatalf("GET %s failed: %v", url, err)
		}
		var hackInfo struct {
			Overview struct {
				Status string `json:"status"`
			} `json:"overview"`
		}
		if err := json.NewDecoder(hr.Body).Decode(&hackInfo); err != nil {
			_ = hr.Body.Close()
			t.Fatalf("decode hack info failed: %v", err)
		}
		_ = hr.Body.Close()
		switch hackInfo.Overview.Status {
		case "AC":
			return
		case "WA":
			return
		case "WJ", "Generating", "Compiling", "Verifying":
		default:
			t.Fatalf("unexpected hack status %q", hackInfo.Overview.Status)
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("timeout waiting for hack %d", hackOut.Id)
}
