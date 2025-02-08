Connection Filters
==================

These are connection-filtering middlewares that operate on the `net.Conn` and `net.Listener` level.
They can be used to build layered filtering systems for TCP connections.
There are several available filter-types, which can be categorized as "Coarse" or "Specific."
Coarse filters filter all traffic coming over the TCP connection, without checking anything about the traffic first.
Specific filters determine if the traffic is of the desired nature prior to performing filtering.

Coarse filters
--------------

- String-Pair Filters: These filters accept slices of strings.
 - Strings from slice A are replaced with strings from slice B, `A[0]->B[0]`, `A[1]->B[1]`, etc.
- Regex Filters: These filters match a regular expression, and replace it with an empty string.
- Function Filters: These filters are user-defined by setting:
 - `ReadFilter  func(b []byte) ([]byte, error)` for filter-on-read
 - `WriteFilter func(b []byte) ([]byte, error)` for filter-on-write

Specific Filters
----------------

- HTTP Filters: HTTP Filters are configured using callback functions like the Function Filters:
 - `type RequestCallback func(*http.Request) error`
 - `type ResponseCallback func(*http.Response) error`
- IRC Filters: IRC Filters are configured using a combination of callbacks and command filters:
 - `OnMessage func(*Message) error`
 - `OnNumeric func(int, *Message) error`

### LICENSE

MIT License