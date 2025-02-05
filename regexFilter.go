package filter

import (
	"bytes"
	"errors"
	"net"
	"regexp"
)

var ErrInvalidRegexFilter = errors.New("invalid regex filter")

type RegexConnFilter struct {
	FunctionConnFilter
	match string
}

// Read reads data from the underlying connection and replaces all occurrences of target regex
// with empty strings. The modified data is then copied back to the provided buffer.
func (c *RegexConnFilter) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if err != nil {
		return n, err
	}
	// Replace all occurrences of regex `match` with nothing
	var buffer bytes.Buffer
	return buffer.Len(), nil
}

// NewRegexConnFilter creates a new RegexConnFilter that replaces occurrences of target regex with empty strings in the data read from the connection.
// target regex is c.match
func (c *RegexConnFilter) ReadFilter(b []byte) ([]byte, error) {
	if c.match == "" {
		return b, nil
	}
	re := regexp.MustCompile(c.match)
	s := re.ReplaceAllString(string(b), ``)

	return []byte(s), nil
}

// WriteFilter creates a new RegexConnFilter that replaces occurrences of target regex with empty strings in the data read from the connection.
// target regex is c.match
func (c *RegexConnFilter) WriteFilter(b []byte) ([]byte, error) {
	if c.match == "" {
		return b, nil
	}
	re := regexp.MustCompile(c.match)
	s := re.ReplaceAllString(string(b), ``)

	return []byte(s), nil
}

// NewRegexConnFilter creates a new RegexConnFilter that replaces occurrences of target regex with empty strings in the data read from the connection.
// It returns an error if the lengths of target and replacement slices are not equal.
func NewRegexConnFilter(parentConn net.Conn, regex string) (net.Conn, error) {
	return &RegexConnFilter{
		FunctionConnFilter: FunctionConnFilter{
			Conn: parentConn,
		},
	}, nil
}
