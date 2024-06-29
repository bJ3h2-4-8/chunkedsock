package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	chunkedIo "github.com/bJ3h2-4-8/chunkedsock/pkg/io"
	"github.com/bJ3h2-4-8/chunkedsock/pkg/logging"
)

func createDialer(log logging.Logger, sourceConn io.ReadWriteCloser, targetAddress string, chunkSizeKb uint) func() (*chunkedIo.ChunkedSocket, error) {

	chAddr := make(chan string)
	chunkedDial := chunkedIo.NewChunkedDial(log, chAddr, 20*time.Second, sourceConn, int(chunkSizeKb)*1024, 1024*10)

	return func() (*chunkedIo.ChunkedSocket, error) {
		chAddr <- targetAddress
		res := <-chunkedDial
		return res.ChunkedSocket, res.Err
	}
}

func createListenSocket(log logging.Logger, addr string) (io.ReadWriteCloser, error) {

	log.Infof("listening on %s", addr)
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := listen.Accept()
	if err != nil {
		return nil, err
	}
	log.Infof("accepted connection from %s", conn.RemoteAddr())
	return chunkedIo.NewIoRemixerNoClose(conn, conn), nil
}

func main() {
	var listenAddress string
	flag.StringVar(&listenAddress, "listen", "", "this ip address")
	flag.StringVar(&listenAddress, "l", "", "this ip address (short)")

	var targetAddress string
	flag.StringVar(&targetAddress, "target", "", "destination tcp address")
	flag.StringVar(&targetAddress, "t", "", "destination tcp address (short)")

	var chunkSizeKb uint
	flag.UintVar(&chunkSizeKb, "chunksize", 0, "chunk size in kilobytes")
	flag.UintVar(&chunkSizeKb, "c", 0, "chunk size in kilobytes")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage of %s:\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "it opens a listen socket and reconnects to the target after chunk bytes.\n")

		flag.PrintDefaults()
	}

	flag.Parse()

	if listenAddress == "" {
		listenAddress = ":33455"
	}

	if targetAddress == "" {
		flag.Usage()
		os.Exit(1)
	}

	if chunkSizeKb == 0 {
		chunkSizeKb = 200
	}

	log := logging.DefaultLogger

	// attach a listen socket and let it accept a connection
	sourceConn, err := createListenSocket(log, listenAddress)
	if err != nil {
		log.Fatalf("failed to create listen socket: %v", err)
	}

	// create a chunked dialer
	var dialer = createDialer(log, sourceConn, targetAddress, chunkSizeKb)

	current, err := dialer()
	if err != nil {
		log.Fatalf("failed to dial first connection: %v", err)
	}

	chDone := make(chan struct{})
	for {
		// begin a transfer
		go func(now *chunkedIo.ChunkedSocket) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			now.StartTransfer(ctx, "main")
			chDone <- struct{}{}
		}(current)

		// dial the next in advance
		current, err = dialer()
		if err != nil {
			log.Infof("warn, could not dial in advance - will retry after this transfer: %v", err)
			current = nil
		}

		// wait for the transfer to finish
		<-chDone
		if current == nil {
			current, err = dialer()
			if err != nil {
				log.Fatalf("failed to dial next connection: %v", err)
			}
		}
		log.Infof("transfer finished, next connection is ready")

	}

}
