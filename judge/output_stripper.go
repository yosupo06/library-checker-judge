package main

import (
	"bytes"
	"errors"
)

// Save stripped output with strip()
type outputStripper struct {
	N        int
	data     []byte
	overflow bool
}

func (os *outputStripper) Write(b []byte) (n int, err error) {
	if os.N <= 20 {
		return -1, errors.New("N is too small")
	}
	blen := len(b)
	cap := os.N - 20 - len(os.data)

	add := blen
	if cap < add {
		add = cap
		os.overflow = true
	}
	os.data = append(os.data, b[:add]...)
	return blen, nil
}

func (os *outputStripper) Bytes() []byte {
	var buf bytes.Buffer
	buf.Write(os.data)
	if os.overflow {
		buf.Write([]byte(" ... stripped"))
	}
	return buf.Bytes()
}
