package langs

// Writer that stores string at most N bytes
type LimitedWriter struct {
	N        int
	data     []byte
	overflow bool
}

const (
	STRIPPED_MESSAGE = " ... stripped"
)

func NewLimitedWriter(n int) *LimitedWriter {
	return &LimitedWriter{
		N: n,
	}
}

func (w *LimitedWriter) Write(b []byte) (n int, err error) {
	blen := len(b)
	cap := w.N - len(w.data)

	add := blen
	if cap < add {
		add = cap
		w.overflow = true
	}
	w.data = append(w.data, b[:add]...)
	return blen, nil
}

func (w *LimitedWriter) Bytes() []byte {
	d := w.data
	if w.overflow {
		l := len([]byte(STRIPPED_MESSAGE))
		if l > w.N {
			copy(w.data[w.N-l:], []byte(STRIPPED_MESSAGE))
		}
	}
	return d
}