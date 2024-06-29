package io

import (
	"context"
	"errors"
	"io"

	"github.com/bJ3h2-4-8/chunkedsock/pkg/logging"
)

func copyReader(ctx context.Context, cancel context.CancelFunc, log logging.Logger, name string, reader io.Reader, blocksize int) chan []byte {

	ch := make(chan []byte)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("%s: failed reading %v", name, err)
			}

			close(ch)
			cancel()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			buf := make([]byte, blocksize)
			n, err := reader.Read(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
			}

			if n == 0 {
				return
			}
			ch <- buf[:n]
		}
	}()

	return ch
}

func AsyncCopy(ctx context.Context, log logging.Logger, name string, reader io.Reader, writer io.Writer, blocksize int) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	chRead := copyReader(ctx, cancel, log, name+"-asyncReader", reader, blocksize)

	for {
		select {
		case <-ctx.Done():
		case buf, ok := <-chRead:
			if !ok {
				return
			}
			n, err := writer.Write(buf)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					log.Errorf("%s: failed to write: %v", name, err)
				}
				return
			}
			if n != len(buf) {
				log.Errorf("%s: partial write, abort", name)
				return
			}
		}
	}

}
