// Package discovery is the service discovery module implemented by ETCD.
//
// Copyright 2017 HenryLee. All Rights Reserved.
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
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
)

// service automatically registered api info to etcd
type service struct {
	addr        string
	serviceKey  string
	excludeApis []string
	serviceInfo *ServiceInfo
	client      *clientv3.Client
	leaseid     clientv3.LeaseID
}

const (
	// minimum lease TTL is 5-second
	minLeaseTTL = 5
)

var (
	_ tp.PostRegPlugin    = new(service)
	_ tp.PostListenPlugin = new(service)
)

// ServicePlugin creates a teleport plugin which automatically registered api info to etcd.
func ServicePlugin(addr string, endpoints []string, excludeApis ...string) tp.Plugin {
	s := ServicePluginFromEtcd(addr, nil, excludeApis...)
	var err error
	s.(*service).client, err = newEtcdClient(endpoints)
	if err != nil {
		tp.Fatalf("%v: %v", err, s.Name())
		return s
	}
	return s
}

// ServicePluginFromEtcd creates a teleport plugin which automatically registered api info to etcd.
func ServicePluginFromEtcd(addr string, etcdClient *clientv3.Client, excludeApis ...string) tp.Plugin {
	return &service{
		addr:        addr,
		serviceKey:  serviceNamespace + addr,
		excludeApis: excludeApis,
		client:      etcdClient,
		serviceInfo: new(ServiceInfo),
	}
}

func (s *service) Name() string {
	return "ETCD(" + s.serviceKey + ")"
}

func (s *service) PostReg(handler *tp.Handler) error {
	api := handler.Name()
	for _, a := range s.excludeApis {
		if a == api {
			return nil
		}
	}
	s.apisMu.Lock()
	s.apis = append(s.apis, api)
	s.apisMu.Unlock()
	return nil
}

func (s *service) PostListen() error {
	ch, err := s.keepAlive()
	if err != nil {
		return err
	}
	go func() {
		name := s.Name()
		for {
			select {
			case <-s.client.Ctx().Done():
				tp.Warnf("%s: etcd server closed", name)
				s.revoke()
				tp.Warnf("%s: stop\n", name)
				return
			case ka, ok := <-ch:
				if !ok {
					tp.Debugf("%s: etcd keep alive channel closed, and restart it", name)
					s.revoke()
					ch = s.anywayKeepAlive()
				} else {
					tp.Tracef("%s: recv etcd ttl:%d", name, ka.TTL)
				}
			}
		}
	}()
	return nil
}

func (s *service) anywayKeepAlive() <-chan *clientv3.LeaseKeepAliveResponse {
	ch, err := s.keepAlive()
	for err != nil {
		time.Sleep(minLeaseTTL)
		ch, err = s.keepAlive()
	}
	return ch
}

func (s *service) keepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	resp, err := s.client.Grant(context.TODO(), minLeaseTTL)
	if err != nil {
		return nil, err
	}

	_, err = s.client.Put(
		context.TODO(),
		s.serviceKey,
		s.serviceInfo.String(),
		clientv3.WithLease(resp.ID),
	)
	if err != nil {
		return nil, err
	}

	s.leaseid = resp.ID

	return s.client.KeepAlive(context.TODO(), resp.ID)
}

func (s *service) revoke() {
	_, err := s.client.Revoke(context.TODO(), s.leaseid)
	if err != nil {
		tp.Errorf("%s: revoke service error: %s", s.Name(), err.Error())
		return
	}
}
