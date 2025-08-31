package integration

import (
    "context"
    "os"
    "testing"
    "time"

    "github.com/yosupo06/library-checker-judge/database"
    "gorm.io/gorm"
)

// waitForProblem waits until a problem exists in DB (uploaded by uploader CLI)
func waitForProblem(t *testing.T, db *gorm.DB, name string, timeout time.Duration) error {
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
    // Ensure default connection to local compose
    _ = os.Setenv("PGHOST", "localhost")
    _ = os.Setenv("PGPORT", "5432")
    _ = os.Setenv("PGDATABASE", "librarychecker")
    _ = os.Setenv("PGUSER", "postgres")
    _ = os.Setenv("PGPASSWORD", "lcdummypassword")

    dsn := database.GetDSNFromEnv()
    db := database.Connect(dsn, false)

    // Make sure problem is already uploaded by uploader step
    if err := waitForProblem(t, db, "aplusb", 2*time.Minute); err != nil {
        t.Fatalf("aplusb problem not found in DB (did uploader run?): %v", err)
    }

    // Minimal AC source for A+B (C++)
    src := `#include <bits/stdc++.h>
using namespace std;
int main(){ios::sync_with_stdio(false);cin.tie(nullptr); long long a,b; if(!(cin>>a>>b)) return 0; cout<<a+b<<"\n"; return 0;}`

    // Create submission
    sub := database.Submission{
        SubmissionTime: time.Now(),
        ProblemName:    "aplusb",
        Lang:           "cpp",
        Status:         "WJ",
        Source:         src,
        MaxTime:        -1,
        MaxMemory:      -1,
    }
    id, err := database.SaveSubmission(db, sub)
    if err != nil {
        t.Fatalf("SaveSubmission failed: %v", err)
    }

    // Enqueue task for judge
    if err := database.PushSubmissionTask(db, database.SubmissionData{ID: id, TleKnockout: false}, 50); err != nil {
        t.Fatalf("PushSubmissionTask failed: %v", err)
    }

    // Poll for result
    deadline := time.Now().Add(10 * time.Minute)
    lastStatus := ""
    for time.Now().Before(deadline) {
        sub2, err := database.FetchSubmission(db, id)
        if err != nil {
            t.Fatalf("FetchSubmission failed: %v", err)
        }
        if sub2.Status != lastStatus {
            t.Logf("submission %d status: %s", id, sub2.Status)
            lastStatus = sub2.Status
        }
        switch sub2.Status {
        case "AC":
            t.Logf("AC confirmed. MaxTime=%d ms, MaxMemory=%d B", sub2.MaxTime, sub2.MaxMemory)
            return
        case "WA", "TLE", "MLE", "RE", "CE", "IE", "ICE":
            t.Fatalf("Unexpected non-AC final status: %s", sub2.Status)
            return
        }
        time.Sleep(3 * time.Second)
    }
    t.Fatalf("Timeout waiting for judging result for submission %d", id)
}
