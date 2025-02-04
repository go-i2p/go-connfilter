# httpinspector
--
    import "github.com/go-i2p/go-connfilter/http"

Package httpinspector provides HTTP traffic inspection and modification
capabilities by wrapping the standard net.Listener interface.

## Usage

```go
var (
	ErrInvalidModification = errors.New("invalid HTTP message modification")
	ErrMalformedHTTP       = errors.New("malformed HTTP message")
	ErrClosedInspector     = errors.New("inspector is closed")
)
```
Common errors returned by the inspector.

#### func  DefaultRequestCallback

```go
func DefaultRequestCallback(*http.Request) error
```
DefaultRequestCallback is a no-op request callback.

#### func  DefaultResponseCallback

```go
func DefaultResponseCallback(*http.Response) error
```
DefaultResponseCallback is a no-op response callback.

#### type Config

```go
type Config struct {
	OnRequest  RequestCallback  // Called for each request
	OnResponse ResponseCallback // Called for each response
}
```

Config contains configuration options for the HTTP inspector.

#### func  DefaultConfig

```go
func DefaultConfig() Config
```
DefaultConfig returns a Config with reasonable defaults.

#### type Inspector

```go
type Inspector struct {
}
```

Inspector wraps a net.Listener to provide HTTP traffic inspection.

#### func  New

```go
func New(listener net.Listener, config Config) *Inspector
```
New creates a new Inspector wrapping the provided listener.

#### func (*Inspector) Accept

```go
func (i *Inspector) Accept() (net.Conn, error)
```
Accept implements the net.Listener Accept method.

#### func (*Inspector) Addr

```go
func (i *Inspector) Addr() net.Addr
```
Addr implements the net.Listener Addr method.

#### func (*Inspector) Close

```go
func (i *Inspector) Close() error
```
Close implements the net.Listener Close method.

#### type RequestCallback

```go
type RequestCallback func(*http.Request) error
```

RequestCallback is called for each HTTP request intercepted.

#### type ResponseCallback

```go
type ResponseCallback func(*http.Response) error
```

ResponseCallback is called for each HTTP response intercepted.
