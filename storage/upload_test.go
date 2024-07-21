package storage

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestJoinHashes(t *testing.T) {
	hash1 := fmt.Sprintf("%x", sha256.Sum256([]byte("abc")))
	hash2 := fmt.Sprintf("%x", sha256.Sum256([]byte("bca")))
	hash3 := fmt.Sprintf("%x", sha256.Sum256([]byte("cab")))

	x := joinHashes([]string{hash1, hash2, hash3})
	y := joinHashes([]string{hash2, hash3, hash1})
	z := joinHashes([]string{hash3, hash1, hash2})

	if x != y || y != z {
		t.Fatal("hash differ:", x, y, z)
	}
}
