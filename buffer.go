package siridb

import (
	"net"
)

// HeaderSize if the size of a package header.
const HeaderSize = 8

// Buffer is used to read data from a connection.
type Buffer struct {
	buf    []byte
	len    uint32
	pkg    *Pkg
	conn   net.Conn
	DataCh chan *Pkg
	ErrCh  chan error
}

// NewBuffer retur a pointer to a new buffer.
func NewBuffer() *Buffer {
	return &Buffer{
		buf:    make([]byte, 0),
		len:    0,
		pkg:    nil,
		conn:   nil,
		DataCh: make(chan *Pkg),
		ErrCh:  make(chan error, 1),
	}
}

// Read listens on a connection for data.
func (buffer Buffer) Read() {
	for {
		// try to read the data
		buf := make([]byte, 8192)
		n, err := buffer.conn.Read(buf)

		if err != nil {
			// send an error if it's encountered
			buffer.ErrCh <- err
			return
		}

		buffer.len += uint32(n)
		buffer.buf = append(buffer.buf, buf[:n]...)

		for buffer.len >= HeaderSize {
			if buffer.pkg == nil {
				buffer.pkg, err = NewPkg(buffer.buf)
				if err != nil {
					buffer.ErrCh <- err
					return
				}
			}

			total := buffer.pkg.size + HeaderSize

			if buffer.len < total {
				break
			}

			buffer.pkg.Data(&buffer.buf, total)

			buffer.DataCh <- buffer.pkg

			buffer.buf = buffer.buf[total:]
			buffer.len -= total
			buffer.pkg = nil
		}
	}
}
