package database

import (
	"testing"
)

func TestRegisterUser(t *testing.T) {
	db := createTestDB(t)

	if err := RegisterUser(db, "name", "id"); err != nil {
		t.Fatal(err)
	}

	if user, err := FetchUserFromName(db, "name"); err != nil || user.Name != "name" || user.UID != "id" {
		t.Fatal(err, user)
	}
}

func TestRegisterExistingUser(t *testing.T) {
	db := createTestDB(t)

	if err := RegisterUser(db, "name", "id"); err != nil {
		t.Fatal(err)
	}

	if err := RegisterUser(db, "name", "id2"); err == nil {
		t.Fatal("failure expected")
	} else {
		t.Log(err)
	}

	if err := RegisterUser(db, "name2", "id"); err == nil {
		t.Fatal("failure expected")
	} else {
		t.Log(err)
	}
}

func TestRegisterInvalidUser(t *testing.T) {
	db := createTestDB(t)

	if err := RegisterUser(db, "", "id"); err == nil {
		t.Fatal("register user is succeeded with empty name")
	} else {
		t.Log(err)
	}

	if err := RegisterUser(db, "name", ""); err == nil {
		t.Fatal("register user is succeeded with empty id")
	} else {
		t.Log(err)
	}

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
	const libraryURL1 = "https://library1.yosupo.com"
	const libraryURL2 = "https://library2.yosupo.com"

	db := createTestDB(t)

	if err := RegisterUser(db, "name1", "id1"); err != nil {
		t.Fatal(err)
	}
	if err := RegisterUser(db, "name2", "id2"); err != nil {
		t.Fatal(err)
	}

	if err := UpdateUser(db, User{
		Name:       "name1",
		UID:        "id1",
		LibraryURL: libraryURL1,
	}); err != nil {
		t.Fatal(err)
	}
	if user, err := FetchUserFromName(db, "name1"); err != nil || user.LibraryURL != libraryURL1 {
		t.Fatal(err, user)
	}
	if user, err := FetchUserFromName(db, "name2"); err != nil || user.LibraryURL != "" {
		t.Fatal(err, user)
	}

	if err := UpdateUser(db, User{
		Name:       "name1",
		UID:        "id1",
		LibraryURL: "",
	}); err != nil {
		t.Fatal(err)
	}
	if err := UpdateUser(db, User{
		Name:       "name2",
		UID:        "id2",
		LibraryURL: libraryURL2,
	}); err != nil {
		t.Fatal(err)
	}
	if user, err := FetchUserFromName(db, "name1"); err != nil || user.LibraryURL != "" {
		t.Fatal(err, user)
	}
	if user, err := FetchUserFromName(db, "name2"); err != nil || user.LibraryURL != libraryURL2 {
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
		LibraryURL: libraryURL,
	}); err == nil {
		t.Fatal("UpdateUser is succeeded with invalid URL")
	} else {
		t.Log(err)
	}
}
