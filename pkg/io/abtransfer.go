package io

import (
	"context"
	"io"
	"sync"

	"github.com/bJ3h2-4-8/chunkedsock/pkg/logging"
)

func ABtransfer(ctx context.Context, log logging.Logger, name string, a, b io.ReadWriteCloser, blocksize int) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if blocksize < 1 {
		blocksize = 1024 * 10
	}

	var wg sync.WaitGroup
	wg.Add(2)

	transfer := func(name string, reader io.Reader, writer io.Writer) {
		defer wg.Done()
		defer cancel()

		AsyncCopy(ctx, log, name, reader, writer, blocksize)
	}

	go transfer("ab"+name, a, b)
	go transfer("ba"+name, b, a)

	<-ctx.Done()
	_ = a.Close()
	_ = b.Close()

	wg.Wait()

}
