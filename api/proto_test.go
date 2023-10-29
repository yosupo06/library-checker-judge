package main

import (
	"testing"
	"time"
)

func TestToProtoTimestamp(t *testing.T) {
	if toProtoTimestamp(time.Time{}) != nil {
		t.Fatal("toProtoTimestamp(time.Time{}) should returns default value")
	}
}
