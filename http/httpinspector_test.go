package httpinspector

import (
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
)

func TestInspector(t *testing.T) {
	// Create a mock listener
	listener := &mockListener{
		conns: make(chan net.Conn, 1),
	}

	// Create inspector with test configuration
	config := Config{
		OnRequest: func(req *http.Request) error {
			req.Header.Set("X-Modified", "true")
			return nil
		},
		OnResponse: func(resp *http.Response) error {
			resp.Header.Set("X-Modified", "true")
			return nil
		},
		LoggingEnabled: true,
	}

	inspector := New(listener, config)
	defer inspector.Close()

	// Test request modification
	t.Run("ModifyRequest", func(t *testing.T) {
		conn := &mockConn{
			readData: []byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"),
		}
		listener.conns <- conn

		inspectedConn, err := inspector.Accept()
		if err != nil {
			t.Fatal(err)
		}

		buf := make([]byte, 1024)
		n, err := inspectedConn.Read(buf)
		if err != nil && err != io.EOF {
			t.Fatal(err)
		}

		if !strings.Contains(string(buf[:n]), "X-Modified: true") {
			t.Error("request modification not applied")
		}
	})

}

// Mock implementations for testing
type mockListener struct {
	conns chan net.Conn
}

func (m *mockListener) Accept() (net.Conn, error) {
	return <-m.conns, nil
}

func (m *mockListener) Close() error {
	close(m.conns)
	return nil
}

func (m *mockListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}
}

type mockConn struct {
	readData []byte
	readPos  int
}

// Implement net.Conn interface methods for mockConn...
