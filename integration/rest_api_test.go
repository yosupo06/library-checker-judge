package integration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/yosupo06/library-checker-judge/database"
)

func waitForREST(t *testing.T, url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == 200 && string(body) == "SERVING" {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}
	return context.DeadlineExceeded
}

func TestREST_ProblemsAndInfo(t *testing.T) {
	// Ensure REST server is up (docker compose service api-rest)
	if err := waitForREST(t, "http://localhost:12381/health", 2*time.Minute); err != nil {
		t.Fatalf("REST /health not ready: %v", err)
	}

	// Wait for aplusb to be uploaded (shared helper)
	// Reuse DB env from existing integration test
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

	// GET /api/problems
	client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Get("http://localhost:12381/problems")
	if err != nil {
		t.Fatalf("GET /api/problems failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("GET /api/problems status=%d", resp.StatusCode)
	}
	var list struct {
		Problems []struct {
			Name  string `json:"name"`
			Title string `json:"title"`
		} `json:"problems"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decode /api/problems failed: %v", err)
	}
	if len(list.Problems) == 0 {
		t.Fatalf("/api/problems returned empty list")
	}
	found := false
	for _, p := range list.Problems {
		if p.Name == "aplusb" {
			found = true
			if p.Title == "" {
				t.Fatalf("aplusb has empty title in /api/problems")
			}
			break
		}
	}
	if !found {
		t.Fatalf("aplusb not found in /api/problems")
	}

	// GET /api/problems/aplusb
    resp2, err := client.Get("http://localhost:12381/problems/aplusb")
	if err != nil {
		t.Fatalf("GET /api/problems/aplusb failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Fatalf("GET /api/problems/aplusb status=%d", resp2.StatusCode)
	}
	var info struct {
		Title            string  `json:"title"`
		SourceURL        string  `json:"source_url"`
		TimeLimit        float64 `json:"time_limit"`
		Version          string  `json:"version"`
		TestcasesVersion string  `json:"testcases_version"`
		OverallVersion   string  `json:"overall_version"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&info); err != nil {
		t.Fatalf("decode /api/problems/aplusb failed: %v", err)
	}
	if info.Title == "" || info.TimeLimit <= 0 {
		t.Fatalf("invalid problem info: %+v", info)
	}
}
