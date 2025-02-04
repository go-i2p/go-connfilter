package ircinspector

import (
	"net"
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
