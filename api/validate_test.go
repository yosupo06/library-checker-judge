package main

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
