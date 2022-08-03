package xtcp

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
)

func TestTCPPack(t *testing.T) {
	body := "test tcp pack"
	p := &TCPProto{
		Ver:  [4]byte{'v', '1', '.', '0'},
		Body: []byte(body),
	}
	buf := new(bytes.Buffer)
	err := p.Pack(buf)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("ttttt:%s\n", buf.String())
}

func TestTCPUnpack(t *testing.T) {
	body := "test tcp pack"
	p := &TCPProto{
		Ver:  [4]byte{'v', '1', '.', '0'},
		Body: []byte(body),
	}
	buf := new(bytes.Buffer)
	err := p.Pack(buf)
	if err != nil {
		fmt.Println("pack err:", err)
		return
	}
	p = &TCPProto{
		Ver:  [4]byte{},
		Body: make([]byte, maxBodySize),
	}
	err = p.Unpack(buf)
	if err != nil {
		fmt.Println("unpack err:", err)
		return
	}
	fmt.Printf("unpack body result:%+v\n", string(p.Body))
}

func TestTCPUnpackMoreByFor(t *testing.T) {
	body := "test tcp pack"
	p := &TCPProto{
		Ver:  [4]byte{'v', '1', '.', '0'},
		Body: []byte(body),
	}
	buf := new(bytes.Buffer)
	// 执行四次封包，模拟粘包
	_ = p.Pack(buf)
	_ = p.Pack(buf)
	_ = p.Pack(buf)
	_ = p.Pack(buf)

	for i := 0; i < 4; i++ {
		p = &TCPProto{}
		err := p.Unpack(buf)
		if err != nil {
			fmt.Printf("tcp unpack more err:%+v\n", err)
			return
		}
		fmt.Printf("tcp unpack more result:%s\n", string(p.Body))
	}
}

func TestTCPUnpackMoreByScanner(t *testing.T) {
	body := "test tcp pack"
	p := &TCPProto{
		Ver:  [4]byte{'v', '1', '.', '0'},
		Body: []byte(body),
	}
	buf := new(bytes.Buffer)
	// 执行四次封包，模拟粘包
	_ = p.Pack(buf)
	_ = p.Pack(buf)
	_ = p.Pack(buf)
	_ = p.Pack(buf)

	scanner := bufio.NewScanner(buf)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF {
			return
		}
		packLen := len(data)
		// 解析逻辑 + 判断逻辑
		if packLen < rawHeaderSize {
			return
		}
		if data[6] != 'v' {
			return
		}
		var l uint32
		_ = binary.Read(bytes.NewReader(data[:4]), binary.BigEndian, &l)
		if l > uint32(packLen) {
			return
		}
		return int(l), data[:l], nil
	})
	for {
		for scanner.Scan() {
			p = &TCPProto{}
			err := p.Unpack(bytes.NewReader(scanner.Bytes()))
			if err != nil {
				fmt.Printf("tcp unpack more err:%+v\n", err)
				return
			}
			fmt.Printf("tcp unpack more result:%s\n", string(p.Body))
		}
	}
}
