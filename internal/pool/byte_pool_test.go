package pool

import (
	"crypto/rand"
	"fmt"
	"io"
	"testing"
)

func useSlice(buf []byte) {
	rand.Read(buf)
	fmt.Fprintln(io.Discard, buf)
}

func BenchmarkBuffer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buffer := GetBufferSized(1024)
		useSlice(buffer.Bytes())
		PutBuffer(buffer)
	}
}

func BenchmarkSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 1024)
		useSlice(buf)
		buf = nil
	}
}
