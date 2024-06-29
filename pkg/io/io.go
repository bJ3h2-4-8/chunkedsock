package io

import "io"

type reWrCloser struct {
	r io.Reader
	w io.Writer
	c io.Closer
}

var _ io.ReadWriteCloser = &reWrCloser{}

func (rwc reWrCloser) Read(p []byte) (n int, err error) {
	return rwc.r.Read(p)
}

func (rwc reWrCloser) Write(p []byte) (n int, err error) {
	return rwc.w.Write(p)
}

func (rwc reWrCloser) Close() error {
	if rwc.c == nil {
		return nil
	}
	return rwc.c.Close()
}

// NewIoRemixer creates a new io.ReadWriteCloser from a reader, writer and closer
func NewIoRemixer(r io.Reader, w io.Writer, c io.Closer) io.ReadWriteCloser {
	if r == nil || w == nil {
		return nil
	}

	return &reWrCloser{
		r: r,
		w: w,
		c: c,
	}
}

func NewIoRemixerNoClose(r io.Reader, w io.Writer) io.ReadWriteCloser {
	return NewIoRemixer(r, w, nil)
}

type LimitedWriter struct {
	w    io.Writer
	n    int64
	easy bool
}

func (lw *LimitedWriter) Write(p []byte) (n int, err error) {
	if lw.n <= 0 {
		return 0, io.EOF
	}

	if !lw.easy && int64(len(p)) > lw.n {
		p = p[:lw.n]
	}

	n, err = lw.w.Write(p)
	lw.n -= int64(n)
	return
}

func (lw *LimitedWriter) Ease() {
	lw.easy = true
}

func LimitWriter(w io.Writer, n int64) *LimitedWriter {
	return &LimitedWriter{
		w: w,
		n: n,
	}
}

// LimitIo limits the read and write operations on a stream
func LimitIo(stream io.ReadWriteCloser, read, write int64) io.ReadWriteCloser {
	return NewIoRemixer(
		io.LimitReader(stream, read),
		LimitWriter(stream, write),
		stream,
	)
}
