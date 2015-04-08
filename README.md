pool 
======

[![GoDoc](https://godoc.org/github.com/djherbis/pool?status.svg)](https://godoc.org/github.com/djherbis/pool)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg)](LICENSE.txt)
[![Build Status](https://travis-ci.org/djherbis/pool.svg?branch=master)](https://travis-ci.org/djherbis/pool) 

Usage
-----

```go
package main

import(
  "net"

  "github.com/djherbis/pool"
)

func main() {
  p := NewPool(func() (net.Conn, error) {
    return net.Dial("tcp", "localhost:10000")
  })
  
  if c, err := p.Get(); err == nil {
    c.Write([]byte("hello world\n"))
    c.Close()
  }
}
```

Installation
------------
```sh
go get github.com/djherbis/pool
```
