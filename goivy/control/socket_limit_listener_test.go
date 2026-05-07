package control

import (
	"net"
	"testing"
	"time"

	"github.com/glycerine/rate"
)

func TestSocketLimitStateAppliesGlobalAndRemoteLimits(t *testing.T) {
	now := time.Date(2026, 5, 7, 12, 0, 0, 0, time.UTC)
	limits := newSocketLimitState(socketLimitConfig{
		GlobalLimit: rate.Limit(2),
		GlobalBurst: 2,
		RemoteEvery: 2 * time.Second,
	}, func() time.Time { return now })

	if !limits.allowAt("203.0.113.10:5000", now) {
		t.Fatalf("first connection from remote should be allowed")
	}
	if limits.allowAt("203.0.113.10:5000", now.Add(time.Second)) {
		t.Fatalf("same remote IP:port was allowed inside the 2 second window")
	}
	if !limits.allowAt("203.0.113.10:5001", now.Add(time.Second)) {
		t.Fatalf("different remote IP:port should be allowed while global tokens remain")
	}
	if limits.allowAt("203.0.113.10:5002", now.Add(time.Second)) {
		t.Fatalf("third global connection inside one second should be rejected")
	}
	if !limits.allowAt("203.0.113.10:5000", now.Add(2*time.Second)) {
		t.Fatalf("same remote IP:port should be allowed again after 2 seconds")
	}
}

func TestSocketLimitListenerReleasesSlotOnClose(t *testing.T) {
	underlying := &scriptedListener{accepts: make(chan net.Conn, 2)}
	limited := newSocketLimitListener(underlying, socketLimitConfig{
		MaxConcurrent: 1,
		GlobalLimit:   rate.Limit(100),
		GlobalBurst:   100,
		RemoteEvery:   2 * time.Second,
	})
	client1, server1 := net.Pipe()
	client2, server2 := net.Pipe()
	defer client1.Close()
	defer client2.Close()
	underlying.accepts <- &remoteAddrConn{Conn: server1, remote: fakeAddr("203.0.113.10:5000")}

	first, err := limited.Accept()
	if err != nil {
		t.Fatalf("first accept: %v", err)
	}
	secondDone := make(chan error, 1)
	go func() {
		underlying.accepts <- &remoteAddrConn{Conn: server2, remote: fakeAddr("203.0.113.10:5001")}
		second, err := limited.Accept()
		if err == nil {
			_ = second.Close()
		}
		secondDone <- err
	}()

	select {
	case err := <-secondDone:
		t.Fatalf("second accept completed while first socket was still connected: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	if err := first.Close(); err != nil {
		t.Fatalf("close first connection: %v", err)
	}
	select {
	case err := <-secondDone:
		if err != nil {
			t.Fatalf("second accept after release: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatalf("second accept did not proceed after first connection closed")
	}
}

type fakeAddr string

func (a fakeAddr) Network() string { return "tcp" }
func (a fakeAddr) String() string  { return string(a) }

type remoteAddrConn struct {
	net.Conn
	remote net.Addr
}

func (c *remoteAddrConn) RemoteAddr() net.Addr { return c.remote }

type scriptedListener struct {
	accepts chan net.Conn
	closed  bool
}

func (l *scriptedListener) Accept() (net.Conn, error) {
	conn, ok := <-l.accepts
	if !ok {
		return nil, net.ErrClosed
	}
	return conn, nil
}

func (l *scriptedListener) Close() error {
	if !l.closed {
		l.closed = true
		close(l.accepts)
	}
	return nil
}

func (l *scriptedListener) Addr() net.Addr { return fakeAddr("127.0.0.1:0") }
