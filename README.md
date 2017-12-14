# go-sixpack [![Travis-CI](https://travis-ci.org/RadekD/go-sixpack.svg)](https://travis-ci.org/RadekD/go-sixpack) [![GoDoc](https://godoc.org/github.com/RadekD/go-sixpack?status.svg)](https://godoc.org/github.com/RadekD/go-sixpack)

Go client library for SeatGeek's Sixpack AB testing framework.

## Usage

```go
import "github.com/RadekD/go-sixpack"

client, err := sixpack.NewClient("https://sixpack.test")
if err != nil {
	// handle error gracefully
}

userAlt, err := client.Participate("my-exp", sixpack.WithAlternatives("a", "b", "c"), sixpack.WithClientID("clienid"))
if err != nil {
	// handle error gracefully
}

err := client.Convert("my-exp", sixpack.WithRequest(w, r))
if err != nil {
	// handle error gracefully
}
```

## What is Sixpack?

[Sixpack](http://sixpack.seatgeek.com/) is a language-agnostic AB testing framework. It makes easy to run A/B tests across multiple web services written in different languages.
