package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/yosupo06/library-checker-judge/database"
	"gorm.io/gorm"
)

// waitForProblem waits until a problem exists in DB (uploaded by uploader CLI)
func waitForProblem(db *gorm.DB, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := database.FetchProblem(db, name); err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return context.DeadlineExceeded
}

func TestAplusB_AC(t *testing.T) {
	// Ensure REST server is up
	if err := waitForREST(t, "http://localhost:12381/health", 2*time.Minute); err != nil {
		t.Fatalf("REST /health not ready: %v", err)
	}

	// Ensure DB has problem (for uploader readiness)
	t.Setenv("PGHOST", "localhost")
	t.Setenv("PGPORT", "5432")
	t.Setenv("PGDATABASE", "librarychecker")
	t.Setenv("PGUSER", "postgres")
	t.Setenv("PGPASSWORD", "lcdummypassword")

	dsn := database.GetDSNFromEnv()
	db := database.Connect(dsn, false)
	if err := waitForProblem(db, "aplusb", 2*time.Minute); err != nil {
		t.Fatalf("aplusb problem not found in DB: %v", err)
	}

	// Minimal AC source for A+B (C++)
	src := `#include <bits/stdc++.h>
using namespace std;
int main(){ios::sync_with_stdio(false);cin.tie(nullptr); long long a,b; if(!(cin>>a>>b)) return 0; cout<<a+b<<"\n"; return 0;}`

	// Submit via REST
	body := map[string]any{
		"problem":      "aplusb",
		"source":       src,
		"lang":         "cpp",
		"tle_knockout": false,
	}
	b, _ := json.Marshal(body)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post("http://localhost:12381/submit", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /submit failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		bb, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST /submit status=%d, body=%s", resp.StatusCode, string(bb))
	}
	var submitRes struct {
		Id int32 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&submitRes); err != nil {
		t.Fatalf("decode submit response: %v", err)
	}

	// Poll /submissions/{id}
	deadline := time.Now().Add(10 * time.Minute)
	lastStatus := ""
	for time.Now().Before(deadline) {
		url := fmt.Sprintf("http://localhost:12381/submissions/%d", submitRes.Id)
		r2, err := client.Get(url)
		if err != nil {
			t.Fatalf("GET %s failed: %v", url, err)
		}
		var info struct {
			Overview struct {
				Status string  `json:"status"`
				Time   float64 `json:"time"`
				Memory int64   `json:"memory"`
			} `json:"overview"`
		}
		if err := json.NewDecoder(r2.Body).Decode(&info); err != nil {
			_ = r2.Body.Close()
			t.Fatalf("decode submission info failed: %v", err)
		}
		_ = r2.Body.Close()
		if info.Overview.Status != lastStatus {
			t.Logf("submission %d status: %s", submitRes.Id, info.Overview.Status)
			lastStatus = info.Overview.Status
		}
		switch info.Overview.Status {
		case "AC":
			t.Logf("AC confirmed. Time=%.3f s, Memory=%d B", info.Overview.Time, info.Overview.Memory)
			return
		case "WA", "TLE", "MLE", "RE", "CE", "IE", "ICE":
			t.Fatalf("Unexpected non-AC final status: %s", info.Overview.Status)
			return
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("Timeout waiting for judging result for submission %d", submitRes.Id)
}
