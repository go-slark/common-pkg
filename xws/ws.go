package xws

import (
	"errors"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type connOption struct {
	wsConn     *websocket.Conn
	in         chan *Msg
	out        chan *Msg
	closing    chan struct{}
	isClosed   bool
	rBuffer    int
	wBuffer    int
	hbInterval time.Duration
	hbTime     int64
	wTime      time.Duration
	hsTime     time.Duration
	rLimit     int64
	sync.Mutex // avoid close chan duplicated
}

func NewWSConn(opts ...Option) *connOption {
	c := &connOption{
		in:         make(chan *Msg, 1000),
		out:        make(chan *Msg, 1000),
		closing:    make(chan struct{}, 1),
		rBuffer:    1024,
		wBuffer:    1024,
		hbInterval: 60,
		hbTime:     time.Now().Unix(),
		wTime:      10,
		hsTime:     3,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type Option func(opt *connOption)

func WithIn(in int) Option {
	return func(opt *connOption) {
		opt.in = make(chan *Msg, in)
	}
}

func WithOut(out int) Option {
	return func(opt *connOption) {
		opt.out = make(chan *Msg, out)
	}
}

func WithHBInterval(hbInterval time.Duration) Option {
	return func(opt *connOption) {
		opt.hbInterval = hbInterval
	}
}

func WithReadBuffer(rb int) Option {
	return func(opt *connOption) {
		opt.rBuffer = rb
	}
}

func WithWriteBuffer(wb int) Option {
	return func(opt *connOption) {
		opt.wBuffer = wb
	}
}

func WithWriteTime(wt time.Duration) Option {
	return func(opt *connOption) {
		opt.wTime = wt
	}
}

func WithHandShakeTime(hst time.Duration) Option {
	return func(opt *connOption) {
		opt.hsTime = hst
	}
}

func WithReadLimit(rLimit int64) Option {
	return func(opt *connOption) {
		opt.rLimit = rLimit
	}
}

type Msg struct {
	Type    int
	Payload []byte
}

type WSConn interface {
	Close()
	Receive() (*Msg, error)
	Send(m *Msg) error
}

func (c *connOption) Init(w http.ResponseWriter, r *http.Request) error {
	ws, err := (&websocket.Upgrader{
		HandshakeTimeout: c.hsTime,
		ReadBufferSize:   c.rBuffer,
		WriteBufferSize:  c.wBuffer,
		CheckOrigin: func(r *http.Request) bool {
			// 校验规则
			if r.Method != http.MethodGet {
				return false
			}
			// 允许跨域
			return true
		},
		EnableCompression: false,
	}).Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	c.wsConn = ws
	go c.read()
	go c.write()
	go c.handleHB()
	return nil
}

func (c *connOption) read() {
	if c.rLimit > 0 {
		c.wsConn.SetReadLimit(c.rLimit)
	}
	_ = c.wsConn.SetReadDeadline(time.Now().Add(c.hbInterval))
	for {
		msgType, payload, err := c.wsConn.ReadMessage()
		if err != nil {
			c.Close()
			break
		}
		m := &Msg{
			Type:    msgType,
			Payload: payload,
		}
		select {
		case c.in <- m:
			atomic.StoreInt64(&c.hbTime, time.Now().Unix())
		case <-c.closing:
			return
		}
	}
}

func (c *connOption) write() {
	tk := time.NewTicker(time.Duration(c.hbInterval) * 4 / 5)
	defer func() {
		tk.Stop()
		c.Close()
	}()

	for {
		select {
		case m := <-c.out:
			_ = c.wsConn.SetWriteDeadline(time.Now().Add(c.wTime))
			err := c.wsConn.WriteMessage(m.Type, m.Payload)
			if err != nil {
				c.Close()
				return
			}
		case <-c.closing:
			return
		case <-tk.C:
			_ = c.wsConn.SetWriteDeadline(time.Now().Add(c.wTime))
			err := c.wsConn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
		}
	}
}

func (c *connOption) handleHB() {
	c.wsConn.SetPongHandler(func(appData string) error {
		_ = c.wsConn.SetReadDeadline(time.Now().Add(c.hbInterval))
		atomic.StoreInt64(&c.hbTime, time.Now().Unix())
		return nil
	})

	for {
		ts := atomic.LoadInt64(&c.hbTime)
		if time.Now().Unix()-ts > int64(c.hbInterval) {
			c.Close()
			break
		}
		time.Sleep(2 * time.Second)
	}
}

func (c *connOption) Receive() (*Msg, error) {
	select {
	case m := <-c.in:
		return m, nil
	case <-c.closing:
		return nil, errors.New("conn is closing")
	}
}

func (c *connOption) Send(m *Msg) error {
	var err error
	select {
	case c.out <- m:
	case <-c.closing:
		err = errors.New("conn is closing")
	}
	return err
}

func (c *connOption) Close() {
	_ = c.wsConn.Close()
	c.Lock()
	defer c.Unlock()
	if c.isClosed {
		return
	}
	close(c.closing)
	c.isClosed = true
}
