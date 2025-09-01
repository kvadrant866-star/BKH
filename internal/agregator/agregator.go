package agregator

import (
	"encoding/binary"
	"hash/fnv"
	"sync"
	"time"
)

type shardKey struct {
	bannerID int
	tsMinute time.Time
}

type shard struct {
	mu sync.Mutex
	m  map[shardKey]int64
}

type Agregator struct {
	shards []shard
}

type Delta struct {
	BannerID int
	TSMinute time.Time
	Count    int64
}

func NewAgregator(numShards int) *Agregator {
	shards := make([]shard, 0, numShards)
	for i := 0; i < numShards; i++ {
		shards = append(shards, shard{m: make(map[shardKey]int64)})
	}
	return &Agregator{
		shards: shards,
	}
}

func (a *Agregator) shardIndex(k shardKey) int {
	h := fnv.New64a()
	m := k.tsMinute.Unix() / 60
	var b [16]byte
	binary.LittleEndian.PutUint64(b[0:8], uint64(k.bannerID))
	binary.LittleEndian.PutUint64(b[8:16], uint64(m))
	_, _ = h.Write(b[:])
	return int(h.Sum64() % uint64(len(a.shards)))
}

func (a *Agregator) Increment(bannerID int, ts time.Time) {
	k := shardKey{bannerID: bannerID, tsMinute: ts}
	idx := a.shardIndex(k)
	s := &a.shards[idx]
	s.mu.Lock()
	s.m[k]++
	s.mu.Unlock()
}

func (a *Agregator) FlushDeltas() []Delta {
	var out []Delta
	for i := range a.shards {
		s := &a.shards[i]
		s.mu.Lock()
		old := s.m
		s.m = make(map[shardKey]int64)
		s.mu.Unlock()
		if len(old) == 0 {
			continue
		}
		if out == nil {
			out = make([]Delta, 0, len(old))
		}
		for k, v := range old {
			out = append(out, Delta{BannerID: k.bannerID, TSMinute: k.tsMinute, Count: v})
		}
	}
	return out
}
