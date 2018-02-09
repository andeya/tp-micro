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
	"time"

	"github.com/coreos/etcd/clientv3"
)

// EtcdClient ETCD v3 client
type EtcdClient = clientv3.Client

// EtcdConfig ETCD client config
type EtcdConfig struct {
	Endpoints   []string      `yaml:"endpoints"    ini:"endpoints"    comment:"list of URLs"`
	DialTimeout time.Duration `yaml:"dial_timeout" ini:"dial_timeout" comment:"timeout for failing to establish a connection"`
	Username    string        `yaml:"username"     ini:"username"     comment:"user name for authentication"`
	Password    string        `yaml:"password"     ini:"password"     comment:"password for authentication"`
}

// NewEtcdClient creates ETCD client.
// Note:
// If etcdConfig.DialTimeout<0, it means unlimit;
// If etcdConfig.DialTimeout=0, use the default value(15s).
func NewEtcdClient(etcdConfig EtcdConfig) (*clientv3.Client, error) {
	if etcdConfig.DialTimeout == 0 {
		etcdConfig.DialTimeout = 15 * time.Second
	} else if etcdConfig.DialTimeout < 0 {
		etcdConfig.DialTimeout = 0
	}
	return clientv3.New(clientv3.Config{
		Endpoints:   etcdConfig.Endpoints,
		DialTimeout: etcdConfig.DialTimeout,
		Username:    etcdConfig.Username,
		Password:    etcdConfig.Password,
	})
}

// types that migrated from etcd 'github.com/coreos/etcd/clientv3'

// OpOption configures Operations like Get, Put, Delete.
type OpOption = clientv3.OpOption

// LeaseID etcd lease ID
type LeaseID = clientv3.LeaseID

// functions that migrated from etcd 'github.com/coreos/etcd/clientv3'

// WithLease attaches a lease ID to a key in 'Put' request.
//  func WithLease(leaseID clientv3.LeaseID) clientv3.OpOption
var WithLease = clientv3.WithLease

// WithLimit limits the number of results to return from 'Get' request.
// If WithLimit is given a 0 limit, it is treated as no limit.
//  func WithLimit(n int64) clientv3.OpOption
var WithLimit = clientv3.WithLimit

// WithRev specifies the store revision for 'Get' request.
// Or the start revision of 'Watch' request.
//  func WithRev(rev int64) clientv3.OpOption
var WithRev = clientv3.WithRev

// WithSort specifies the ordering in 'Get' request. It requires
// 'WithRange' and/or 'WithPrefix' to be specified too.
// 'target' specifies the target to sort by: key, version, revisions, value.
// 'order' can be either 'SortNone', 'SortAscend', 'SortDescend'.
//  func WithSort(target SortTarget, order SortOrder) clientv3.OpOption
var WithSort = clientv3.WithSort

// WithPrefix enables 'Get', 'Delete', or 'Watch' requests to operate
// on the keys with matching prefix. For example, 'Get(foo, WithPrefix())'
// can return 'foo1', 'foo2', and so on.
//  func WithPrefix() clientv3.OpOption
var WithPrefix = clientv3.WithPrefix

// WithRange specifies the range of 'Get', 'Delete', 'Watch' requests.
// For example, 'Get' requests with 'WithRange(end)' returns
// the keys in the range [key, end).
// endKey must be lexicographically greater than start key.
//  func WithRange(endKey string) clientv3.OpOption
var WithRange = clientv3.WithRange

// WithFromKey specifies the range of 'Get', 'Delete', 'Watch' requests
// to be equal or greater than the key in the argument.
//  func WithFromKey() clientv3.OpOption
var WithFromKey = clientv3.WithFromKey

// WithSerializable makes 'Get' request serializable. By default,
// it's linearizable. Serializable requests are better for lower latency
// requirement.
//  func WithSerializable() clientv3.OpOption
var WithSerializable = clientv3.WithSerializable

// WithKeysOnly makes the 'Get' request return only the keys and the corresponding
// values will be omitted.
//  func WithKeysOnly() clientv3.OpOption
var WithKeysOnly = clientv3.WithKeysOnly

