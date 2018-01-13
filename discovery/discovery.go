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
	"errors"
	"net"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	tp "github.com/henrylee2cn/teleport"
)

type Service struct {
	addr        string
	serviceKey  string
	excludeApis []string
	apis        []string
	apisMu      sync.RWMutex
	client      *clientv3.Client
	leaseid     clientv3.LeaseID
	stop        chan error
}

const (
	dialTimeout = 15 * time.Second
	// Namespace the service prefix of ETCD key
	Namespace = "ANTS-SRV@"
)

func NewService(namespace string, addr string, endpoints []string, excludeApis ...string) *Service {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})

	if err != nil {
		tp.Fatalf("discovery: %v", err)
		return nil
	}

	return &Service{
		addr:        addr,
		serviceKey:  Namespace + addr,
		excludeApis: excludeApis,
		client:      cli,
		stop:        make(chan error),
	}
}

func (s *Service) Name() string {
	return s.apiKey
}

func (s *Service) PostReg(handler *tp.Handler) error {
	for _, a := range s.excludeApis {
		if a == api {
			return nil
		}
	}
	s.apisMu.Lock()
	s.apis = append(s.apis, handler.Name())
	s.apisMu.Unlock()
	return nil
}

func (s *Service) PostListen() error {
	ch, err := s.keepAlive()
	if err != nil {
		return err
	}

	for {
		select {
		case err := <-s.stop:
			s.revoke()
			return err
		case <-s.client.Ctx().Done():
			return errors.New("server closed")
		case ka, ok := <-ch:
			if !ok {
				log.Println("keep alive channel closed")
				s.revoke()
				return nil
			} else {
				log.Printf("Recv reply from service: %s, ttl:%d", s.Name, ka.TTL)
			}
		}
	}
}

func (s *Service) Stop() {
	s.stop <- nil
}

func (s *Service) keepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	info := &s.Info

	key := "services/" + s.Name
	value, _ := json.Marshal(info)

	// minimum lease TTL is 5-second
	resp, err := s.client.Grant(context.TODO(), 5)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	_, err = s.client.Put(context.TODO(), key, string(value), clientv3.WithLease(resp.ID))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	s.leaseid = resp.ID

	return s.client.KeepAlive(context.TODO(), resp.ID)
}

func (s *Service) revoke() error {
	_, err := s.client.Revoke(context.TODO(), s.leaseid)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("servide:%s stop\n", s.Name)
	return err
}
