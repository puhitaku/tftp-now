package main_test

import (
	"bytes"
	"net"
	"os"
	"testing"
	"time"

	"github.com/pin/tftp/v3"
	"github.com/puhitaku/tftp-now/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

var _ = func() any {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	os.Remove("testfile")
	return nil
}()

func TestWriteRead(t *testing.T) {
	eg := errgroup.Group{}
	svr := tftp.NewServer(server.ReadHandler, server.WriteHandler)
	addr := "localhost:10069"

	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		t.Fatalf("failed to resolve addr: %s", err)
	}
	conn, err := net.ListenUDP("udp", a)
	if err != nil {
		t.Fatalf("failed to listen: %s", err)
	}

	eg.Go(func() error {
		return svr.Serve(conn)
	})

	time.Sleep(100 * time.Millisecond)

	cli, err := tftp.NewClient(addr)
	if err != nil {
		t.Fatalf("failed to create a client: %s", err)
	}

	rf, err := cli.Send("testfile", "octet")
	if err != nil {
		t.Fatalf("failed to start sending: %s", err)
	}

	defer os.Remove("testfile")

	body := []byte{0, 1, 2, 3}
	readBuf := bytes.NewBuffer(body)
	_, err = rf.ReadFrom(readBuf)
	if err != nil {
		t.Fatalf("failed to send: %s", err)
	}

	wt, err := cli.Receive("testfile", "octet")
	if err != nil {
		t.Fatalf("failed to start receiving: %s", err)
	}

	writeBuf := bytes.NewBuffer(nil)
	_, err = wt.WriteTo(writeBuf)
	if err != nil {
		t.Fatalf("failed to receive: %s", err)
	}

	if b := writeBuf.Bytes(); !bytes.Equal(body, b) {
		t.Errorf("received data differ, expect: %+v actual: %+v", body, b)
	}

	svr.Shutdown()
	err = eg.Wait()
	if err != nil {
		t.Fatalf("server returned an error: %s", err)
	}
}
