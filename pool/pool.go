package pool

import (
	"bytes"
	"sync"
)

var (
	minBufLen = 1 * 1024
	maxBufLen = 8 * 1024

	byteBufferPool *sync.Pool
)

func init() {
	byteBufferPool = &sync.Pool{New: func() interface{} {
		return new(bytes.Buffer)
	}}
}

// GetBuffer gets back a borrowed buffer from pool
func GetBuffer() *bytes.Buffer {
	return byteBufferPool.Get().(*bytes.Buffer)
}

// PutBuffer puts back a borrowed buffer to pool
func PutBuffer(buffer *bytes.Buffer) {
	buffer.Reset()
	if buffer.Cap() < minBufLen {
		return
	}
	if buffer.Cap() > maxBufLen {
		return
	}
	byteBufferPool.Put(buffer)
}
