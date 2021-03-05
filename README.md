# go-lumber
[![Build Status](https://beats-ci.elastic.co/job/Library/job/go-lumber-mbp/job/master/badge/icon)](https://beats-ci.elastic.co/job/Library/job/go-lumber-mbp/job/master/)
[![Go Report
Card](https://goreportcard.com/badge/github.com/elastic/go-lumber)](https://goreportcard.com/report/github.com/elastic/go-lumber)
[![Contributors](https://img.shields.io/github/contributors/elastic/go-lumber.svg)](https://github.com/elastic/go-lumber/graphs/contributors)
[![GitHub release](https://img.shields.io/github/release/elastic/go-lumber.svg?label=changelog)](https://github.com/elastic/go-lumber/releases/latest)

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
