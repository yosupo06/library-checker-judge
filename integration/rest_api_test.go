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
	defer func() { _ = resp.Body.Close() }()
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
	defer func() { _ = resp2.Body.Close() }()
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

func TestREST_LangsAndCategories(t *testing.T) {
	// Ensure REST server is up
	if err := waitForREST(t, "http://localhost:12381/health", 2*time.Minute); err != nil {
		t.Fatalf("REST /health not ready: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}

	// GET /langs
	resp, err := client.Get("http://localhost:12381/langs")
	if err != nil {
		t.Fatalf("GET /langs failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		t.Fatalf("GET /langs status=%d", resp.StatusCode)
	}
	var langs struct {
		Langs []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"langs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&langs); err != nil {
		t.Fatalf("decode /langs failed: %v", err)
	}
	if len(langs.Langs) == 0 {
		t.Fatalf("/langs returned empty list")
	}
	foundCpp := false
	for _, l := range langs.Langs {
		if l.ID == "cpp" {
			foundCpp = true
		}
		if l.ID == "" || l.Name == "" || l.Version == "" {
			t.Fatalf("invalid lang entry: %+v", l)
		}
	}
	if !foundCpp {
		t.Fatalf("cpp not found in /langs")
	}

	// GET /categories (poll until non-empty because uploader populates metadata)
	var catsResp *http.Response
	var lastErr error
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		catsResp, lastErr = client.Get("http://localhost:12381/categories")
		if lastErr == nil && catsResp.StatusCode == 200 {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if lastErr != nil {
		t.Fatalf("GET /categories failed: %v", lastErr)
	}
	defer func() { _ = catsResp.Body.Close() }()
	var categories struct {
		Categories []struct {
			Title    string   `json:"title"`
			Problems []string `json:"problems"`
		} `json:"categories"`
	}
	if err := json.NewDecoder(catsResp.Body).Decode(&categories); err != nil {
		t.Fatalf("decode /categories failed: %v", err)
	}
	if len(categories.Categories) == 0 {
		t.Fatalf("/categories returned empty list")
	}
	hasUnionFind := false
	for _, c := range categories.Categories {
		if c.Title == "" || len(c.Problems) == 0 {
			t.Fatalf("invalid category: %+v", c)
		}
		for _, p := range c.Problems {
			if p == "unionfind" {
				hasUnionFind = true
			}
		}
	}
	if !hasUnionFind {
		t.Fatalf("unionfind not present in any category list")
	}
}
