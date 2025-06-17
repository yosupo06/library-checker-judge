package database

import (
	"database/sql"
	"testing"
	"time"
)

func TestFetchRanking(t *testing.T) {
	db := CreateTestDB(t)

	// Create test problems
	problems := []Problem{
		{Name: "aplusb", Title: "A + B", Timelimit: 2000},
		{Name: "amodb", Title: "A mod B", Timelimit: 1000},
		{Name: "unionfind", Title: "Union Find", Timelimit: 5000},
	}
	for _, problem := range problems {
		if err := SaveProblem(db, problem); err != nil {
			t.Fatal(err)
		}
	}

	// Create test users
	users := []string{"alice", "bob", "charlie", "david"}
	for i, user := range users {
		if err := RegisterUser(db, user, user+"_id"+string(rune('0'+i))); err != nil {
			t.Fatal(err)
		}
	}

	// Create test submissions with different AC counts
	// alice: 3 ACs (aplusb, amodb, unionfind)
	// bob: 2 ACs (aplusb, amodb)  
	// charlie: 1 AC (aplusb)
	// david: 0 ACs (only WA submissions)
	submissions := []Submission{
		// alice's submissions
		{ProblemName: "aplusb", UserName: sql.NullString{Valid: true, String: "alice"}, Status: "AC", Source: "alice_aplusb", SubmissionTime: time.Now()},
		{ProblemName: "amodb", UserName: sql.NullString{Valid: true, String: "alice"}, Status: "AC", Source: "alice_amodb", SubmissionTime: time.Now()},
		{ProblemName: "unionfind", UserName: sql.NullString{Valid: true, String: "alice"}, Status: "AC", Source: "alice_unionfind", SubmissionTime: time.Now()},
		
		// bob's submissions
		{ProblemName: "aplusb", UserName: sql.NullString{Valid: true, String: "bob"}, Status: "AC", Source: "bob_aplusb", SubmissionTime: time.Now()},
		{ProblemName: "amodb", UserName: sql.NullString{Valid: true, String: "bob"}, Status: "AC", Source: "bob_amodb", SubmissionTime: time.Now()},
		{ProblemName: "unionfind", UserName: sql.NullString{Valid: true, String: "bob"}, Status: "WA", Source: "bob_unionfind_wa", SubmissionTime: time.Now()},
		
		// charlie's submissions
		{ProblemName: "aplusb", UserName: sql.NullString{Valid: true, String: "charlie"}, Status: "AC", Source: "charlie_aplusb", SubmissionTime: time.Now()},
		{ProblemName: "amodb", UserName: sql.NullString{Valid: true, String: "charlie"}, Status: "WA", Source: "charlie_amodb_wa", SubmissionTime: time.Now()},
		
		// david's submissions
		{ProblemName: "aplusb", UserName: sql.NullString{Valid: true, String: "david"}, Status: "WA", Source: "david_aplusb_wa", SubmissionTime: time.Now()},
		
		// anonymous submission (should be excluded)
		{ProblemName: "aplusb", UserName: sql.NullString{Valid: false}, Status: "AC", Source: "anon_aplusb", SubmissionTime: time.Now()},
	}

	for _, submission := range submissions {
		if _, err := SaveSubmission(db, submission); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("fetch all users", func(t *testing.T) {
		results, totalCount, err := FetchRanking(db, 0, 10)
		if err != nil {
			t.Fatal(err)
		}

		// Should return 3 users with AC submissions (alice, bob, charlie)
		// david has 0 ACs so should not appear in ranking
		if totalCount != 3 {
			t.Errorf("expected totalCount = 3, got %d", totalCount)
		}

		if len(results) != 3 {
			t.Errorf("expected 3 results, got %d", len(results))
		}

		// Check ranking order: alice (3 ACs), bob (2 ACs), charlie (1 AC)
		expected := []UserStatistics{
			{UserName: "alice", AcCount: 3},
			{UserName: "bob", AcCount: 2},
			{UserName: "charlie", AcCount: 1},
		}

		for i, expected := range expected {
			if i >= len(results) {
				t.Errorf("missing result at index %d", i)
				continue
			}
			if results[i].UserName != expected.UserName {
				t.Errorf("expected results[%d].UserName = %s, got %s", i, expected.UserName, results[i].UserName)
			}
			if results[i].AcCount != expected.AcCount {
				t.Errorf("expected results[%d].AcCount = %d, got %d", i, expected.AcCount, results[i].AcCount)
			}
		}
	})

	t.Run("pagination", func(t *testing.T) {
		// Test pagination: skip 1, limit 2
		results, totalCount, err := FetchRanking(db, 1, 2)
		if err != nil {
			t.Fatal(err)
		}

		// Total count should still be 3
		if totalCount != 3 {
			t.Errorf("expected totalCount = 3, got %d", totalCount)
		}

		// Should return 2 results (bob and charlie)
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}

		// Check that we got the right users (skip alice, get bob and charlie)
		if results[0].UserName != "bob" || results[0].AcCount != 2 {
			t.Errorf("expected first result: bob with 2 ACs, got %s with %d ACs", results[0].UserName, results[0].AcCount)
		}
		if results[1].UserName != "charlie" || results[1].AcCount != 1 {
			t.Errorf("expected second result: charlie with 1 AC, got %s with %d ACs", results[1].UserName, results[1].AcCount)
		}
	})

	t.Run("pagination boundary", func(t *testing.T) {
		// Test pagination at boundary: skip 3, limit 2 (should return empty)
		results, totalCount, err := FetchRanking(db, 3, 2)
		if err != nil {
			t.Fatal(err)
		}

		// Total count should still be 3
		if totalCount != 3 {
			t.Errorf("expected totalCount = 3, got %d", totalCount)
		}

		// Should return 0 results (beyond available data)
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})

	t.Run("same AC count sorting", func(t *testing.T) {
		// Create two users with same AC count to test lexicographical sorting
		if err := RegisterUser(db, "zebra", "zebra_id"); err != nil {
			t.Fatal(err)
		}
		if err := RegisterUser(db, "alpha", "alpha_id"); err != nil {
			t.Fatal(err)
		}

		// Both get 1 AC (same as charlie)
		sameACSubmissions := []Submission{
			{ProblemName: "aplusb", UserName: sql.NullString{Valid: true, String: "zebra"}, Status: "AC", Source: "zebra_aplusb", SubmissionTime: time.Now()},
			{ProblemName: "aplusb", UserName: sql.NullString{Valid: true, String: "alpha"}, Status: "AC", Source: "alpha_aplusb", SubmissionTime: time.Now()},
		}

		for _, submission := range sameACSubmissions {
			if _, err := SaveSubmission(db, submission); err != nil {
				t.Fatal(err)
			}
		}

		results, _, err := FetchRanking(db, 0, 10)
		if err != nil {
			t.Fatal(err)
		}

		// Should have 5 users now: alice(3), bob(2), alpha(1), charlie(1), zebra(1)
		// Users with same AC count should be sorted alphabetically
		if len(results) != 5 {
			t.Errorf("expected 5 results, got %d", len(results))
		}

		// Check that users with 1 AC are sorted alphabetically: alpha, charlie, zebra
		usersWithOneAC := []string{}
		for _, result := range results {
			if result.AcCount == 1 {
				usersWithOneAC = append(usersWithOneAC, result.UserName)
			}
		}

		expectedOrder := []string{"alpha", "charlie", "zebra"}
		for i, expected := range expectedOrder {
			if i >= len(usersWithOneAC) {
				t.Errorf("missing user with 1 AC at index %d", i)
				continue
			}
			if usersWithOneAC[i] != expected {
				t.Errorf("expected usersWithOneAC[%d] = %s, got %s", i, expected, usersWithOneAC[i])
			}
		}
	})
}