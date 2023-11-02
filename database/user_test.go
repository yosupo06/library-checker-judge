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
}

func TestRegisterInvalidUserName(t *testing.T) {
	db := createTestDB(t)

	if err := RegisterUser(db, "name cannot contains space", "id"); err == nil {
		t.Fatal("register user is succeeded with invalid name")
	} else {
		t.Log(err)
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

func TestUpdateUser(t *testing.T) {
	const libraryURL = "https://judge.yosupo.com"

	db := createTestDB(t)

	if err := RegisterUser(db, "name", "id"); err != nil {
		t.Fatal(err)
	}
	if err := UpdateUser(db, User{
		Name:       "name",
		UID:        "id",
		LibraryURL: libraryURL,
	}); err != nil {
		t.Fatal(err)
	}

	if user, err := FetchUser(db, "name"); err != nil || user.LibraryURL != libraryURL {
		t.Fatal(err, user)
	}
}

func TestUpdateUserWithInvalidURL(t *testing.T) {
	const libraryURL = "invalid-url"

	db := createTestDB(t)

	if err := RegisterUser(db, "name", "id"); err != nil {
		t.Fatal(err)
	}
	if err := UpdateUser(db, User{
		Name:       "name",
		UID:        "id",
		LibraryURL: libraryURL,
	}); err == nil {
		t.Fatal("UpdateUser is succeeded with invalid URL")
	} else {
		t.Log(err)
	}
}
