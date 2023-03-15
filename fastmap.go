package fastmap

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/zeebo/xxh3"
)

const (
	growMul            = 4
	defaultBucketCount = 16
	decrement          = ^uint64(0)
)

type SetOption func(m *FastMap)

func WithSize(size uint64) func(m *FastMap) {
	return func(m *FastMap) {
		m.bucketsCount.Store(size)
		m.buckets = make([]bucket, size)
	}
}

func WithLoadFactor(lf float32) func(m *FastMap) {
	return func(m *FastMap) {
		m.loadFactor = lf
	}
}

type FastMap struct {
	loadFactor        float32
	size              atomic.Uint64
	seed              uint64
	bucketsCount      atomic.Uint64
	buckets           []bucket
	bucketsLock       sync.Mutex
	isEvacuating      atomic.Bool
	evacuatedSize     atomic.Uint64
	evacuationBuckets []bucket
}

type bucket struct {
	itemsLock sync.RWMutex
	items     []entry
}

type entry struct {
	key   string
	value interface{}
}

func New(opts ...SetOption) *FastMap {
	fm := FastMap{
		loadFactor:  0.75,
		seed:        uint64(time.Now().UnixNano()),
		bucketsLock: sync.Mutex{},
	}
	fm.bucketsCount.Store(defaultBucketCount)
	fm.buckets = make([]bucket, defaultBucketCount)
	for _, setOption := range opts {
		setOption(&fm)
	}
	return &fm
}

func (m *FastMap) Len() uint64 {
	return m.size.Load() + m.evacuatedSize.Load()
}

func (m *FastMap) Set(key string, value interface{}) {
	keyHash := xxh3.HashStringSeed(key, m.seed)
	if m.isEvacuating.Load() {
		m.setEv(key, value, keyHash)
		m.evacuatePart()
		return
	}
	m.set(key, value, keyHash)
	if m.shouldEvacuate() {
		m.bucketsLock.Lock()
		if m.isEvacuating.Load() {
			return
		}
		m.evacuationBuckets = make([]bucket, m.bucketsCount.Load()*growMul)
		m.isEvacuating.Store(true)
		m.bucketsLock.Unlock()
	}
}

func (m *FastMap) Get(key string) (interface{}, bool) {
	keyHash := xxh3.HashStringSeed(key, m.seed)
	if m.isEvacuating.Load() {
		m.evacuatePart()
		return m.getEv(key, keyHash)
	}
	return m.get(key, keyHash)
}

func (m *FastMap) Delete(key string) {
	keyHash := xxh3.HashStringSeed(key, m.seed)
	if m.isEvacuating.Load() {
		m.delEv(key, keyHash)
		m.evacuatePart()
		return
	}
	m.del(key, keyHash)
}

func (m *FastMap) setEv(key string, value interface{}, keyHash uint64) {
	bucketNum := keyHash % (m.bucketsCount.Load() * growMul)
	isNew := m.evacuationBuckets[bucketNum].set(key, value)
	if isNew {
		m.evacuatedSize.Add(1)
	}
}

func (m *FastMap) set(key string, value interface{}, keyHash uint64) {
	bucketNum := keyHash % m.bucketsCount.Load()
	isNew := m.buckets[bucketNum].set(key, value)
	if isNew {
		m.size.Add(1)
	}
}

func (m *FastMap) delEv(key string, keyHash uint64) {
	bucketNum := keyHash % m.bucketsCount.Load() * growMul
	found := m.evacuationBuckets[bucketNum].del(key)
	if !found {
		m.del(key, keyHash)
		return
	}
	m.evacuatedSize.Add(decrement)
}

func (m *FastMap) del(key string, keyHash uint64) {
	bucketNum := keyHash % m.bucketsCount.Load()
	found := m.buckets[bucketNum].del(key)
	if !found {
		return
	}
	m.size.Add(decrement)
}

func (m *FastMap) getEv(key string, keyHash uint64) (value interface{}, found bool) {
	bucketNum := keyHash % m.bucketsCount.Load() * growMul
	value, found = m.evacuationBuckets[bucketNum].get(key)
	if !found {
		value, found = m.get(key, keyHash)
	}
	return value, found
}

func (m *FastMap) get(key string, keyHash uint64) (interface{}, bool) {
	bucketNum := keyHash % m.bucketsCount.Load()
	return m.buckets[bucketNum].get(key)
}

func (m *FastMap) shouldEvacuate() bool {
	if float32(m.size.Load())/float32(m.bucketsCount.Load()) >= m.loadFactor {
		return true
	}
	return false
}

func (m *FastMap) evacuatePart() {
	var found bool
	var e entry
	for i := range m.buckets {
		for j := range m.buckets[i].items {
			e = m.buckets[i].items[j]
			found = true
		}
	}
	if !found {
		m.bucketsLock.Lock()
		m.buckets = m.evacuationBuckets
		m.evacuationBuckets = nil
		m.bucketsLock.Unlock()
		m.bucketsCount.Store(m.bucketsCount.Load() * growMul)
		m.size.Swap(m.evacuatedSize.Load())
		m.evacuatedSize.Store(0)
		m.isEvacuating.Store(false)
		return
	}
	keyHash := xxh3.HashStringSeed(e.key, m.seed)
	m.setEv(e.key, e.value, keyHash)
	m.del(e.key, keyHash)
}

func (b *bucket) set(key string, value interface{}) (isNew bool) {
	b.itemsLock.RLock()
	for i := range b.items {
		e := &b.items[i]
		if e.key == key {
			e.value = value
			b.itemsLock.RUnlock()
			return false
		}
	}
	b.itemsLock.RUnlock()
	b.itemsLock.Lock()
	b.items = append(b.items, entry{
		key:   key,
		value: value,
	})
	b.itemsLock.Unlock()
	return true
}

func (b *bucket) get(key string) (interface{}, bool) {
	b.itemsLock.RLock()
	for i := range b.items {
		e := &b.items[i]
		if e.key == key {
			b.itemsLock.RUnlock()
			return e.value, true
		}
	}
	b.itemsLock.RUnlock()
	return nil, false
}

func (b *bucket) del(key string) bool {
	var i int
	var found bool
	b.itemsLock.RLock()
	for i = range b.items {
		e := &b.items[i]
		if e.key == key {
			found = true
			break
		}
	}
	b.itemsLock.RUnlock()
	if !found {
		return false
	}
	b.itemsLock.Lock()
	b.items[i] = b.items[len(b.items)-1]
	b.items[len(b.items)-1] = entry{}
	b.items = b.items[:len(b.items)-1]
	b.itemsLock.Unlock()
	return true
}
