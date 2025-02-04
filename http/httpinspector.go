// Package httpinspector provides HTTP traffic inspection and modification capabilities
// by wrapping the standard net.Listener interface.
package httpinspector

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Common errors returned by the inspector.
var (
	ErrInvalidModification = errors.New("invalid HTTP message modification")
	ErrMalformedHTTP       = errors.New("malformed HTTP message")
	ErrClosedInspector     = errors.New("inspector is closed")
)

// RequestCallback is called for each HTTP request intercepted.
type RequestCallback func(*http.Request) error

// ResponseCallback is called for each HTTP response intercepted.
type ResponseCallback func(*http.Response) error

// Config contains configuration options for the HTTP inspector.
type Config struct {
	OnRequest      RequestCallback  // Called for each request
	OnResponse     ResponseCallback // Called for each response
	LoggingEnabled bool             // Enable debug logging
	ReadTimeout    time.Duration    // Timeout for reading HTTP messages
	ModifyTimeout  time.Duration    // Timeout for modification callbacks
	MaxHeaderBytes int              // Maximum size of HTTP headers
	MaxBodyBytes   int64            // Maximum size of HTTP body
}

// DefaultConfig returns a Config with reasonable defaults.
func DefaultConfig() Config {
	return Config{
		ReadTimeout:    30 * time.Second,
		ModifyTimeout:  5 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
		MaxBodyBytes:   1 << 26, // 64MB
	}
}

// Inspector wraps a net.Listener to provide HTTP traffic inspection.
type Inspector struct {
	listener net.Listener
	config   Config
	closed   bool
	mu       sync.RWMutex // Protects closed field
}

// New creates a new Inspector wrapping the provided listener.
func New(listener net.Listener, config Config) *Inspector {
	return &Inspector{
		listener: listener,
		config:   config,
	}
}

// Accept implements the net.Listener Accept method.
func (i *Inspector) Accept() (net.Conn, error) {
	i.mu.RLock()
	if i.closed {
		i.mu.RUnlock()
		return nil, ErrClosedInspector
	}
	i.mu.RUnlock()

	conn, err := i.listener.Accept()
	if err != nil {
		return nil, err
	}

	return &inspectedConn{
		Conn:   conn,
		config: i.config,
	}, nil
}

// Close implements the net.Listener Close method.
func (i *Inspector) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.closed {
		return ErrClosedInspector
	}

	i.closed = true
	return i.listener.Close()
}

// Addr implements the net.Listener Addr method.
func (i *Inspector) Addr() net.Addr {
	return i.listener.Addr()
}

// inspectedConn wraps a net.Conn to provide HTTP inspection.
type inspectedConn struct {
	net.Conn
	config     Config
	reader     *bufio.Reader
	writer     *bufio.Writer
	readMu     sync.Mutex
	writeMu    sync.Mutex
	firstRead  bool
	firstWrite bool
}

// Read implements the net.Conn Read method with HTTP inspection.
func (c *inspectedConn) Read(b []byte) (int, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	if c.reader == nil {
		c.reader = bufio.NewReader(c.Conn)
	}

	// Only inspect the first read for HTTP requests
	if !c.firstRead && c.config.OnRequest != nil {
		c.firstRead = true
		return c.handleHTTPRequest(b)
	}

	return c.reader.Read(b)
}

// Write implements the net.Conn Write method with HTTP inspection.
func (c *inspectedConn) Write(b []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if c.writer == nil {
		c.writer = bufio.NewWriter(c.Conn)
	}

	// Only inspect the first write for HTTP responses
	if !c.firstWrite && c.config.OnResponse != nil {
		c.firstWrite = true
		return c.handleHTTPResponse(b)
	}

	return c.writer.Write(b)
}

// handleHTTPRequest processes incoming HTTP requests.
func (c *inspectedConn) handleHTTPRequest(b []byte) (int, error) {
	// Peek to verify HTTP request
	peek, err := c.reader.Peek(4)
	if err != nil {
		return 0, err
	}

	// Check for HTTP method
	if !isHTTPMethod(string(peek)) {
		return c.reader.Read(b)
	}

	// Read and parse the request
	req, err := http.ReadRequest(c.reader)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrMalformedHTTP, err)
	}
	defer req.Body.Close()

	// Apply request callback
	if err := c.config.OnRequest(req); err != nil {
		return 0, fmt.Errorf("request modification failed: %w", err)
	}

	// Buffer the modified request
	var buf bytes.Buffer
	if err := req.Write(&buf); err != nil {
		return 0, fmt.Errorf("%w: %v", ErrInvalidModification, err)
	}

	// Copy the modified request to the output buffer
	return copy(b, buf.Bytes()), nil
}

// handleHTTPResponse processes outgoing HTTP responses.
func (c *inspectedConn) handleHTTPResponse(b []byte) (int, error) {
	// Parse the response
	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(b)), nil)
	if err != nil {
		return c.writer.Write(b)
	}
	defer resp.Body.Close()

	// Apply response callback
	if err := c.config.OnResponse(resp); err != nil {
		return 0, fmt.Errorf("response modification failed: %w", err)
	}

	// Buffer the modified response
	var buf bytes.Buffer
	if err := resp.Write(&buf); err != nil {
		return 0, fmt.Errorf("%w: %v", ErrInvalidModification, err)
	}

	// Write the modified response
	return c.writer.Write(buf.Bytes())
}

// isHTTPMethod checks if the given string starts with an HTTP method.
func isHTTPMethod(s string) bool {
	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"}
	for _, method := range methods {
		if strings.HasPrefix(s, method) {
			return true
		}
	}
	return false
}
