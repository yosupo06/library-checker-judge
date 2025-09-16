package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

type firebaseSignUpResp struct {
	IdToken string `json:"idToken"`
	LocalId string `json:"localId"`
	Email   string `json:"email"`
}

func waitForFirebaseEmulator(t *testing.T, base string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		// Try a dry-run signUp call with invalid body to just see emulator up
		req, _ := http.NewRequest(http.MethodPost, base+"/identitytoolkit.googleapis.com/v1/accounts:signUp?key=dev", bytes.NewReader([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			_, _ = io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode > 0 { // any response means emulator is up
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("firebase emulator not ready at %s", base)
}

func getEmulatorIDToken(t *testing.T) string {
	// Ensure emulator is up
	if err := waitForFirebaseEmulator(t, "http://localhost:9099", 1*time.Minute); err != nil {
		t.Fatalf("firebase emulator not ready: %v", err)
	}
	email := fmt.Sprintf("user-%s@example.com", uuid.NewString())
	payload := map[string]any{
		"email":             email,
		"password":          "password",
		"returnSecureToken": true,
	}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:9099/identitytoolkit.googleapis.com/v1/accounts:signUp?key=dev", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("signUp request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		bb, _ := io.ReadAll(resp.Body)
		t.Fatalf("signUp status=%d body=%s", resp.StatusCode, string(bb))
	}
	var out firebaseSignUpResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode signUp: %v", err)
	}
	if out.IdToken == "" {
		t.Fatalf("empty idToken from emulator")
	}
	return out.IdToken
}

func TestREST_Auth_RegisterAndCurrentUser(t *testing.T) {
	// Wait for REST server
	if err := waitForREST(t, "http://localhost:12381/health", 2*time.Minute); err != nil {
		t.Fatalf("REST /health not ready: %v", err)
	}

	idToken := getEmulatorIDToken(t)
	name := fmt.Sprintf("restuser-%s", uuid.NewString()[:8])

	// Register
	regBody := map[string]any{"name": name}
	rb, _ := json.Marshal(regBody)
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:12381/auth/register", bytes.NewReader(rb))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+idToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /auth/register failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		bb, _ := io.ReadAll(resp.Body)
		t.Fatalf("register status=%d body=%s", resp.StatusCode, string(bb))
	}

	// current_user
	req2, _ := http.NewRequest(http.MethodGet, "http://localhost:12381/auth/current_user", nil)
	req2.Header.Set("Authorization", "Bearer "+idToken)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("GET /auth/current_user failed: %v", err)
	}
	defer func() { _ = resp2.Body.Close() }()
	if resp2.StatusCode != 200 {
		t.Fatalf("current_user status=%d", resp2.StatusCode)
	}
	var info struct {
		User *struct {
			Name        string `json:"name"`
			LibraryURL  string `json:"library_url"`
			IsDeveloper bool   `json:"is_developer"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&info); err != nil {
		t.Fatalf("decode current_user: %v", err)
	}
	if info.User == nil || info.User.Name != name {
		t.Fatalf("unexpected user: %+v", info.User)
	}

	// Update current user info
	patch := map[string]any{
		"user": map[string]any{
			"name":         name,
			"library_url":  "https://example.com",
			"is_developer": false,
		},
	}
	pb, _ := json.Marshal(patch)
	req3, _ := http.NewRequest(http.MethodPatch, "http://localhost:12381/auth/current_user", bytes.NewReader(pb))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("Authorization", "Bearer "+idToken)
	resp3, err := http.DefaultClient.Do(req3)
	if err != nil {
		t.Fatalf("PATCH /auth/current_user failed: %v", err)
	}
	_ = resp3.Body.Close()
	if resp3.StatusCode != 200 {
		t.Fatalf("PATCH current_user status=%d", resp3.StatusCode)
	}

	// Verify update
	req4, _ := http.NewRequest(http.MethodGet, "http://localhost:12381/auth/current_user", nil)
	req4.Header.Set("Authorization", "Bearer "+idToken)
	resp4, err := http.DefaultClient.Do(req4)
	if err != nil {
		t.Fatalf("GET /auth/current_user(2) failed: %v", err)
	}
	defer func() { _ = resp4.Body.Close() }()
	var info2 struct {
		User *struct {
			Name       string `json:"name"`
			LibraryURL string `json:"library_url"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp4.Body).Decode(&info2); err != nil {
		t.Fatalf("decode current_user(2): %v", err)
	}
	if info2.User == nil || info2.User.LibraryURL != "https://example.com" {
		t.Fatalf("library_url not updated: %+v", info2.User)
	}
}
