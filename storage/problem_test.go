package storage

import (
	_ "embed"
	"os"
	"path"
	"reflect"
	"testing"
)

//go:embed aplusb_info.toml
var infoToml string

func TestParseInfo(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	file, err := os.Create(path.Join(tempDir, "info.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(infoToml); err != nil {
		t.Fatal(err)
	}
	file.Close()

	info, err := ParseInfo(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	if info.Title != "A + B" {
		t.Fatal("info.Title is not expected", info)
	}
	if info.TimeLimit != 2.0 {
		t.Fatal("info.TimeLimit is not expected", info)
	}
	names := info.TestCaseNames()
	if !reflect.DeepEqual(names, []string{
		"example_00", "example_01", "random_00", "random_01", "random_02",
	}) {
		t.Fatal("info.testCaseNames() is not expected", names)
	}
}

func TestTestCasesKey(t *testing.T) {
	p := Problem{
		Name:         "aplusb",
		Version:      "version",
		TestCaseHash: "hash",
	}

	key := p.testCasesKey()
	if key != "v3/aplusb/testcase/hash.tar.gz" {
		t.Fatal("TestCasesKey is differ", key)
	}
}

func TestPublicFileKeyPrefix(t *testing.T) {
	p := Problem{
		Name:         "aplusb",
		Version:      "version",
		TestCaseHash: "hash",
	}

	prefix := p.publicFileKeyPrefix()
	if prefix != "v3/aplusb/files/version" {
		t.Fatal("TestPublicFileKeyPrefix is differ", prefix)
	}
}

func TestPublicFileKey(t *testing.T) {
	p := Problem{
		Name:         "aplusb",
		Version:      "version",
		TestCaseHash: "hash",
	}

	key := p.publicFileKey("key")
	if key != "v3/aplusb/files/version/key" {
		t.Fatal("TestPublicFileKey is not expected", key)
	}
}

func TestPublicTestCaseKey(t *testing.T) {
	p := Problem{
		Name:         "aplusb",
		Version:      "version",
		TestCaseHash: "hash",
	}

	key := p.publicTestCaseKey("in/example_00.in")
	if key != "v3/aplusb/testcase/hash/in/example_00.in" {
		t.Fatal("TestPublicTestCase is not expected", key)
	}
}
