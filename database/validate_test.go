package database

import (
	"strings"
	"testing"
)

func TestUserNameValidate(t *testing.T) {
	type Param struct {
		Name string `validate:"username"`
	}

	for _, name := range []string{"a", "Bb", "1234", "a_a"} {
		if err := validate.Struct(&Param{
			Name: name,
		}); err != nil {
			t.Fatalf("%v should be valid user name: %v", name, err)
		}
	}

	for _, name := range []string{"a a", "", "@", " ", strings.Repeat("a", 1000)} {
		if err := validate.Struct(&Param{
			Name: name,
		}); err == nil {
			t.Fatalf("%v should not be valid user name", name)
		}
	}
}

func TestLibraryURLValidate(t *testing.T) {
	type Param struct {
		LibraryURL string `validate:"libraryURL"`
	}

	for _, url := range []string{"https://judge.yosupo.com", "", "https://" + strings.Repeat("a", 10) + ".com"} {
		if err := validate.Struct(&Param{
			LibraryURL: url,
		}); err != nil {
			t.Fatalf("%v should be valid library URL: %v", url, err)
		}
	}

	for _, url := range []string{"a a", "@", " ", "https://" + strings.Repeat("a", 1000) + ".com"} {
		if err := validate.Struct(&Param{
			LibraryURL: url,
		}); err == nil {
			t.Fatalf("%v should not be valid library URL", url)
		}
	}
}
