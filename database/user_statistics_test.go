package database

import (
	"database/sql"
	"testing"
)

func TestFetchUserSolvedStatuses(t *testing.T) {
	db := CreateTestDB(t)

	problems := []Problem{
		{
			Name:             "aplusb",
			Title:            "A + B",
			SourceUrl:        "https://example.com/aplusb",
			Timelimit:        2000,
			TestCasesVersion: "v2",
			Version:          "1",
		},
		{
			Name:             "aplusb_old",
			Title:            "Old A + B",
			SourceUrl:        "https://example.com/aplusb_old",
			Timelimit:        2000,
			TestCasesVersion: "v5",
			Version:          "1",
		},
	}
	for _, p := range problems {
		if err := SaveProblem(db, p); err != nil {
			t.Fatalf("save problem %s: %v", p.Name, err)
		}
	}

	if err := RegisterUser(db, "alice", "uid-alice"); err != nil {
		t.Fatalf("register user: %v", err)
	}

	submissions := []Submission{
		{
			ProblemName:      "aplusb",
			UserName:         sql.NullString{String: "alice", Valid: true},
			Status:           "AC",
			TestCasesVersion: "v1",
			Source:           "#include <bits/stdc++.h>\nint main(){}",
		},
		{
			ProblemName:      "aplusb",
			UserName:         sql.NullString{String: "alice", Valid: true},
			Status:           "AC",
			TestCasesVersion: "v2",
			Source:           "#include <bits/stdc++.h>\nint main(){}",
		},
		{
			ProblemName:      "aplusb_old",
			UserName:         sql.NullString{String: "alice", Valid: true},
			Status:           "AC",
			TestCasesVersion: "legacy",
			Source:           "#include <bits/stdc++.h>\nint main(){}",
		},
		{
			ProblemName:      "aplusb",
			UserName:         sql.NullString{String: "alice", Valid: true},
			Status:           "WA",
			TestCasesVersion: "v2",
			Source:           "#include <bits/stdc++.h>\nint main(){}",
		},
	}
	for i, submission := range submissions {
		if _, err := SaveSubmission(db, submission); err != nil {
			t.Fatalf("save submission %d: %v", i, err)
		}
	}

	statuses, err := FetchUserSolvedStatuses(db, "alice")
	if err != nil {
		t.Fatalf("fetch user solved statuses: %v", err)
	}
	got := map[string]bool{}
	for _, status := range statuses {
		got[status.ProblemName] = status.LatestAC
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 problems, got %d", len(got))
	}
	if !got["aplusb"] {
		t.Fatalf("expected aplusb to be latest AC: %v", got)
	}
	if got["aplusb_old"] {
		t.Fatalf("expected aplusb_old to be stale AC: %v", got)
	}

	empty, err := FetchUserSolvedStatuses(db, "bob")
	if err != nil {
		t.Fatalf("unexpected error for missing user: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected empty result for missing user, got %v", empty)
	}

	if _, err := FetchUserSolvedStatuses(db, ""); err == nil {
		t.Fatalf("expected error for empty name")
	}
}
