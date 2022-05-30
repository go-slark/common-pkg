package resolver

import (
	"context"
	"github.com/pkg/errors"
	"github.com/smallfish-root/common-pkg/xgrpc/registry"
	"google.golang.org/grpc/resolver"
	"strings"
	"time"
)

const name = "discovery"

type builder struct {
	discoverer registry.Discovery
	timeout    time.Duration
}

type Option func(*builder)

func NewBuilder(d registry.Discovery, opts ...Option) resolver.Builder {
	b := &builder{
		discoverer: d,
		timeout:    time.Second * 10,
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

func (b *builder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	var (
		err error
		w   registry.Watcher
	)
	done := make(chan struct{}, 1)
	// 父context
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		w, err = b.discoverer.Watch(ctx, strings.TrimPrefix(target.URL.Path, "/"))
		close(done)
	}()
	timer := time.NewTimer(b.timeout)
	select {
	case <-done:
	case <-timer.C:
		err = errors.New("discovery create watcher timeout")
	}
	if err != nil {
		cancel()
		return nil, err
	}
	r := &discoveryResolver{
		w:      w,
		cc:     cc,
		ctx:    ctx,
		cancel: cancel,
	}
	go r.watch()
	return r, nil
}

func (*builder) Scheme() string {
	return name
}
