package siridb

import (
	"encoding/binary"
	"fmt"

	"github.com/transceptor-technology/go-qpack"
)

// Pkg contains of a header and data.
type Pkg struct {
	size uint32
	pid  uint16
	tp   uint8
	data []byte
}

// NewPkg returns a poiter to a new pkg.
func NewPkg(b []byte) (*Pkg, error) {
	tp := b[6]
	check := b[7]

	if check != '\xff'^tp {
		return nil, fmt.Errorf("invalid checkbit")
	}

	return &Pkg{
		size: binary.LittleEndian.Uint32(b),
		pid:  binary.LittleEndian.Uint16(b[4:]),
		tp:   tp,
		data: nil,
	}, nil
}

// Data sets package data
func (p *Pkg) Data(b *[]byte, size uint32) {
	p.data = (*b)[8:size]
}

// pack returns a byte array containing a header with serialized data.
func pack(pid uint16, tp uint8, v interface{}) ([]byte, error) {
	var err error
	var data []byte

	if v != nil {
		data, err = qpack.Pack(v)
		if err != nil {
			return nil, err
		}
	}

	return packBin(pid, tp, data)
}

// packbin
func packBin(pid uint16, tp uint8, data []byte) ([]byte, error) {

	// create pkg with final size.
	pkg := make([]byte, 8+len(data))

	// set package length.
	binary.LittleEndian.PutUint32(pkg[0:], uint32(len(data)))

	// set package pid.
	binary.LittleEndian.PutUint16(pkg[4:], pid)

	// set package type and check bit.
	pkg[6] = tp
	pkg[7] = '\xff' ^ tp

	// copy data
	copy(pkg[8:], data)

	return pkg, nil
}
