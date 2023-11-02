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

func TestUpdateUserWithUID(t *testing.T) {
	const libraryURL = "https://library.yosupo.com"

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
	if user, err := FetchUser(db, "name"); err != nil || user.UID != "id" || user.LibraryURL != libraryURL {
		t.Fatal(err, user)
	}
}

func TestUpdateUserWithoutUID(t *testing.T) {
	const libraryURL = "https://library.yosupo.com"

	db := createTestDB(t)

	if err := RegisterUser(db, "name", "id"); err != nil {
		t.Fatal(err)
	}

	if err := UpdateUser(db, User{
		Name:       "name",
		LibraryURL: libraryURL,
	}); err != nil {
		t.Fatal(err)
	}
	if user, err := FetchUser(db, "name"); err != nil || user.UID != "id" || user.LibraryURL != libraryURL {
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
		LibraryURL: libraryURL1,
	}); err != nil {
		t.Fatal(err)
	}
	if user, err := FetchUser(db, "name1"); err != nil || user.LibraryURL != libraryURL1 {
		t.Fatal(err, user)
	}
	if user, err := FetchUser(db, "name2"); err != nil || user.LibraryURL != "" {
		t.Fatal(err, user)
	}

	if err := UpdateUser(db, User{
		Name:       "name1",
		LibraryURL: "",
	}); err != nil {
		t.Fatal(err)
	}
	if err := UpdateUser(db, User{
		Name:       "name2",
		LibraryURL: libraryURL2,
	}); err != nil {
		t.Fatal(err)
	}
	if user, err := FetchUser(db, "name1"); err != nil || user.LibraryURL != "" {
		t.Fatal(err, user)
	}
	if user, err := FetchUser(db, "name2"); err != nil || user.LibraryURL != libraryURL2 {
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
