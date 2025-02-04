# filter
--
    import "github.com/go-i2p/go-connfilter"


## Usage

```go
var ErrInvalidFilter = errors.New("target and replacement must have the same length")
```

```go
var ErrInvalidFunctionFilter = errors.New("invalid Function filter")
```

```go
var ErrInvalidRegexFilter = errors.New("invalid regex filter")
```

#### func  NewConnFilter

```go
func NewConnFilter(parentConn net.Conn, targets, replacements []string) (net.Conn, error)
```
NewConnFilter creates a new ConnFilter that replaces occurrences of target
strings with replacement strings in the data read from the connection. It
returns an error if the lengths of target and replacement slices are not equal.

#### func  NewFunctionConnFilter

```go
func NewFunctionConnFilter(parentConn net.Conn, readFilter, writeFilter func(b []byte) ([]byte, error)) (net.Conn, error)
```
NewFunctionConnFilter creates a new FunctionConnFilter that has the powerful
ability to rewrite any byte that comes across the net.Conn with user-defined
functions. By default, the filters are no-op functions.

#### func  NewRegexConnFilter

```go
func NewRegexConnFilter(parentConn net.Conn, regex string) (net.Conn, error)
```
NewRegexConnFilter creates a new RegexConnFilter that replaces occurrences of
target regex with empty strings in the data read from the connection. It returns
an error if the lengths of target and replacement slices are not equal.

#### type ConnFilter

```go
type ConnFilter struct {
	net.Conn
}
```


#### func (*ConnFilter) Read

```go
func (c *ConnFilter) Read(b []byte) (n int, err error)
```
Read reads data from the underlying connection and replaces all occurrences of
target strings with their corresponding replacement strings. The replacements
are made sequentially for each target-replacement pair. The modified data is
then copied back to the provided buffer.

#### func (*ConnFilter) Write

```go
func (c *ConnFilter) Write(b []byte) (n int, err error)
```
Write writes the data to the underlying connection after replacing all
occurrences of target strings with their corresponding replacement strings. The
replacements are made sequentially for each target-replacement pair.

#### type FunctionConnFilter

```go
type FunctionConnFilter struct {
	net.Conn
	ReadFilter  func(b []byte) ([]byte, error)
	WriteFilter func(b []byte) ([]byte, error)
}
```


#### func (*FunctionConnFilter) Read

```go
func (c *FunctionConnFilter) Read(b []byte) (n int, err error)
```
Read reads data from the underlying connection and modifies the bytes according
to c.Filter

#### func (*FunctionConnFilter) Write

```go
func (c *FunctionConnFilter) Write(b []byte) (n int, err error)
```
Write modifies the bytes according to c.Filter and writes the result to the
underlying connection

#### type RegexConnFilter

```go
type RegexConnFilter struct {
	FunctionConnFilter
}
```


#### func (*RegexConnFilter) Read

```go
func (c *RegexConnFilter) Read(b []byte) (n int, err error)
```
Read reads data from the underlying connection and replaces all occurrences of
target regex with empty strings. The modified data is then copied back to the
provided buffer.

#### func (*RegexConnFilter) ReadFilter

```go
func (c *RegexConnFilter) ReadFilter(b []byte) ([]byte, error)
```
NewRegexConnFilter creates a new RegexConnFilter that replaces occurrences of
target regex with empty strings in the data read from the connection. target
regex is c.match

#### func (*RegexConnFilter) WriteFilter

```go
func (c *RegexConnFilter) WriteFilter(b []byte) ([]byte, error)
```
WriteFilter creates a new RegexConnFilter that replaces occurrences of target
regex with empty strings in the data read from the connection. target regex is
c.match
