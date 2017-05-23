package siridb

// Host is where a client saves a connection
type Host struct {
	conn        *Connection
	isBackup    bool
	isAvailable bool
	weight      int
}

// NewHost return a pointer to a new host.
func NewHost(host string, port uint16, logCh chan string) *Host {
	h := Host{
		conn:        NewConnection(host, port),
		isBackup:    false,
		isAvailable: false,
		weight:      1,
	}
	h.conn.OnClose = h.onClose
	h.conn.LogCh = logCh
	return &h
}

func (host *Host) onClose() {
	host.isAvailable = false
}
