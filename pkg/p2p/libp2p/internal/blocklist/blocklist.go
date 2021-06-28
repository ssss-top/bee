// Copyright 2020 The Swarm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package blocklist

import (
	"strings"
	"time"

	"github.com/ethersphere/bee/pkg/p2p"
	"github.com/ethersphere/bee/pkg/storage"
	"github.com/ethersphere/bee/pkg/swarm"
)

var keyPrefix = "blocklist-"

// timeNow is used to deterministically mock time.Now() in tests.
var timeNow = time.Now

type Blocklist struct {
	store storage.StateStorer
}

func NewBlocklist(store storage.StateStorer) *Blocklist {
	return &Blocklist{store: store}
}

type entry struct {
	Timestamp time.Time `json:"timestamp"`
	Duration  string    `json:"duration"` // Duration is string because the time.Duration does not implement MarshalJSON/UnmarshalJSON methods.
}

func (b *Blocklist) Exists(overlay swarm.Address) (bool, error) {
	key := generateKey(overlay)
	timestamp, duration, err := b.get(key)
	if err != nil {
		if err == storage.ErrNotFound {
			return false, nil
		}

		return false, err
	}

	// using timeNow.Sub() so it can be mocked in unit tests
	if timeNow().Sub(timestamp) > duration && duration != 0 {
		_ = b.store.Delete(key)
		return false, nil
	}

	return true, nil
}

// 添加address到blocklist，阻塞duration时间
func (b *Blocklist) Add(overlay swarm.Address, duration time.Duration) (err error) {
	// 产生key, 并查找是否已经存在
	key := generateKey(overlay)
	_, d, err := b.get(key)
	if err != nil {
		if err != storage.ErrNotFound {
			return err
		}
	}

	// if peer is already blacklisted, blacklist it for the maximum amount of time
	// 如果peer之前在blocklist中，使用两者间最大的阻塞时间
	if duration < d && duration != 0 || d == 0 {
		duration = d
	}

	// 存储
	return b.store.Put(key, &entry{
		Timestamp: timeNow(),
		Duration:  duration.String(),
	})
}

// Peers returns all currently blocklisted peers.
func (b *Blocklist) Peers() ([]p2p.Peer, error) {
	var peers []p2p.Peer
	if err := b.store.Iterate(keyPrefix, func(k, v []byte) (bool, error) {
		if !strings.HasPrefix(string(k), keyPrefix) {
			return true, nil
		}
		addr, err := unmarshalKey(string(k))
		if err != nil {
			return true, err
		}

		t, d, err := b.get(string(k))
		if err != nil {
			return true, err
		}

		if timeNow().Sub(t) > d && d != 0 {
			// skip to the next item
			return false, nil
		}

		p := p2p.Peer{Address: addr}
		peers = append(peers, p)
		return false, nil
	}); err != nil {
		return nil, err
	}

	return peers, nil
}

// 获取key还需被阻塞的时间戳和时间间隔
func (b *Blocklist) get(key string) (timestamp time.Time, duration time.Duration, err error) {
	var e entry
	if err := b.store.Get(key, &e); err != nil {
		return time.Time{}, -1, err
	}

	duration, err = time.ParseDuration(e.Duration)
	if err != nil {
		return time.Time{}, -1, err
	}

	return e.Timestamp, duration, nil
}

// 根据address生成带前缀的key
func generateKey(overlay swarm.Address) string {
	// blocklist- + 16进制字符串
	return keyPrefix + overlay.String()
}

func unmarshalKey(s string) (swarm.Address, error) {
	addr := strings.TrimPrefix(s, keyPrefix)
	return swarm.ParseHexAddress(addr)
}
