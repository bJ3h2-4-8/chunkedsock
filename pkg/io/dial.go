package io

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/bJ3h2-4-8/chunkedsock/pkg/logging"
)

type ChunkResult struct {
	ChunkedSocket *ChunkedSocket
	Err           error
}

type ChunkedSocket struct {
	src, stream io.ReadWriteCloser

	log logging.Logger

	chunkSize,
	blocksize int
}

func NewChunkedSocket(log logging.Logger, src, stream io.ReadWriteCloser, chunksize, blocksize int) *ChunkedSocket {
	if stream == nil {
		return nil
	}
	return &ChunkedSocket{
		src:       src,
		stream:    stream,
		log:       log,
		chunkSize: chunksize,
		blocksize: blocksize,
	}
}

func (cs *ChunkedSocket) StartTransfer(ctx context.Context, name string) {
	reader := io.LimitReader(cs.stream, int64(cs.chunkSize))

	writer := LimitWriter(cs.stream, int64(cs.chunkSize))
	writer.Ease()

	stream := NewIoRemixer(
		reader,
		writer,
		cs.stream,
	)
	ABtransfer(ctx, cs.log, name, cs.src, stream, cs.blocksize)
}

func NewChunkedDial(log logging.Logger, chAddress chan string, timeout time.Duration, src io.ReadWriteCloser, chunksize, blocksize int) chan ChunkResult {
	if chAddress == nil {
		return nil
	}

	chChunkResult := make(chan ChunkResult, 1)

	go func() {
		defer close(chChunkResult)

		for {
			address, ok := <-chAddress
			if !ok {
				return
			}

			// dial to address
			dialer := net.Dialer{
				Timeout: timeout,
			}

			conn, err := dialer.Dial("tcp", address)

			chChunkResult <- ChunkResult{
				ChunkedSocket: NewChunkedSocket(log, src, conn, chunksize, blocksize),
				Err:           err,
			}
		}

	}()

	return chChunkResult
}
