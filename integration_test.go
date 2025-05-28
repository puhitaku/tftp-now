package main_test

import (
	"bytes"
	"crypto/rand"
	"net"
	"os"
	"testing"

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

	cli, err := tftp.NewClient(addr)
	if err != nil {
		t.Fatalf("failed to create a client: %s", err)
	}

	defer os.Remove("testfile")

	const (
		blockLenDefault = 512
		blockLenFitMTU  = 1428 // From RFC 2348. MTU 1500 - 72
		blockLenJumbo   = 8928 // Jumbo frame. MTU 9000 - 72
	)

	conditions := []struct {
		Name           string
		ClientBlockLen int
		ServerBlockLen int
	}{
		{Name: "C=default,S=default", ClientBlockLen: blockLenDefault, ServerBlockLen: blockLenDefault},
		{Name: "C=default,S=fit MTU", ClientBlockLen: blockLenDefault, ServerBlockLen: blockLenFitMTU},
		{Name: "C=default,S=jumbo", ClientBlockLen: blockLenDefault, ServerBlockLen: blockLenJumbo},
		{Name: "C=fit MTU,S=default", ClientBlockLen: blockLenFitMTU, ServerBlockLen: blockLenDefault},
		{Name: "C=fit MTU,S=fit MTU", ClientBlockLen: blockLenFitMTU, ServerBlockLen: blockLenFitMTU},
		{Name: "C=fit MTU,S=jumbo", ClientBlockLen: blockLenFitMTU, ServerBlockLen: blockLenJumbo},
		{Name: "C=jumbo,S=default", ClientBlockLen: blockLenJumbo, ServerBlockLen: blockLenDefault},
		{Name: "C=jumbo,S=fit MTU", ClientBlockLen: blockLenJumbo, ServerBlockLen: blockLenFitMTU},
		{Name: "C=jumbo,S=jumbo", ClientBlockLen: blockLenJumbo, ServerBlockLen: blockLenJumbo},
	}

	for _, condition := range conditions {
		t.Run(condition.Name, func(t *testing.T) {
			svr.SetBlockSize(condition.ServerBlockLen)
			cli.SetBlockSize(condition.ClientBlockLen)

			rf, err := cli.Send("testfile", "octet")
			if err != nil {
				t.Fatalf("failed to start sending: %s", err)
			}

			body := make([]byte, 65536) // Covers the TFTP's block size max limit 65464
			_, _ = rand.Read(body)
			_, err = rf.ReadFrom(bytes.NewReader(body))
			if err != nil {
				t.Fatalf("failed to send: %s", err)
			}
			defer os.Remove("testfile")

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
		})
	}

	svr.Shutdown()
	err = eg.Wait()
	if err != nil {
		t.Fatalf("server returned an error: %s", err)
	}
}
