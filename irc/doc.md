# ircinspector
--
    import "github.com/go-i2p/go-connfilter/irc"


## Usage

#### type Config

```go
type Config struct {
	OnMessage func(*Message) error
	OnNumeric func(int, *Message) error
	Logger    Logger
}
```

Config contains inspector configuration

#### type Filter

```go
type Filter struct {
	Command  string
	Channel  string
	Prefix   string
	Callback func(*Message) error
}
```

Filter defines criteria for message filtering

#### type Inspector

```go
type Inspector struct {
}
```

Inspector implements the net.Listener interface with IRC inspection

#### func  New

```go
func New(listener net.Listener, config Config) *Inspector
```
New creates a new IRC inspector wrapping an existing listener

#### func (*Inspector) Accept

```go
func (i *Inspector) Accept() (net.Conn, error)
```
Accept implements net.Listener Accept method

#### func (*Inspector) AddFilter

```go
func (i *Inspector) AddFilter(filter Filter)
```

#### func (*Inspector) Addr

```go
func (i *Inspector) Addr() net.Addr
```
Addr implements net.Listener Addr method

#### func (*Inspector) Close

```go
func (i *Inspector) Close() error
```
Close implements net.Listener Close method

#### type Logger

```go
type Logger interface {
	Debug(format string, args ...interface{})
	Error(format string, args ...interface{})
}
```

Logger interface for customizable logging

#### type Message

```go
type Message struct {
	Raw      string
	Prefix   string
	Command  string
	Params   []string
	Trailing string
}
```

Message represents a parsed IRC message

#### func (*Message) String

```go
func (m *Message) String() string
```