// WithCountOnly makes the 'Get' request return only the count of keys.
//  func WithCountOnly() clientv3.OpOption
var WithCountOnly = clientv3.WithCountOnly

// WithMinModRev filters out keys for Get with modification revisions less than the given revision.
//  func WithMinModRev(rev int64) clientv3.OpOption
var WithMinModRev = clientv3.WithMinModRev

// WithMaxModRev filters out keys for Get with modification revisions greater than the given revision.
//  func WithMaxModRev(rev int64) clientv3.OpOption
var WithMaxModRev = clientv3.WithMaxModRev

// WithMinCreateRev filters out keys for Get with creation revisions less than the given revision.
//  func WithMinCreateRev(rev int64) clientv3.OpOption
var WithMinCreateRev = clientv3.WithMinCreateRev

// WithMaxCreateRev filters out keys for Get with creation revisions greater than the given revision.
//  func WithMaxCreateRev(rev int64) clientv3.OpOption
var WithMaxCreateRev = clientv3.WithMaxCreateRev

// WithFirstCreate gets the key with the oldest creation revision in the request range.
//  func WithFirstCreate() []clientv3.OpOption
var WithFirstCreate = clientv3.WithFirstCreate

// WithLastCreate gets the key with the latest creation revision in the request range.
//  func WithLastCreate() []clientv3.OpOption
var WithLastCreate = clientv3.WithLastCreate

// WithFirstKey gets the lexically first key in the request range.
//  func WithFirstKey() []clientv3.OpOption
var WithFirstKey = clientv3.WithFirstKey

// WithLastKey gets the lexically last key in the request range.
//  func WithLastKey() []clientv3.OpOption
var WithLastKey = clientv3.WithLastKey

// WithFirstRev gets the key with the oldest modification revision in the request range.
//  func WithFirstRev() []clientv3.OpOption
var WithFirstRev = clientv3.WithFirstRev

// WithLastRev gets the key with the latest modification revision in the request range.
//  func WithLastRev
var WithLastRev = clientv3.WithLastRev

// every 10 minutes when there is no incoming events.
// Progress updates have zero events in WatchResponse.
//  func WithProgressNotify() clientv3.OpOption
var WithProgressNotify = clientv3.WithProgressNotify

// WithCreatedNotify makes watch server sends the created event.
//  func WithCreatedNotify() clientv3.OpOption
var WithCreatedNotify = clientv3.WithCreatedNotify

// WithFilterPut discards PUT events from the watcher.
//  func WithFilterPut() clientv3.OpOption
var WithFilterPut = clientv3.WithFilterPut

// WithFilterDelete discards DELETE events from the watcher.
//  func WithFilterDelete() clientv3.OpOption
var WithFilterDelete = clientv3.WithFilterDelete

// WithPrevKV gets the previous key-value pair before the event happens.
// If the previous KV is already compacted, nothing will be returned.
//  func WithPrevKV() clientv3.OpOption
var WithPrevKV = clientv3.WithPrevKV

// This option can not be combined with non-empty values.
// Returns an error if the key does not exist.
//  func WithIgnoreValue() clientv3.OpOption
var WithIgnoreValue = clientv3.WithIgnoreValue

// This option can not be combined with WithLease.
// Returns an error if the key does not exist.
//  func WithIgnoreLease() clientv3.OpOption
var WithIgnoreLease = clientv3.WithIgnoreLease

// WithAttachedKeys makes TimeToLive list the keys attached to the given lease ID.
//  func WithAttachedKeys() LeaseOption
var WithAttachedKeys = clientv3.WithAttachedKeys

// WithCompactPhysical makes Compact wait until all compacted entries are
// removed from the etcd server's storage.
//  func WithCompactPhysical() CompactOption
var WithCompactPhysical = clientv3.WithCompactPhysical

// WithRequireLeader requires client requests to only succeed
// when the cluster has a leader.
//  func WithRequireLeader(ctx context.Context) context.Context
var WithRequireLeader = clientv3.WithRequireLeader
