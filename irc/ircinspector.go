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
