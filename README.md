# go-lumber

[![Go Report
Card](https://goreportcard.com/badge/github.com/elastic/go-lumber)](https://goreportcard.com/report/github.com/elastic/go-lumber)

Lumberjack protocol client and server implementations for go.

## Example Server

There is an example server in [cmd/tst-lj](cmd/tst-lj/main.go). It will accept
connections and log when it receives batches of events.

```
# Install to $GOPATH/bin.
go install github.com/elastic/go-lumber/cmd/tst-lj@latest

# Start server.
tst-lj -bind=localhost:5044 -v2
2022/08/14 00:13:54 Server config: server.options{timeout:30000000000, keepalive:3000000000, decoder:(server.jsonDecoder)(0x100d88e80), tls:(*tls.Config)(nil), v1:false, v2:true, ch:(chan *lj.Batch)(nil)}
2022/08/14 00:13:54 tcp server up
```