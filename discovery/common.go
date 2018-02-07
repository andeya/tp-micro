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
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/henrylee2cn/goutil"
)

const (
	// serviceNamespace the service prefix of ETCD key
	serviceNamespace = "ANT-SRV@"
)

func createServiceKey(addr string) string {
	return serviceNamespace + addr
}

func getAddr(serviceKey string) string {
	return strings.TrimPrefix(serviceKey, serviceNamespace)
}

// EtcdConfig ETCD client config
type EtcdConfig struct {
	Endpoints   []string      `yaml:"endpoints"    ini:"endpoints"    comment:"list of URLs"`
	DialTimeout time.Duration `yaml:"dial_timeout" ini:"dial_timeout" comment:"timeout for failing to establish a connection"`
	Username    string        `yaml:"username"     ini:"username"     comment:"user name for authentication"`
	Password    string        `yaml:"password"     ini:"password"     comment:"password for authentication"`
}

// NewEtcdClient creates ETCD client.
// Note:
// If cfg.DialTimeout<0, it means unlimit;
// If cfg.DialTimeout=0, use the default value(15s).
func NewEtcdClient(cfg EtcdConfig) (*clientv3.Client, error) {
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 15 * time.Second
	} else if cfg.DialTimeout < 0 {
		cfg.DialTimeout = 0
	}
	return clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: cfg.DialTimeout,
		Username:    cfg.Username,
		Password:    cfg.Password,
	})
}

// ServiceInfo serivce info
type ServiceInfo struct {
	UriPaths []string `json:"uri_paths"`
	mu       sync.RWMutex
}

// String returns the JSON string.
func (s *ServiceInfo) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, _ := json.Marshal(s)
	return goutil.BytesToString(b)
}

// Append appends uri path
func (s *ServiceInfo) Append(uriPath ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.UriPaths = append(s.UriPaths, uriPath...)
}
