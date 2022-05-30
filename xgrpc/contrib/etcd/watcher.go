package etcd

import (
	"context"
	"encoding/json"
	"github.com/smallfish-root/common-pkg/xgrpc/registry"
	"go.etcd.io/etcd/client/v3"
)

var _ registry.Watcher = &watcher{}

type watcher struct {
	key         string
	ctx         context.Context
	cancel      context.CancelFunc
	watchChan   clientv3.WatchChan
	watcher     clientv3.Watcher
	kv          clientv3.KV
	first       bool
	serviceName string
}

func newWatcher(ctx context.Context, key, name string, client *clientv3.Client) (*watcher, error) {
	w := &watcher{
		key:         key,
		watcher:     clientv3.NewWatcher(client),
		kv:          clientv3.NewKV(client),
		first:       true,
		serviceName: name,
	}
	// 子context
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.watchChan = w.watcher.Watch(w.ctx, key, clientv3.WithPrefix(), clientv3.WithRev(0))
	return w, w.watcher.RequestProgress(context.Background())
}

func (w *watcher) Next() ([]*registry.ServiceInstance, error) {
	if w.first {
		item, err := w.getInstance()
		w.first = false
		return item, err
	}

	select {
	// 子context
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	case <-w.watchChan:
		return w.getInstance()
	}
}

func (w *watcher) Stop() error {
	w.cancel()
	return w.watcher.Close()
}

func (w *watcher) getInstance() ([]*registry.ServiceInstance, error) {
	result, err := w.kv.Get(w.ctx, w.key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	items := make([]*registry.ServiceInstance, 0, len(result.Kvs))
	for _, kv := range result.Kvs {
		service := &registry.ServiceInstance{}
		err = json.Unmarshal(kv.Value, service)
		if err != nil {
			return nil, err
		}
		if service.Name != w.serviceName {
			continue
		}
		items = append(items, service)
	}
	return items, nil
}
