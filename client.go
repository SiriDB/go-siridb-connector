package siridb

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

	qpack "github.com/transceptor-technology/go-qpack"
)

const defaultPingInterval = 30

// Client can be used to communicate with a SiriDB cluster.
type Client struct {
	PingInterval time.Duration
	username     string
	password     string
	dbname       string
	hosts        []*Host
	selector     []*Host
	logCh        chan string
}

// NewClient returns a pointer to a new client object.
// Example hostlist:
// [][]interface{}{
//	 {"myhost1", 9000}, 		// hostname/ip and port are required
//   {"myhost2", 9000, 2},      // an optional integer value can be used as weight
//								// (default weight is 1)
//   {"myhost3", 9000, true},   // if true is added as third argument the host
//								// will be used only when other hosts are not available
// }
//
func NewClient(username, password, dbname string, hostlist [][]interface{}, logCh chan string) *Client {
	client := Client{
		username:     username,
		password:     password,
		dbname:       dbname,
		logCh:        logCh,
		PingInterval: defaultPingInterval,
	}

	for _, v := range hostlist {
		host := NewHost(v[0].(string), uint16(v[1].(int)), logCh)

		// read optional backup mode and weight
		if len(v) == 3 {
			t := reflect.TypeOf(v[2])
			switch t.Kind() {
			case reflect.Bool:
				host.isBackup = v[2].(bool)
			case reflect.Int:
				host.weight = v[2].(int)
			default:
				client.sendLog("unknown type: %s", t.Kind().String())
			}
		}

		// append host to hosts, x times depending on the weight
		client.hosts = append(client.hosts, host)

		// append host to selector, x times depending on the weight
		for i := 0; i < host.weight; i++ {
			client.selector = append(client.selector, host)
		}
	}
	return &client
}

// Connect to SiriDB.
// make sure to initialize random: rand.Seed(time.Now().Unix())
func (client Client) Connect() {
	ok := make(chan bool)
	go client.ping(ok)
	<-ok
}

// IsConnected return true if at least one connection is connected
func (client Client) IsConnected() bool {
	for _, host := range client.hosts {
		if host.conn.IsConnected() {
			return true
		}
	}
	return false
}

// IsAvailable return true if at least one connection is connected
func (client Client) IsAvailable() bool {
	for _, host := range client.hosts {
		if host.isAvailable {
			return true
		}
	}
	return false
}

// Query a SiriDB database.
func (client Client) Query(query string, timeout uint16) (interface{}, error) {
	firstTry := true
	for {
		host := client.pickHost(false)

		if host == nil && firstTry {
			firstTry = false
			host = client.pickHost(true)
		}

		if host == nil {
			return nil, fmt.Errorf("no available conections found")
		}

		res, err := host.conn.Query(query, timeout)
		if err == nil {
			return res, err
		}

		if serr, ok := err.(*Error); ok && serr.Type() == CprotoErrServer {
			client.sendLog(
				"got a server error on %s: %s",
				host.conn.ToString(),
				serr.Error())
			host.isAvailable = false
			continue
		}

		return res, err
	}
}

// Insert data into a SiriDB database.
func (client Client) Insert(data interface{}, timeout uint16) (interface{}, error) {
	var err error
	var dataBin []byte

	if data != nil {
		dataBin, err = qpack.Pack(data)
		if err != nil {
			return nil, err
		}
	}
	return client.InsertBin(dataBin, timeout)

}

// Insert binary data into a SiriDB database.
func (client Client) InsertBin(data []byte, timeout uint16) (interface{}, error) {
	firstTry := true
	for {
		host := client.pickHost(false)

		if host == nil && firstTry {
			firstTry = false
			host = client.pickHost(true)
		}

		if host == nil {
			return nil, fmt.Errorf("no available connections found")
		}

		res, err := host.conn.InsertBin(data, timeout)
		if err == nil {
			return res, err
		}

		if serr, ok := err.(*Error); ok && serr.Type() == CprotoErrServer {
			client.sendLog(
				"got a server error on %s: %s",
				host.conn.ToString(),
				serr.Error())
			host.isAvailable = false
			continue
		}

		return res, err
	}
}

// Close will close all open connections.
func (client Client) Close() {
	for _, host := range client.hosts {
		host.conn.Close()
	}
}

func (client Client) pickHost(tryUnavailable bool) *Host {
	var available []*Host
	var nonBackup []*Host

	for _, host := range client.selector {

		if host.isAvailable || (tryUnavailable && host.conn.IsConnected()) {
			available = append(available, host)
		}
	}

	for _, host := range available {
		if !host.isBackup {
			nonBackup = append(nonBackup, host)
		}
	}

	if len(nonBackup) > 0 {
		return nonBackup[rand.Intn(len(nonBackup))]
	}

	if len(available) > 0 {
		return available[rand.Intn(len(available))]
	}

	return nil
}

func (client *Client) sendLog(s string, a ...interface{}) {
	msg := fmt.Sprintf(s, a...)
	if client.logCh == nil {
		fmt.Println(msg)
	} else {
		client.logCh <- msg
	}
}

func (client Client) ping(ok chan bool) {
	firstLoop := true
	for {
		for _, host := range client.hosts {
			if host.conn.IsConnected() {
				_, err := host.conn.Send(CprotoReqPing, nil, 5)
				if err != nil {
					client.sendLog("ping failed: %s", err)
					host.isAvailable = false
				} else {
					client.sendLog("ping! (%s:%d)", host.conn.host, host.conn.port)
					host.isAvailable = true
				}
			} else {
				err := host.conn.Connect(client.username, client.password, client.dbname)
				if err != nil {
					client.sendLog(err.Error())
					host.conn.Close()
				} else {
					client.sendLog("authenticated to %s:%d", host.conn.host, host.conn.port)
					host.isAvailable = true
				}
			}
		}
		if firstLoop {
			firstLoop = false
			ok <- true
		}
		time.Sleep(client.PingInterval * time.Second)
	}
}
