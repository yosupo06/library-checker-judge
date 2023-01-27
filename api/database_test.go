package main

import (
	"testing"
)

func TestRegisterUser(t *testing.T) {
	db := createTestDB(t)

	for _, user := range []string{"a", "Bb", "1234"} {
		err := registerUser(db, user, "password", false)

		if err != nil {
			t.Fatalf("failed to create user %v, %v", user, err)
		}
	}

	for _, user := range []string{"", "a a", "b!", "c "} {
		err := registerUser(db, user, "password", false)

		if err == nil {
			t.Fatalf("failed to reject invalid user: %v", user)
		}
	}
}
