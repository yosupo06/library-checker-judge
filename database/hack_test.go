package database

import (
	"database/sql"
	"testing"
)

func TestHack(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	if err := RegisterUser(db, "user1", "id1"); err != nil {
		t.Fatal(err)
	}

	subID, err := SaveSubmission(db, Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
		Source:      "source",
	})
	if err != nil {
		t.Fatal(err)
	}

	hackID, err := SaveHack(db, Hack{
		SubmissionID: subID,
		TestCaseCpp:  []byte{},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(hackID)

	hack, err := FetchHack(db, hackID)
	if err != nil {
		t.Fatal(err)
	}
	if hack.Submission.ID != subID {
		t.Fatal("hack.Submission.ID != subID", hack, subID)
	}
}

func TestSave(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	if err := RegisterUser(db, "user1", "id1"); err != nil {
		t.Fatal(err)
	}

	subID, err := SaveSubmission(db, Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
		Source:      "source",
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := SaveHack(db, Hack{
		SubmissionID: subID,
	}); err == nil {
		t.Fatal("success to save")
	} else {
		t.Log(err)
	}

	if _, err := SaveHack(db, Hack{
		SubmissionID: subID,
		TestCaseCpp:  []byte{},
		TestCaseTxt:  []byte{},
	}); err == nil {
		t.Fatal("success to save")
	} else {
		t.Log(err)
	}
}

func TestFetchInvalidHack(t *testing.T) {
	db := CreateTestDB(t)

	_, err := FetchHack(db, 123)
	if err != ErrNotExist {
		t.Fatal(err)
	}
}

func TestUpdateHack(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	if err := RegisterUser(db, "user1", "id1"); err != nil {
		t.Fatal(err)
	}

	subID, err := SaveSubmission(db, Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
		Source:      "source",
	})
	if err != nil {
		t.Fatal(err)
	}

	hackID, err := SaveHack(db, Hack{
		SubmissionID: subID,
		TestCaseTxt:  []byte{},
	})
	if err != nil {
		t.Fatal(err)
	}

	hack, err := FetchHack(db, hackID)
	if err != nil {
		t.Fatal(err)
	}

	hack.Status = "AC"
	if err := UpdateHack(db, hack); err != nil {
		t.Fatal(err)
	}

	hack2, err := FetchHack(db, hackID)
	if err != nil {
		t.Fatal(err)
	}
	if hack2.Status != "AC" {
		t.Fatal("hack2.Status is not AC", hack2)
	}
}

func TestHackUserRelation(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	// Register a user
	if err := RegisterUser(db, "testuser", "uid123"); err != nil {
		t.Fatal(err)
	}

	// Create submission by the user
	subID, err := SaveSubmission(db, Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "testuser"},
		Source:      "source",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Test 1: Hack with valid user (UserName is set)
	hackWithUser, err := SaveHack(db, Hack{
		SubmissionID: subID,
		UserName:     sql.NullString{Valid: true, String: "testuser"},
		TestCaseTxt:  []byte("test case"),
		Status:       "WJ",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Test 2: Anonymous hack (UserName is null)
	hackAnonymous, err := SaveHack(db, Hack{
		SubmissionID: subID,
		UserName:     sql.NullString{Valid: false, String: ""},
		TestCaseTxt:  []byte("anonymous test case"),
		Status:       "WJ",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Fetch hack with user - should have User relation populated
	fetchedHackWithUser, err := FetchHack(db, hackWithUser)
	if err != nil {
		t.Fatal(err)
	}

	// Verify User relation is properly populated
	if fetchedHackWithUser.User == nil {
		t.Fatal("Expected User relation to be populated for hack with valid UserName")
	}
	if fetchedHackWithUser.User.Name != "testuser" {
		t.Fatalf("Expected User.Name to be 'testuser', got '%s'", fetchedHackWithUser.User.Name)
	}
	if !fetchedHackWithUser.UserName.Valid {
		t.Fatal("Expected UserName to be valid for hack with user")
	}
	if fetchedHackWithUser.UserName.String != "testuser" {
		t.Fatalf("Expected UserName.String to be 'testuser', got '%s'", fetchedHackWithUser.UserName.String)
	}

	// Fetch anonymous hack - should have User relation as nil
	fetchedAnonymousHack, err := FetchHack(db, hackAnonymous)
	if err != nil {
		t.Fatal(err)
	}

	// Verify User relation is nil for anonymous hack
	if fetchedAnonymousHack.User != nil {
		t.Fatal("Expected User relation to be nil for anonymous hack")
	}
	if fetchedAnonymousHack.UserName.Valid {
		t.Fatal("Expected UserName to be invalid for anonymous hack")
	}
}

func TestFetchHackList(t *testing.T) {
	db := CreateTestDB(t)

	createDummyProblem(t, db)

	// Register users
	if err := RegisterUser(db, "user1", "uid1"); err != nil {
		t.Fatal(err)
	}
	if err := RegisterUser(db, "user2", "uid2"); err != nil {
		t.Fatal(err)
	}

	// Create submissions
	subID1, err := SaveSubmission(db, Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user1"},
		Source:      "source1",
	})
	if err != nil {
		t.Fatal(err)
	}

	subID2, err := SaveSubmission(db, Submission{
		ProblemName: "aplusb",
		UserName:    sql.NullString{Valid: true, String: "user2"},
		Source:      "source2",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create hacks: 2 by user1, 1 by user2, 1 anonymous
	hackIDs := make([]int32, 4)

	// Hack by user1
	hackIDs[0], err = SaveHack(db, Hack{
		SubmissionID: subID1,
		UserName:     sql.NullString{Valid: true, String: "user1"},
		TestCaseTxt:  []byte("test1"),
		Status:       "AC",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Another hack by user1
	hackIDs[1], err = SaveHack(db, Hack{
		SubmissionID: subID2,
		UserName:     sql.NullString{Valid: true, String: "user1"},
		TestCaseTxt:  []byte("test2"),
		Status:       "WA",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Hack by user2
	hackIDs[2], err = SaveHack(db, Hack{
		SubmissionID: subID1,
		UserName:     sql.NullString{Valid: true, String: "user2"},
		TestCaseTxt:  []byte("test3"),
		Status:       "AC",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Anonymous hack
	hackIDs[3], err = SaveHack(db, Hack{
		SubmissionID: subID2,
		UserName:     sql.NullString{Valid: false, String: ""},
		TestCaseTxt:  []byte("test4"),
		Status:       "WJ",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Test fetching all hacks
	allHacks, err := FetchHackList(db, 0, 10, "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(allHacks) != 4 {
		t.Fatalf("Expected 4 hacks, got %d", len(allHacks))
	}

	// Verify User relations are properly set
	for i, hack := range allHacks {
		if hack.UserName.Valid {
			// For hacks with users, User relation should be populated
			if hack.User == nil {
				t.Fatalf("Hack %d: Expected User relation to be populated", i)
			}
			if hack.User.Name != hack.UserName.String {
				t.Fatalf("Hack %d: User.Name (%s) doesn't match UserName.String (%s)",
					i, hack.User.Name, hack.UserName.String)
			}
		} else {
			// For anonymous hacks, User relation should be nil
			if hack.User != nil {
				t.Fatalf("Hack %d: Expected User relation to be nil for anonymous hack", i)
			}
		}
	}

	// Test filtering by user
	user1Hacks, err := FetchHackList(db, 0, 10, "user1", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(user1Hacks) != 2 {
		t.Fatalf("Expected 2 hacks by user1, got %d", len(user1Hacks))
	}

	// Test filtering by status
	acHacks, err := FetchHackList(db, 0, 10, "", "AC", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(acHacks) != 2 {
		t.Fatalf("Expected 2 AC hacks, got %d", len(acHacks))
	}

	// Test count
	totalCount, err := CountHacks(db, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if totalCount != 4 {
		t.Fatalf("Expected total count 4, got %d", totalCount)
	}

	user1Count, err := CountHacks(db, "user1", "")
	if err != nil {
		t.Fatal(err)
	}
	if user1Count != 2 {
		t.Fatalf("Expected user1 count 2, got %d", user1Count)
	}
}
