package database

import (
	"testing"
)

func TestRegisterUser(t *testing.T) {
	db := createTestDB(t)

	if err := RegisterUser(db, "name", "id"); err != nil {
		t.Fatal(err)
	}

	if user, err := FetchUser(db, "name"); err != nil || user.Name != "name" || user.UID != "id" {
		t.Fatal(err, user)
	}

	if user, err := FetchUser(db, "name"); err != nil || user.Name != "name" || user.UID != "id" {
		t.Fatal(err, user)
	}
}

func TestFetchUserFromUID(t *testing.T) {
	db := createTestDB(t)

	if err := RegisterUser(db, "name", "id"); err != nil {
		t.Fatal(err)
	}

	if user, err := FetchUserFromUID(db, "id"); err != nil || user.Name != "name" {
		t.Fatal(err, user)
	}
}
