package filter

import (
	"bytes"
	"errors"
	"net"
)

var ErrInvalidFilter = errors.New("target and replacement must have the same length")

type ConnFilter struct {
	net.Conn
	targets      []string
	replacements []string
}

// Read reads data from the underlying connection and replaces all occurrences of target strings
// with their corresponding replacement strings. The replacements are made sequentially for each
// target-replacement pair. The modified data is then copied back to the provided buffer.
func (c *ConnFilter) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if err != nil {
		return n, err
	}

	// Replace all occurrences of target with replacement
	var buffer bytes.Buffer
	buffer.Write(b[:n])
	for i := range c.targets {
		buffer.Reset()
		buffer.Write(bytes.ReplaceAll(b[:n], []byte(c.targets[i]), []byte(c.replacements[i])))
	}
	copy(b, buffer.Bytes())
	return buffer.Len(), nil
}

// Write writes the data to the underlying connection after replacing all occurrences of target strings
// with their corresponding replacement strings. The replacements are made sequentially for each
// target-replacement pair.
func (c *ConnFilter) Write(b []byte) (n int, err error) {
	for i := range c.targets {
		b = bytes.ReplaceAll(b, []byte(c.targets[i]), []byte(c.replacements[i]))
	}
	return c.Conn.Write(b)
}

// NewConnFilter creates a new ConnFilter that replaces occurrences of target strings with replacement strings in the data read from the connection.
// It returns an error if the lengths of target and replacement slices are not equal.
func NewConnFilter(parentConn net.Conn, targets, replacements []string) (net.Conn, error) {
	if len(targets) != len(replacements) {
		return nil, ErrInvalidFilter
	}
	return &ConnFilter{
		Conn:         parentConn,
		targets:      targets,
		replacements: replacements,
	}, nil
}
