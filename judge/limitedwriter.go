package main

import (
	"bytes"
	"errors"
)

// Writer that stores string at most N bytes
type LimitedWriter struct {
	N        int
	data     []byte
	overflow bool
}

func NewLimitedWriter(n int) (*LimitedWriter, error) {
	if n <= 20 {
		return nil, errors.New("n is too small")
	}
	return &LimitedWriter{
		N: n,
	}, nil
}

func (w *LimitedWriter) Write(b []byte) (n int, err error) {
	blen := len(b)
	cap := w.N - 20 - len(w.data)

	add := blen
	if cap < add {
		add = cap
		w.overflow = true
	}
	w.data = append(w.data, b[:add]...)
	return blen, nil
}

func (w *LimitedWriter) Bytes() []byte {
	var buf bytes.Buffer
	buf.Write(w.data)
	if w.overflow {
		buf.Write([]byte(" ... stripped"))
	}
	return buf.Bytes()
}
