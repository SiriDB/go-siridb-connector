# Go-SiriDB-Connector

A SiriDB-Connector for the Go language

---------------------------------------
  * [Installation](#installation)
  * [Usage](#usage)
    * [Single connection](#single-connection)
  
---------------------------------------

## Installation
Simple install the package to your [$GOPATH](https://github.com/golang/go/wiki/GOPATH "GOPATH") with the [go tool](https://golang.org/cmd/go/ "go command") from shell:
```bash
$ go get github.com/transceptor-technology/go-siridb-connector
```
Make sure [Git is installed](https://git-scm.com/downloads) on your machine and in your system's `PATH`.

## Usage
_Go SiriDB Connector_ can be used to communicate with a single SiriDB server and a more advanced client is provided which can connect to multiple SiriDB servers so queries and inserts are balanced.

### Single connection
```go
package main

import (
	"fmt"

	"github.com/transceptor-technology/go-siridb-connector"
)

func example(conn *siridb.Connection, ok chan bool) {
	// make sure the connection will be closed
	defer conn.Close()

	// connect to database 'dbtest' using user 'iris' and password 'siri'
	// this is an example but usually you should do some error handling...
	if err := conn.Connect("iris", "siri", "dbtest"); err == nil {

		// perform a query
		if res, err := conn.Query("list series", 10); err == nil {
			fmt.Printf("Query result: %s\n", res)
		}
	}

	// send to the channel
	ok <- true
}

func main() {
	// create a new connection
	conn := siridb.NewConnection("localhost", 9000)

	// a connection will send output to stdout except when a log channel is used.
	// setup a log channel using:
	//  	conn.LogCh = myLogChannel

	// create a channel
	ok := make(chan bool)

	// run the example
	go example(conn, ok)

	// wait for the channel
	<-ok
}
```
