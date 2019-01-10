# Go-SiriDB-Connector

[![Go Report Card](https://goreportcard.com/badge/github.com/SiriDB/go-siridb-connector)](https://goreportcard.com/report/github.com/SiriDB/go-siridb-connector)
[![GoDoc](https://godoc.org/github.com/SiriDB/go-siridb-connector?status.svg)](https://godoc.org/github.com/SiriDB/go-siridb-connector)

A SiriDB-Connector for the Go language

---------------------------------------
  * [Installation](#installation)
  * [Usage](#usage)
    * [Single connection](#single-connection)
    * [SiriDB client](#siridb-client)
    * [Logging](#logging)

---------------------------------------

## Installation
Simple install the package to your [$GOPATH](https://github.com/golang/go/wiki/GOPATH "GOPATH") with the [go tool](https://golang.org/cmd/go/ "go command") from shell:
```bash
$ go get github.com/SiriDB/go-siridb-connector
```
Make sure [Git is installed](https://git-scm.com/downloads) on your machine and in your system's `PATH`.

## Usage
_Go SiriDB Connector_ can be used to communicate with a single SiriDB server and a more advanced client is provided which can connect to multiple SiriDB servers so queries and inserts are balanced.

### Single connection
This is some example code for how to use the Go-SiriDB-Connector as a single connection.
```go
package main

import (
	"fmt"

	"github.com/SiriDB/go-siridb-connector"
)

func example(conn *siridb.Connection, ok chan bool) {
	// make sure the connection will be closed
	defer conn.Close()

	// create a new database with the following configuration
	options := make(map[string]interface{})
	options["dbname"] = "dbtest"
	options["time_precision"] = "s"
	options["buffer_size"] = 1024
	options["duration_num"] = "1w"
	options["duration_log"] = "1d"

	// using the default service account 'sa' and password 'siri'
	if res, err := conn.Manage("sa", "siri", siridb.AdminNewDatabase, options); err == nil {
		fmt.Printf("Manage result: %v\n", res)
	}

	// connect to database 'dbtest' using user 'iris' and password 'siri'
	// this is an example but usually you should do some error handling...
	if err := conn.Connect("iris", "siri", "dbtest"); err == nil {

		seriename := "serie-001"
		timestamp := 1471254705
		value := 1.5
		serie := make(map[string][][]interface{})
		serie[seriename] = [][]interface{}{{timestamp, value}}

		// insert data
		if res, err := conn.Insert(serie, 10); err == nil {
			fmt.Printf("Insert result: %s\n", res)
		}

		// perform a query
		if res, err := conn.Query("select * from *", 10); err == nil {
			fmt.Printf("Query result: %v\n", res)
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
### SiriDB client
And one example for using the client. A client can be used for connecting to multiple siridb servers. Queries and inserts will be send to a random siridb server. When a connection is lost, it will retry to setup the connection each 30 seconds.
```go
package main

import (
	"fmt"

	"github.com/SiriDB/go-siridb-connector"
)

func example(client *siridb.Client, ok chan bool) {
	// make sure the connection will be closed
	defer client.Close()

	client.Connect()

	// IsConnected() returns true if at least one server is connected.
	// Failed connections will retry creating a connection each 30 seconds.
	if client.IsConnected() {
		if res, err := client.Query("list series", 2); err == nil {
			fmt.Printf("Query result: %s\n", res)
		}
	} else {
		fmt.Println("not even a single server is connected...")
	}

	// send to the channel
	ok <- true
}

func main() {
	// create a new client
	client := siridb.NewClient(
		"iris",   // username
		"siri",   // password
		"dbtest", // database
		[][]interface{}{
			{"server1", 9000},
			{"server2", 9000},
		}, // siridb server(s)
		nil, // optional log channel
	)

	// create a channel
	ok := make(chan bool)

	// run the example
	go example(client, ok)

	// wait for the channel
	<-ok
}
```
### Logging
Both a `Connection` and a `Client` can send logging to the standard output *or* to a channel for custom log handling.

For example you can create you own log handler like this:
```go
func printLogs(logCh chan string) {
	for {
		msg := <-logCh
		fmt.Printf("Log: %s\n", msg)
	}
}
```
And set-up the channel like this:
```go
logCh := make(chan string)
go printLogs(logCh)
```

If you plan to use the log channel with a `Connection` you should use the `.LogCh` property. For example:
```go
conn := siridb.NewConnection(...) // create a new connection
conn.LogCh = logCh // set-up custom log channel
```
The `Client` simple accepts the channel as argument. For example:
```go
client := siridb.NewClient(
	"user",
	"password",
	"database",
	[][]interface{}{...}, // array of servers, see SiriDB client for more info
	logCh, // logCh is allowed to be nil for logging to the standard output
)
```
