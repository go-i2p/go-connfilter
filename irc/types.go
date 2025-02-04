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
