package filter

import (
	"errors"
	"net"
)

var ErrInvalidFunctionFilter = errors.New("invalid Function filter")

func noopReadFilter(b []byte) ([]byte, error) {
	return b, nil
}

func noopWriteFilter(b []byte) ([]byte, error) {
	return b, nil
}

type FunctionConnFilter struct {
	net.Conn
	ReadFilter  func(b []byte) ([]byte, error)
	WriteFilter func(b []byte) ([]byte, error)
}

var ex net.Conn = &FunctionConnFilter{}

// Write modifies the bytes according to c.Filter and writes the result to the underlying connection
func (c *FunctionConnFilter) Write(b []byte) (n int, err error) {
	b2, err := c.WriteFilter(b)
	if err != nil {
		return 0, err
	}
	return c.Conn.Write(b2)
}

// Read reads data from the underlying connection and modifies the bytes according to c.Filter
func (c *FunctionConnFilter) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if err != nil {
		return 0, err
	}
	b2, err := c.ReadFilter(b[:n])
	if err != nil {
		return 0, err
	}
	copy(b, b2)
	return len(b2), nil
}

// NewFunctionConnFilter creates a new FunctionConnFilter that has the powerful ability to rewrite any byte that comes across the net.Conn with user-defined functions. By default, the filters are no-op functions.
func NewFunctionConnFilter(parentConn net.Conn, readFilter, writeFilter func(b []byte) ([]byte, error)) (net.Conn, error) {
	if readFilter == nil {
		readFilter = noopReadFilter
	}
	if writeFilter == nil {
		writeFilter = noopWriteFilter
	}
	return &FunctionConnFilter{
		Conn:        parentConn,
		ReadFilter:  readFilter,
		WriteFilter: writeFilter,
	}, nil
}
