package main

import (
	"bytes"
	"testing"
)

func TestOutputStripper(t *testing.T) {
	shortStr := []byte("short string")

	os := &outputStripper{N: 100}
	_, err := os.Write(shortStr)
	if err != nil {
		t.Fatal("outputStripper Error ", err)
	}
	res := os.Bytes()
	if !bytes.Equal(shortStr, res) {
		t.Fatal("outputStripper Differ")
	}
}

func TestOutputStripperLong(t *testing.T) {
	longStrBase := []byte("long string")
	longStr := []byte{}
	for i := 0; i < 100; i++ {
		longStr = append(longStr, longStrBase...)
	}

	os := &outputStripper{N: 100}
	_, err := os.Write(longStr)
	if err != nil {
		t.Fatal("outputStripper Error ", err)
	}
	res := os.Bytes()
	if len(res) > 100 {
		t.Fatal("outputStripper Differ")
	}
	t.Log(string(res))
}
