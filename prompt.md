Project Path: /home/idk/go/src/github.com/go-i2p/go-connfilter/irc

Source Tree:

```
irc
├── example
│   └── example.go
├── ircinspector.go
└── types.go

```

`/home/idk/go/src/github.com/go-i2p/go-connfilter/irc/example/example.go`:

```````go
package main

import (
	"log"
	"net"

	ircinspector "github.com/go-i2p/go-connfilter/irc"
)

func main() {
	listener, err := net.Listen("tcp", ":6667")
	if err != nil {
		log.Fatal(err)
	}

	inspector := ircinspector.New(listener, ircinspector.Config{
		OnMessage: func(msg *ircinspector.Message) error {
			log.Printf("Received message: %s", msg.Raw)
			return nil
		},
		OnNumeric: func(numeric int, msg *ircinspector.Message) error {
			log.Printf("Received numeric response: %d", numeric)
			return nil
		},
	})

	inspector.AddFilter(ircinspector.Filter{
		Command: "PRIVMSG",
		Channel: "#mychannel",
		Callback: func(msg *ircinspector.Message) error {
			msg.Trailing = "[modified] " + msg.Trailing
			return nil
		},
	})

	for {
		conn, err := inspector.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go HandleConnection(conn)
	}
}

func HandleConnection(conn net.Conn) {
	panic("unimplemented")
}

```````

`/home/idk/go/src/github.com/go-i2p/go-connfilter/irc/ircinspector.go`:

```````go
package ircinspector

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type defaultLogger struct{}

// Debug implements Logger.
func (d *defaultLogger) Debug(format string, args ...interface{}) {
	log.Printf("DBG:"+format, args...)
}

// Error implements Logger.
func (d *defaultLogger) Error(format string, args ...interface{}) {
	log.Printf("ERR:"+format, args...)
}

// New creates a new IRC inspector wrapping an existing listener
func New(listener net.Listener, config Config) *Inspector {
	if config.Logger == nil {
		config.Logger = &defaultLogger{}
	}

	return &Inspector{
		listener: listener,
		config:   config,
		filters:  make([]Filter, 0),
		mu:       sync.RWMutex{},
	}
}

// Accept implements net.Listener Accept method
func (i *Inspector) Accept() (net.Conn, error) {
	conn, err := i.listener.Accept()
	if err != nil {
		return nil, err
	}

	return &ircConn{
		Conn:      conn,
		inspector: i,
	}, nil
}

// Close implements net.Listener Close method
func (i *Inspector) Close() error {
	return i.listener.Close()
}

// Addr implements net.Listener Addr method
func (i *Inspector) Addr() net.Addr {
	return i.listener.Addr()
}

type ircConn struct {
	net.Conn
	inspector *Inspector
	reader    *bufio.Reader
	writer    *bufio.Writer
}

func (c *ircConn) Read(b []byte) (n int, err error) {
	if c.reader == nil {
		c.reader = bufio.NewReader(c.Conn)
	}

	line, err := c.reader.ReadString('\n')
	if err != nil {
		return 0, err
	}

	msg, err := parseMessage(line)
	if err != nil {
		c.inspector.config.Logger.Error("parse error: %v", err)
		copy(b, line)
		return len(line), nil
	}

	if err := c.inspector.processMessage(msg); err != nil {
		c.inspector.config.Logger.Error("process error: %v", err)
	}

	modified := msg.String()
	copy(b, modified)
	return len(modified), nil
}

func (c *ircConn) Write(b []byte) (n int, err error) {
	if c.writer == nil {
		c.writer = bufio.NewWriter(c.Conn)
		defer c.writer.Flush()
	}
	msg, err := parseMessage(string(b))
	if err != nil {
		return c.writer.Write(b)
	}

	defer c.writer.Flush()

	if err := c.inspector.processMessage(msg); err != nil {
		c.inspector.config.Logger.Error("process error: %v", err)
	}

	return c.writer.Write([]byte(msg.String()))
}

func parseMessage(raw string) (*Message, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("empty message")
	}

	msg := &Message{Raw: raw}

	if raw[0] == ':' {
		parts := strings.SplitN(raw[1:], " ", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid message format")
		}
		msg.Prefix = parts[0]
		raw = parts[1]
	}

	parts := strings.SplitN(raw, " :", 2)
	if len(parts) > 1 {
		msg.Trailing = parts[1]
	}

	words := strings.Fields(parts[0])
	if len(words) == 0 {
		return nil, fmt.Errorf("no command found")
	}

	msg.Command = words[0]
	if len(words) > 1 {
		msg.Params = words[1:]
	}

	return msg, nil
}

func (i *Inspector) AddFilter(filter Filter) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.filters = append(i.filters, filter)
}

// parseNumeric converts a 3-character IRC command string to its numeric equivalent.
// It returns an error if the command length is not exactly 3 characters.
func parseNumeric(command string) (int, error) {
	if len(command) != 3 {
		return 0, fmt.Errorf("invalid command length")
	}
	for _, char := range command {
		if char < '0' || char > '9' {
			return 0, fmt.Errorf("command contains non-numeric characters")
		}
	}
	return int(command[0]-'0')*100 + int(command[1]-'0')*10 + int(command[2]-'0'), nil
}

func (i *Inspector) processMessage(msg *Message) error {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// Process global message handler
	if i.config.OnMessage != nil {
		if err := i.config.OnMessage(msg); err != nil {
			return err
		}
	}

	// Process numeric responses
	if numeric, err := parseNumeric(msg.Command); err == nil && i.config.OnNumeric != nil {
		if err := i.config.OnNumeric(numeric, msg); err != nil {
			return err
		}
	}

	// Process filters
	for _, filter := range i.filters {
		if filter.Command == "" || filter.Command == msg.Command {
			if err := filter.Callback(msg); err != nil {
				return err
			}
		}
	}

	return nil
}

```````

`/home/idk/go/src/github.com/go-i2p/go-connfilter/irc/types.go`:

```````go
package ircinspector

import (
	"net"
	"strings"
	"sync"
)

// Message represents a parsed IRC message
type Message struct {
	Raw      string
	Prefix   string
	Command  string
	Params   []string
	Trailing string
}

func (m *Message) String() string {
	var parts []string
	if m.Prefix != "" {
		parts = append(parts, ":"+m.Prefix)
	}
	parts = append(parts, m.Command)
	if len(m.Params) > 0 {
		parts = append(parts, strings.Join(m.Params, " "))
	}
	if m.Trailing != "" {
		parts = append(parts, ":"+m.Trailing)
	}
	return strings.Join(parts, " ") + "\r\n"
}

// Filter defines criteria for message filtering
type Filter struct {
	Command  string
	Channel  string
	Prefix   string
	Callback func(*Message) error
}

// Config contains inspector configuration
type Config struct {
	OnMessage func(*Message) error
	OnNumeric func(int, *Message) error
	Logger    Logger
}

// Logger interface for customizable logging
type Logger interface {
	Debug(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// Inspector implements the net.Listener interface with IRC inspection
type Inspector struct {
	listener net.Listener
	config   Config
	filters  []Filter
	mu       sync.RWMutex
}

```````