package xtcp

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	PING = "PING"
	PONG = "PONG"
)

type connOption struct {
	tcpConn    *net.TCPConn
	keepAlive  bool
	sndBuffer  int
	recBuffer  int
	hbTime     int64
	hbInterval time.Duration
	in         chan *TCPProto
	out        chan *TCPProto
	isClosed   bool
	closing    chan struct{}
	sync.Mutex
}

func newTCPConn(opts ...Option) *connOption {
	conn := &connOption{
		keepAlive:  true,
		sndBuffer:  1024,
		recBuffer:  1024,
		hbTime:     time.Now().Unix(),
		hbInterval: 60,
		in:         make(chan *TCPProto, 1000),
		out:        make(chan *TCPProto, 1000),
		closing:    make(chan struct{}, 1),
	}
	for _, opt := range opts {
		opt(conn)
	}
	return conn
}

type Option func(opt *connOption)

func WithSndBuffer(buf int) Option {
	return func(opt *connOption) {
		opt.sndBuffer = buf
	}
}

func WithRecBuffer(buf int) Option {
	return func(opt *connOption) {
		opt.recBuffer = buf
	}
}

func WithKeepAlive(ka bool) Option {
	return func(opt *connOption) {
		opt.keepAlive = ka
	}
}

func WithHBInterval(hbInterval time.Duration) Option {
	return func(opt *connOption) {
		opt.hbInterval = hbInterval
	}
}

func WithIn(in int) Option {
	return func(opt *connOption) {
		opt.in = make(chan *TCPProto, in)
	}
}

func WithOut(out int) Option {
	return func(opt *connOption) {
		opt.out = make(chan *TCPProto, out)
	}
}

type TCPConn interface {
	Send(msg *TCPProto) error
	Receive() (*TCPProto, error)
}

func Open(addr string, num int, opts ...Option) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	lis, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	for i := 0; i < num; i++ {
		go accept(lis, opts...)
	}
	return nil
}

func accept(lis *net.TCPListener, opts ...Option) {
	for {
		conn, err := lis.AcceptTCP()
		if err != nil {
			return
		}
		c := newTCPConn(opts...)
		err = conn.SetKeepAlive(c.keepAlive)
		if err != nil {
			return
		}
		err = conn.SetReadBuffer(c.recBuffer)
		if err != nil {
			return
		}
		err = conn.SetWriteBuffer(c.sndBuffer)
		if err != nil {
			return
		}

		go c.serverTCP(conn)
	}
}

func (c *connOption) serverTCP(conn *net.TCPConn) {
	c.tcpConn = conn
	go c.read()
	go c.write()
	go c.handleHB()
}

func (c *connOption) write() {
	// pack
	tk := time.NewTicker(time.Duration(c.hbInterval) * 4 / 5)
	defer func() {
		tk.Stop()
		c.Close()
	}()

	for {
		select {
		case m := <-c.out:
			err := m.Pack(c.tcpConn)
			if err != nil {
				return
			}
		case <-c.closing:
			return
		case <-tk.C:
			p := &TCPProto{Ver: [4]byte{'v', '1', '.', '0'}}
			err := p.PackHB(c.tcpConn)
			if err != nil {
				return
			}
		}
	}
}

func (c *connOption) read() {
	// unpack
	sc := bufio.NewScanner(c.tcpConn)
	sc.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF {
			return
		}
		packLen := len(data)
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
		// parse bytes
		for sc.Scan() {
			pk := sc.Bytes()
			p := &TCPProto{}
			err := p.Unpack(bytes.NewReader(pk))
			if err != nil {
				break
			}

			atomic.StoreInt64(&c.hbTime, time.Now().Unix())
			if len(pk) == rawHeaderSize+heartSize && string(p.Body) == PONG {
				continue
			}
			select {
			case c.in <- p:
			case <-c.closing:
				return
			}
		}
	}
}

func (c *connOption) handleHB() {
	for {
		ts := atomic.LoadInt64(&c.hbTime)
		if time.Now().Unix()-ts > int64(c.hbInterval) {
			c.Close()
			break
		}
		time.Sleep(2 * time.Second)
	}
}

func (c *connOption) Close() {
	_ = c.tcpConn.Close()
	c.Lock()
	defer c.Unlock()
	if c.isClosed {
		return
	}
	close(c.closing)
	c.isClosed = true
}

// expose interface rewrite

func (c *connOption) Send(m *TCPProto) error {
	var err error
	select {
	case c.out <- m:
	case <-c.closing:
		err = errors.New("conn is closing")
	}
	return err
}

func (c *connOption) Receive() (*TCPProto, error) {
	select {
	case m := <-c.in:
		return m, nil
	case <-c.closing:
		return nil, errors.New("conn is closing")
	}
}
