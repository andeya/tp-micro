// Copyright 2018 HenryLee. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package discovery

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/henrylee2cn/ants"
	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
)

type linker struct {
	client   *clientv3.Client
	nodes    goutil.Map
	uriPaths goutil.Map
	delCh    chan string
}

const (
	linkerName = "ETCD(ANT-LINKER)"
	// health health state
	health = 0
	// subHealth sub-health state
	subHealth = -1
)

// NewLinker creates a etct service linker.
func NewLinker(endpoints []string) ants.Linker {
	etcdClient, err := newEtcdClient(endpoints)
	if err != nil {
		tp.Fatalf("%s: %v", linkerName, err)
		return nil
	}
	return NewLinkerFromEtcd(etcdClient)
}

// NewLinkerFromEtcd creates a etct service linker.
func NewLinkerFromEtcd(etcdClient *clientv3.Client) ants.Linker {
	l := &linker{
		client:   etcdClient,
		nodes:    goutil.AtomicMap(),
		uriPaths: goutil.AtomicMap(),
		delCh:    make(chan string, 256),
	}
	if err := l.initNodes(); err != nil {
		tp.Fatalf("%s: %v", linkerName, err)
	}
	go l.watchNodes()
	return l
}

func (l *linker) addNode(key string, info *ServiceInfo) {
	addr := getAddr(key)
	node := &Node{
		Addr:  addr,
		Info:  info,
		State: health,
	}
	l.nodes.Store(addr, node)
	var (
		v          interface{}
		ok         bool
		uriPathMap goutil.Map
	)
	for _, uriPath := range info.UriPaths {
		if v, ok = l.uriPaths.Load(uriPath); !ok {
			uriPathMap = goutil.RwMap(1)
			uriPathMap.Store(addr, node)
			l.uriPaths.Store(uriPath, uriPathMap)
		} else {
			uriPathMap = v.(goutil.Map)
			uriPathMap.Store(addr, node)
		}
	}
}

func (l *linker) delNode(key string) {
	addr := getAddr(key)
	_node, ok := l.nodes.Load(addr)
	if !ok {
		return
	}
	l.nodes.Delete(addr)
	for _, uriPath := range _node.(*Node).Info.UriPaths {
		_uriPathMap, ok := l.uriPaths.Load(uriPath)
		if !ok {
			continue
		}
		uriPathMap := _uriPathMap.(goutil.Map)
		if _, ok := uriPathMap.Load(addr); ok {
			uriPathMap.Delete(addr)
			if uriPathMap.Len() == 0 {
				l.uriPaths.Delete(uriPath)
			}
		}
	}
	l.delCh <- addr
}

func (l *linker) initNodes() error {
	resp, err := l.client.Get(context.TODO(), serviceNamespace, clientv3.WithPrefix())
	if err != nil || len(resp.Kvs) == 0 {
		return err
	}
	for _, kv := range resp.Kvs {
		l.addNode(string(kv.Key), getServiceInfo(kv.Value))
		tp.Printf("%s: INIT %q : %q\n", linkerName, kv.Key, kv.Value)
	}
	return nil
}

func (l *linker) watchNodes() {
	rch := l.client.Watch(context.TODO(), serviceNamespace, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				l.addNode(string(ev.Kv.Key), getServiceInfo(ev.Kv.Value))
				tp.Printf("%s: %s %q : %q\n", linkerName, ev.Type, ev.Kv.Key, ev.Kv.Value)
			case clientv3.EventTypeDelete:
				l.delNode(string(ev.Kv.Key))
				tp.Printf("%s: %s %q\n", linkerName, ev.Type, ev.Kv.Key)
			}
		}
	}
}

func getServiceInfo(value []byte) *ServiceInfo {
	info := &ServiceInfo{}
	err := json.Unmarshal(value, info)
	if err != nil {
		tp.Errorf("%s", err.Error())
	}
	return info
}

// NotFoundService reply error: not found service
var NotFoundService = tp.NewRerror(tp.CodeDialFailed, "Dial Failed", "not found service")

// Select selects a service address by URI path.
func (l *linker) Select(uriPath string) (string, *tp.Rerror) {
	iface, exist := l.uriPaths.Load(uriPath)
	if !exist {
		return "", NotFoundService
	}
	nodes := iface.(goutil.Map)
	if nodes.Len() == 0 {
		return "", NotFoundService
	}
	var node *Node
	for i := 0; i < nodes.Len(); i++ {
		if _, iface, exist = nodes.Random(); exist {
			if node = iface.(*Node); node.getState() == health {
				return node.Addr, nil
			}
		}
	}
	if node == nil {
		return "", NotFoundService
	}
	return node.Addr, nil
}

// EventDel pushs service node offline notification.
func (l *linker) EventDel() <-chan string {
	return l.delCh
}

// Sick marks the address status is unhealthy.
func (l *linker) Sick(addr string) {
	_node, ok := l.nodes.Load(addr)
	if ok {
		_node.(*Node).setState(subHealth)
	}
}

// Close closes the linker.
func (l *linker) Close() {
	close(l.delCh)
	l.client.Close()
}

// Node a service node info.
type Node struct {
	Addr  string
	Info  *ServiceInfo
	State int8
	mu    sync.RWMutex
}

func (n *Node) getState() int8 {
	n.mu.RLock()
	state := n.State
	n.mu.RUnlock()
	return state
}

func (n *Node) setState(state int8) {
	n.mu.Lock()
	n.State = state
	n.mu.Unlock()
}
