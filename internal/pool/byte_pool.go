package pool

import (
	"bytes"
	"sync"
)

var bufPool = &sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

// GetBuffer returns a buffer whose len == 0 from the pool.
func GetEmptyBuffer() *bytes.Buffer {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// GetBufferSized returns a buffer with len == sized. If the len of the
// buffer is lower than sized, only the difference is added.
func GetBufferSized(size int) *bytes.Buffer {
	buf := bufPool.Get().(*bytes.Buffer)
	if size > buf.Len() {
		diff := size - buf.Len()
		buf.Write(make([]byte, diff))
	}
	return buf
}

func PutBuffer(buf *bytes.Buffer) { bufPool.Put(buf) }
