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
	})
	if err != nil {
		t.Fatal(err)
	}

	hackID, err := SaveHack(db, Hack{
		SubmissionID: subID,
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
	})
	if err != nil {
		t.Fatal(err)
	}

	hackID, err := SaveHack(db, Hack{
		SubmissionID: subID,
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
