package database

import "testing"

func TestMetadata(t *testing.T) {
	db := CreateTestDB(t)

	if _, err := FetchMetadata(db, "key1"); err == nil {
		t.Fatal("fetch succeeded")
	}

	if err := SaveMetadata(db, "key1", "value1"); err != nil {
		t.Fatal(err)
	}

	if err := SaveMetadata(db, "key2", "value2"); err != nil {
		t.Fatal(err)
	}

	if v, err := FetchMetadata(db, "key1"); err != nil || *v != "value1" {
		t.Fatal(*v, err)
	}

	if v, err := FetchMetadata(db, "key2"); err != nil || *v != "value2" {
		t.Fatal(*v, err)
	}

	if err := SaveMetadata(db, "key1", "value3"); err != nil {
		t.Fatal(err)
	}

	if v, err := FetchMetadata(db, "key1"); err != nil || *v != "value3" {
		t.Fatal(*v, err)
	}
}
