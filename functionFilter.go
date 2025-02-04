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
		return len(b), err
	}
	return c.Conn.Write(b2)
}

// Read reads data from the underlying connection and modifies the bytes according to c.Filter
func (c *FunctionConnFilter) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if err != nil {
		return n, err
	}
	b2, err := c.ReadFilter(b)
	if err != nil {
		return n, err
	}
	copy(b, b2)
	return len(b2), nil
}

// NewFunctionConnFilter creates a new FunctionConnFilter that has the powerful ability to rewrite any byte that comes across the net.Conn with a user-defined function. By default a no-op function.
func NewFunctionConnFilter(parentConn net.Conn, Function string) (net.Conn, error) {
	return &FunctionConnFilter{
		Conn:        parentConn,
		ReadFilter:  noopReadFilter,
		WriteFilter: noopWriteFilter,
	}, nil
}
