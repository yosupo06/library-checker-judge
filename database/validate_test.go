package database

import (
	"testing"
)

func TestUserNameValidate(t *testing.T) {
	type UserNameParam struct {
		UserName string `validate:"username"`
	}

	for _, name := range []string{"a", "Bb", "1234", "a_a"} {
		if err := validate.Struct(&UserNameParam{
			UserName: name,
		}); err != nil {
			t.Fatalf("%v should be valid user name: %v", name, err)
		}
	}

	for _, name := range []string{"a a", "", "@", " "} {
		if err := validate.Struct(&UserNameParam{
			UserName: name,
		}); err == nil {
			t.Fatalf("%v should not be valid user name", name)
		}
	}
}

func TestLibraryURLValidate(t *testing.T) {
	type Param struct {
		LibraryURL string `validate:"libraryURL"`
	}

	for _, url := range []string{"https://judge.yosupo.com", ""} {
		if err := validate.Struct(&Param{
			LibraryURL: url,
		}); err != nil {
			t.Fatalf("%v should be valid library URL: %v", url, err)
		}
	}

	for _, url := range []string{"a a", "@", " "} {
		if err := validate.Struct(&Param{
			LibraryURL: url,
		}); err == nil {
			t.Fatalf("%v should not be valid user name", url)
		}
	}
}
