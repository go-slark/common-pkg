package xws

import (
	"bufio"
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
	verSize       = 2
	heartSize     = 4
	rawHeaderSize = packSize + headerSize + verSize
	maxPackSize   = maxBodySize + rawHeaderSize

	// offset
	packOffset   = 0
	headerOffset = packOffset + packSize
	verOffset    = headerOffset + headerSize
)

type TCPProto struct {
	Ver  uint16
	Body []byte
}

func (p *TCPProto) Pack(w io.Writer) error {
	//writer := bufio.NewWriter(w)
	//writer.Write()
	//packLen := rawHeaderSize + len(p.Body)
	//var buf []byte

	//var err error
	//// 先发送包头 packLen headerLen ver
	//// 再发送包体
	//err = binary.Write(w, binary.BigEndian, buf)
	//err = binary.Write(w, binary.BigEndian, &p.Body)
	//return err
	return nil
}

func (p *TCPProto) Unpack(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	bytes := scanner.Bytes()
	if len(bytes) < rawHeaderSize {
		return errors.New("invalid package")
	}
	buf := bytes[:rawHeaderSize]
	packLen := binary.BigEndian.Uint32(buf[packOffset:headerOffset])
	headerLen := binary.BigEndian.Uint16(buf[headerOffset:verOffset])
	p.Ver = binary.BigEndian.Uint16(buf[verOffset:])
	if packLen > maxPackSize {
		return errors.New("invalid package length")
	}
	if headerLen != rawHeaderSize {
		return errors.New("invalid header length")
	}
	bodyLen := packLen - uint32(headerLen)
	if bodyLen > 0 {
		p.Body = bytes[rawHeaderSize:]
	} else {
		p.Body = nil
	}
	return nil
}
