package gologger

import (
	"bytes"
	"sync"
)

var byteBufferPool *sync.Pool

// GetBufferPool gets back a borrowed buffer from pool
func GetBuffer() *bytes.Buffer {
	return byteBufferPool.Get().(*bytes.Buffer)
}

// PutBufferPool puts back a borrowed buffer to pool
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
