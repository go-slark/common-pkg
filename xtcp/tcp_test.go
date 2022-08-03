package xtcp

import (
	"bytes"
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
	fmt.Printf("unpack result:%+v\n", p)
}
