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
	"github.com/coreos/etcd/clientv3"
)

const (
	// serviceNamespace the service prefix of ETCD key
	serviceNamespace = "ANTS-SRV@"
)

func newEtcdClient(endpoints []string) (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 15 * time.Second,
	})
}

// ServiceInfo serivce info
type ServiceInfo struct {
	Apis []string
	mu   sync.RWMutex
}

func (s *ServiceInfo) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, _ := json.Marshal(s)
	return goutil.BytesToString(b)
}
