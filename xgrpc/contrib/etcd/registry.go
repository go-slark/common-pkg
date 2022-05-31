package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/smallfish-root/common-pkg/xgrpc/registry"
	"go.etcd.io/etcd/client/v3"
	"time"
)

var (
	_ registry.Registrar = &Registry{}
	_ registry.Discovery = &Registry{}
)

type Option func(*options)

type options struct {
	ctx       context.Context
	namespace string
	ttl       time.Duration
	retry     uint
}

func Context(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

func Namespace(ns string) Option {
	return func(o *options) {
		o.namespace = ns
	}
}

func RegisterTTL(ttl time.Duration) Option {
	return func(o *options) {
		o.ttl = ttl
	}
}

func Retry(num uint) Option {
	return func(o *options) {
		o.retry = num
	}
}

type Registry struct {
	opts   *options
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

func NewRegistry(client *clientv3.Client, opts ...Option) *Registry {
	opt := &options{
		ctx:       context.Background(),
		namespace: "grpc-etcd",
		ttl:       time.Second * 15,
		retry:     5,
	}
	for _, o := range opts {
		o(opt)
	}
	return &Registry{
		opts:   opt,
		client: client,
		kv:     clientv3.NewKV(client),
	}
}

func (r *Registry) Register(ctx context.Context, service *registry.ServiceInstance) error {
	key := fmt.Sprintf("%s/%s/%s", r.opts.namespace, service.Name, service.ID)
	value, err := json.Marshal(service)
	if err != nil {
		return err
	}
	if r.lease != nil {
		_ = r.lease.Close()
	}
	r.lease = clientv3.NewLease(r.client)
	leaseID, err := r.registerWithKV(ctx, key, string(value))
	if err != nil {
		return err
	}

	go r.keepAlive(r.opts.ctx, leaseID, key, string(value))
	return nil
}

func (r *Registry) Deregister(ctx context.Context, service *registry.ServiceInstance) error {
	defer func() {
		if r.lease != nil {
			_ = r.lease.Close()
		}
	}()
	key := fmt.Sprintf("%s/%s/%s", r.opts.namespace, service.Name, service.ID)
	_, err := r.client.Delete(ctx, key)
	return err
}

func (r *Registry) GetService(ctx context.Context, name string) ([]*registry.ServiceInstance, error) {
	key := fmt.Sprintf("%s/%s", r.opts.namespace, name)
	result, err := r.kv.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	items := make([]*registry.ServiceInstance, 0, len(result.Kvs))
	for _, kv := range result.Kvs {
		svc := &registry.ServiceInstance{}
		err = json.Unmarshal(kv.Value, svc)
		if err != nil {
			return nil, err
		}
		if svc.Name != name {
			continue
		}
		items = append(items, svc)
	}
	return items, nil
}

func (r *Registry) Watch(ctx context.Context, name string) (registry.Watcher, error) {
	key := fmt.Sprintf("%s/%s", r.opts.namespace, name)
	return newWatcher(ctx, key, name, r.client)
}

func (r *Registry) registerWithKV(ctx context.Context, key string, value string) (clientv3.LeaseID, error) {
	grant, err := r.lease.Grant(ctx, int64(r.opts.ttl.Seconds()))
	if err != nil {
		return 0, err
	}
	_, err = r.client.Put(ctx, key, value, clientv3.WithLease(grant.ID))
	if err != nil {
		return 0, err
	}
	return grant.ID, nil
}

func (r *Registry) keepAlive(ctx context.Context, leaseID clientv3.LeaseID, key string, value string) {
	curLeaseID := leaseID
	k, err := r.client.KeepAlive(ctx, leaseID)
	if err != nil {
		curLeaseID = 0
	}

	for {
		var retry uint
		if curLeaseID != 0 {
			goto label
		}

		for ; retry < r.opts.retry; retry++ {
			if ctx.Err() != nil {
				return
			}

			idChan := make(chan clientv3.LeaseID, 1)
			errChan := make(chan error, 1)
			cancelCtx, cancel := context.WithCancel(ctx)
			go func() {
				defer cancel()
				id, registerErr := r.registerWithKV(cancelCtx, key, value)
				if registerErr != nil {
					errChan <- registerErr
				} else {
					idChan <- id
				}
			}()

			timer := time.NewTimer(3 * time.Second)
			select {
			case <-timer.C:
				cancel()
				continue
			case <-errChan:
				continue
			case curLeaseID = <-idChan:
			}

			k, err = r.client.KeepAlive(ctx, curLeaseID)
			if err == nil {
				break
			}
			time.Sleep(time.Duration(1<<retry) * time.Second)
		}
		if _, ok := <-k; !ok {
			// retry failed
			return
		}
	label:
		select {
		case _, ok := <-k:
			if !ok {
				if ctx.Err() != nil {
					// channel closed due to context cancel
					return
				}
				// need to retry registration
				curLeaseID = 0
				continue
			}
		case <-r.opts.ctx.Done():
			return
		}
	}
}
