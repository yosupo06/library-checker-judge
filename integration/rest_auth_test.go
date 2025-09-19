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
	"github.com/yosupo06/library-checker-judge/database"
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

func TestREST_HackAplusB(t *testing.T) {
	if err := waitForREST(t, "http://localhost:12381/health", 2*time.Minute); err != nil {
		t.Fatalf("REST /health not ready: %v", err)
	}

	t.Setenv("PGHOST", "localhost")
	t.Setenv("PGPORT", "5432")
	t.Setenv("PGDATABASE", "librarychecker")
	t.Setenv("PGUSER", "postgres")
	t.Setenv("PGPASSWORD", "lcdummypassword")

	db := database.Connect(database.GetDSNFromEnv(), false)
	if err := waitForProblem(db, "aplusb", 2*time.Minute); err != nil {
		t.Fatalf("aplusb problem not found: %v", err)
	}

	idToken := getEmulatorIDToken(t)
	name := fmt.Sprintf("hackuser-%s", uuid.NewString()[:8])

	regBody := map[string]any{"name": name}
	rb, _ := json.Marshal(regBody)
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:12381/auth/register", bytes.NewReader(rb))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+idToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /auth/register failed: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bb, _ := io.ReadAll(resp.Body)
		t.Fatalf("register status=%d body=%s", resp.StatusCode, string(bb))
	}

	client := &http.Client{Timeout: 5 * time.Second}

	badSrc := `#include <bits/stdc++.h>
using namespace std;
int main(){ios::sync_with_stdio(false);cin.tie(nullptr); long long a,b; if(!(cin>>a>>b)) return 0; if(a==12345 && b==54321){cout<<a-b<<"\n";} else {cout<<a+b<<"\n";} return 0;}`
	submitReq := map[string]any{
		"problem":      "aplusb",
		"source":       badSrc,
		"lang":         "cpp",
		"tle_knockout": false,
	}
	sb, _ := json.Marshal(submitReq)
	submitResp, err := client.Post("http://localhost:12381/submit", "application/json", bytes.NewReader(sb))
	if err != nil {
		t.Fatalf("POST /submit failed: %v", err)
	}
	defer func() {
		if cerr := submitResp.Body.Close(); cerr != nil {
			t.Fatalf("close submit response body: %v", cerr)
		}
	}()
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
		r, err := client.Get(url)
		if err != nil {
			t.Fatalf("GET %s failed: %v", url, err)
		}
		var info struct {
			Overview struct {
				Status string `json:"status"`
			} `json:"overview"`
		}
		if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
			_ = r.Body.Close()
			t.Fatalf("decode submission info failed: %v", err)
		}
		_ = r.Body.Close()
		switch info.Overview.Status {
		case "AC":
			goto SUBMISSION_DONE
		case "WA", "TLE", "MLE", "RE", "CE", "IE", "ICE":
			t.Fatalf("unexpected submission status %q", info.Overview.Status)
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("timeout waiting for submission %d", submitOut.Id)

SUBMISSION_DONE:
	hackPayload := struct {
		Submission  int32  `json:"submission"`
		TestCaseTxt []byte `json:"test_case_txt"`
	}{
		Submission:  submitOut.Id,
		TestCaseTxt: []byte("12345 54321\n"),
	}
	hb, _ := json.Marshal(hackPayload)
	hReq, _ := http.NewRequest(http.MethodPost, "http://localhost:12381/hacks", bytes.NewReader(hb))
	hReq.Header.Set("Content-Type", "application/json")
	hReq.Header.Set("Authorization", "Bearer "+idToken)
	hResp, err := client.Do(hReq)
	if err != nil {
		t.Fatalf("POST /hacks failed: %v", err)
	}
	defer func() {
		if cerr := hResp.Body.Close(); cerr != nil {
			t.Fatalf("close hack response body: %v", cerr)
		}
	}()
	if hResp.StatusCode != http.StatusOK {
		bb, _ := io.ReadAll(hResp.Body)
		t.Fatalf("POST /hacks status=%d body=%s", hResp.StatusCode, string(bb))
	}
	var hackOut struct {
		Id int32 `json:"id"`
	}
	if err := json.NewDecoder(hResp.Body).Decode(&hackOut); err != nil {
		t.Fatalf("decode hack response: %v", err)
	}

	var hackInfo struct {
		Overview struct {
			Status   string  `json:"status"`
			UserName *string `json:"user_name"`
		} `json:"overview"`
		TestCaseTxt []byte `json:"test_case_txt"`
		JudgeOutput []byte `json:"judge_output"`
	}
	deadline = time.Now().Add(10 * time.Minute)
	for time.Now().Before(deadline) {
		url := fmt.Sprintf("http://localhost:12381/hacks/%d", hackOut.Id)
		r, err := client.Get(url)
		if err != nil {
			t.Fatalf("GET %s failed: %v", url, err)
		}
		if err := json.NewDecoder(r.Body).Decode(&hackInfo); err != nil {
			_ = r.Body.Close()
			t.Fatalf("decode hack info failed: %v", err)
		}
		_ = r.Body.Close()
		switch hackInfo.Overview.Status {
		case "WJ", "Generating", "Compiling", "Verifying":
		case "WA":
			goto HACK_DONE
		case "AC":
			t.Fatalf("hack did not break the submission (status %q)", hackInfo.Overview.Status)
		default:
			if hackInfo.Overview.Status != "" {
				t.Fatalf("unexpected hack status %q", hackInfo.Overview.Status)
			}
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("timeout waiting for hack %d", hackOut.Id)

HACK_DONE:
	if hackInfo.Overview.UserName == nil || *hackInfo.Overview.UserName != name {
		t.Fatalf("hack user mismatch: %+v", hackInfo.Overview.UserName)
	}
	if string(hackInfo.TestCaseTxt) != "12345 54321\n" {
		t.Fatalf("unexpected test case contents: %q", string(hackInfo.TestCaseTxt))
	}
	if len(hackInfo.JudgeOutput) == 0 {
		t.Fatalf("expected judge output to be populated")
	}

	listURL := fmt.Sprintf("http://localhost:12381/hacks?user=%s", name)
	listResp, err := client.Get(listURL)
	if err != nil {
		t.Fatalf("GET %s failed: %v", listURL, err)
	}
	defer func() {
		if cerr := listResp.Body.Close(); cerr != nil {
			t.Fatalf("close hack list response body: %v", cerr)
		}
	}()
	if listResp.StatusCode != http.StatusOK {
		bb, _ := io.ReadAll(listResp.Body)
		t.Fatalf("GET /hacks?user status=%d body=%s", listResp.StatusCode, string(bb))
	}
	var list struct {
		Hacks []struct {
			Id     int32  `json:"id"`
			Status string `json:"status"`
		} `json:"hacks"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&list); err != nil {
		t.Fatalf("decode hack list: %v", err)
	}
	if len(list.Hacks) == 0 {
		t.Fatalf("hack list empty for user %s", name)
	}
	found := false
	for _, h := range list.Hacks {
		if h.Id == hackOut.Id {
			found = true
			if h.Status != "WA" {
				t.Fatalf("expected hack status WA, got %s", h.Status)
			}
		}
	}
	if !found {
		t.Fatalf("hack %d not found in list", hackOut.Id)
	}
}
