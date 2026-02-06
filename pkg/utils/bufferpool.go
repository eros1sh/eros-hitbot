// Package utils provides high-performance buffer pooling to reduce GC pressure.
package utils

import (
	"bytes"
	"sync"
)

// BufferPool provides a pool of reusable bytes.Buffer objects.
// This significantly reduces memory allocations in high-throughput scenarios.
type BufferPool struct {
	pool sync.Pool
}

// NewBufferPool creates a new buffer pool.
func NewBufferPool() *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

// Get retrieves a buffer from the pool.
// The buffer is reset and ready for use.
func (p *BufferPool) Get() *bytes.Buffer {
	buf := p.pool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// Put returns a buffer to the pool.
func (p *BufferPool) Put(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	// Prevent memory leak: truncate if too large
	if buf.Cap() > 1024*1024 { // 1MB limit
		return // Let GC collect large buffers
	}
	p.pool.Put(buf)
}

// Global instance for package-level convenience
var defaultBufferPool = NewBufferPool()

// GetBuffer gets a buffer from the default pool.
func GetBuffer() *bytes.Buffer {
	return defaultBufferPool.Get()
}

// PutBuffer returns a buffer to the default pool.
func PutBuffer(buf *bytes.Buffer) {
	defaultBufferPool.Put(buf)
}

// BytePool provides a pool of reusable byte slices.
type BytePool struct {
	pool sync.Pool
	size int
}

// NewBytePool creates a new byte slice pool with fixed size.
func NewBytePool(size int) *BytePool {
	return &BytePool{
		pool: sync.Pool{
			New: func() interface{} {
				b := make([]byte, size)
				return &b
			},
		},
		size: size,
	}
}

// Get retrieves a byte slice from the pool.
func (p *BytePool) Get() []byte {
	b := p.pool.Get().(*[]byte)
	return (*b)[:p.size]
}

// Put returns a byte slice to the pool.
func (p *BytePool) Put(b []byte) {
	if cap(b) < p.size {
		return
	}
	p.pool.Put(&b)
}

// StringPool provides a pool of reusable strings (via string builder).
type StringPool struct {
	pool sync.Pool
}

// NewStringPool creates a new string builder pool.
func NewStringPool() *StringPool {
	return &StringPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &stringsBuilderWrapper{}
			},
		},
	}
}

type stringsBuilderWrapper struct {
	buf []byte
}

// Get retrieves a string builder from the pool.
func (p *StringPool) Get() *stringsBuilderWrapper {
	w := p.pool.Get().(*stringsBuilderWrapper)
	w.buf = w.buf[:0] // Reset
	return w
}

// Put returns a string builder to the pool.
func (p *StringPool) Put(w *stringsBuilderWrapper) {
	if cap(w.buf) > 1024*1024 { // 1MB limit
		return // Let GC collect large buffers
	}
	p.pool.Put(w)
}

// Write appends bytes to the builder.
func (w *stringsBuilderWrapper) Write(p []byte) {
	w.buf = append(w.buf, p...)
}

// WriteString appends a string to the builder.
func (w *stringsBuilderWrapper) WriteString(s string) {
	w.buf = append(w.buf, s...)
}

// String returns the accumulated string.
func (w *stringsBuilderWrapper) String() string {
	return string(w.buf)
}

// Len returns the length of the accumulated data.
func (w *stringsBuilderWrapper) Len() int {
	return len(w.buf)
}

// Reset clears the builder.
func (w *stringsBuilderWrapper) Reset() {
	w.buf = w.buf[:0]
}

// ObjectPool is a generic object pool using generics (Go 1.18+).
// For older Go versions, use type-specific pools.
type ObjectPool struct {
	pool sync.Pool
}

// NewObjectPool creates a new object pool with a factory function.
func NewObjectPool(newFunc func() interface{}) *ObjectPool {
	return &ObjectPool{
		pool: sync.Pool{
			New: newFunc,
		},
	}
}

// Get retrieves an object from the pool.
func (p *ObjectPool) Get() interface{} {
	return p.pool.Get()
}

// Put returns an object to the pool.
func (p *ObjectPool) Put(v interface{}) {
	p.pool.Put(v)
}

// LimitPool is a pool with a maximum size to prevent unbounded growth.
type LimitPool struct {
	pool     chan interface{}
	factory  func() interface{}
	maxSize  int
}

// NewLimitPool creates a new limited-size pool.
func NewLimitPool(maxSize int, factory func() interface{}) *LimitPool {
	return &LimitPool{
		pool:    make(chan interface{}, maxSize),
		factory: factory,
		maxSize: maxSize,
	}
}

// Get retrieves an object from the pool or creates a new one.
func (p *LimitPool) Get() interface{} {
	select {
	case v := <-p.pool:
		return v
	default:
		return p.factory()
	}
}

// Put returns an object to the pool (non-blocking).
func (p *LimitPool) Put(v interface{}) {
	select {
	case p.pool <- v:
	default:
		// Pool full, discard object
	}
}
