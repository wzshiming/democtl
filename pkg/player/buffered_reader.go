package player

import (
	"bytes"
	"context"
	"io"
	"sync"
	"time"
)

type syncBuffer struct {
	buf     *bytes.Buffer
	mu      sync.Mutex
	updated chan struct{}
}

func newSyncBuffer() *syncBuffer {
	return &syncBuffer{
		updated: make(chan struct{}, 1),
		buf:     bytes.NewBuffer(nil),
	}
}

func (b *syncBuffer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}

	b.mu.Lock()
	defer func() {
		select {
		case b.updated <- struct{}{}:
		default:
		}
		b.mu.Unlock()
	}()

	return b.buf.Write(p)
}

func (b *syncBuffer) Read(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Read(p)
}

func (b *syncBuffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	close(b.updated)
	return nil
}

func (b *syncBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Len()
}

func (b *syncBuffer) Updated() <-chan struct{} {
	select {
	case <-b.updated:
	default:
	}
	return b.updated
}

type bufferedReader struct {
	source io.Reader
	buffer *syncBuffer
}

func newBufferedReader(r io.Reader) *bufferedReader {
	return &bufferedReader{
		source: r,
		buffer: newSyncBuffer(),
	}
}

func (r *bufferedReader) Run() {
	io.Copy(r.buffer, r.source)
	r.buffer.Close()
}

func (r *bufferedReader) ReadWithTimeout(p []byte, timeout time.Duration) (n int, err error) {
	if r.buffer.Len() > 0 {
		return r.buffer.Read(p)
	}

	select {
	case _, ok := <-r.buffer.Updated():
		if !ok {
			return 0, io.EOF
		}
	case <-time.After(timeout):
		return 0, context.DeadlineExceeded
	}
	return r.buffer.Read(p)
}
