package filter

import (
	"bytes"
	"errors"
	"net"
)

var ErrInvalidFilter = errors.New("target and replacement must have the same length")

type connFilter struct {
	net.Conn
	targets      []string
	replacements []string
}

func (c *connFilter) Read(b []byte) (n int, err error) {
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

// NewConnFilter creates a new connFilter that replaces occurrences of target strings with replacement strings in the data read from the connection.
// It returns an error if the lengths of target and replacement slices are not equal.
func NewConnFilter(parentConn net.Conn, targets []string, replacements []string) (net.Conn, error) {
	if len(targets) != len(replacements) {
		return nil, ErrInvalidFilter
	}
	return &connFilter{
		Conn:         parentConn,
		targets:      targets,
		replacements: replacements,
	}, nil
}
