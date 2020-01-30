# go-lumber

[![Go Report
Card](https://goreportcard.com/badge/github.com/elastic/go-lumber)](https://goreportcard.com/report/github.com/elastic/go-lumber)

Lumberjack protocol client and server implementations for go.

## Server Build Instructions

```
mkdir -p "$HOME/go/src/github.com/elastic/go-lumber"
git clone https://github.com/elastic/go-lumber "$HOME/go/src/github.com/elastic/go-lumber"
cd "$HOME/go/src/github.com/elastic/go-lumber"
glide update
go build cmd/tst-lj/main.go
./main --version
```
