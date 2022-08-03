package xtcp

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	// max body size
	maxBodySize = 1024

	// size
	packSize      = 4
	headerSize    = 2
	verSize       = 4
	heartSize     = 4
	rawHeaderSize = packSize + headerSize + verSize
	maxPackSize   = maxBodySize + rawHeaderSize

	// offset
	packOffset   = 0
	headerOffset = packOffset + packSize
	verOffset    = headerOffset + headerSize
)

// import: 实现并没有将header中的所有字段都定义在TCPProto中，而是将有的字段如包长/头长直接写在了TCP流中

type TCPProto struct {
	Ver  [verSize]byte // v1.0
	Body []byte
}

func (p *TCPProto) Pack(w io.Writer) error {
	var err error
	packLen := rawHeaderSize + len(p.Body)
	write := func(data interface{}) {
		if err != nil {
			return
		}
		err = binary.Write(w, binary.BigEndian, data)
	}
	write(uint32(packLen))
	write(uint16(rawHeaderSize))
	write(&p.Ver)
	write(&p.Body)
	return err
}

func (p *TCPProto) Unpack(r io.Reader) error {
	var (
		err       error
		packLen   uint32
		headerLen uint16
	)
	err = binary.Read(r, binary.BigEndian, &packLen)
	if packLen > maxPackSize {
		return errors.New("invalid package size")
	}
	err = binary.Read(r, binary.BigEndian, &headerLen)
	if headerLen != rawHeaderSize {
		return errors.New("invalid header size")
	}
	err = binary.Read(r, binary.BigEndian, &p.Ver)
	bodyLen := packLen - uint32(headerLen)
	if bodyLen > 0 {
		p.Body = make([]byte, bodyLen)
		err = binary.Read(r, binary.BigEndian, &p.Body)
	} else {
		p.Body = nil
	}
	return err
}

// heartbeat as body

func (p *TCPProto) PackHB(w io.Writer) error {
	var err error
	packLen := rawHeaderSize + heartSize
	write := func(data interface{}) {
		if err != nil {
			return
		}
		err = binary.Write(w, binary.BigEndian, data)
	}
	write(uint32(packLen))
	write(uint16(rawHeaderSize))
	write(&p.Ver)
	write([]byte(PING))
	return err
}
