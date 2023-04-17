package fastmap

import (
	"fmt"
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
		m.bucketsCount = size
		m.buckets = make([]bucket, size)
	}
}

func WithLoadFactor(lf float32) func(m *FastMap) {
	return func(m *FastMap) {
		m.loadFactor = lf
	}
}

type FastMap struct {
	seed        uint64
	loadFactor  float32
	loadCounter atomic.Uint64
	//size                  atomic.Uint64
	bucketsCount uint64
	buckets      []bucket
	bucketsLock  sync.RWMutex
	isEvacuating atomic.Bool
	//evacuatedSize         atomic.Uint64
	evacuationBucketsCount uint64
	evacuationBuckets      []bucket
	evacuationBucketsLock  sync.RWMutex
}

type bucket struct {
	items skiplist
}

type entry struct {
	key   string
	value interface{}
}

func New(opts ...SetOption) *FastMap {
	fm := FastMap{
		loadFactor:            0.75,
		seed:                  uint64(time.Now().UnixNano()),
		bucketsLock:           sync.RWMutex{},
		evacuationBucketsLock: sync.RWMutex{},
	}
	fm.bucketsCount = defaultBucketCount
	fm.buckets = make([]bucket, defaultBucketCount)
	for _, setOption := range opts {
		setOption(&fm)
	}
	return &fm
}

func (m *FastMap) Len() (out uint64) {
	m.bucketsLock.RLock()
	for i := range m.buckets {
		out += m.buckets[i].len()
	}
	m.bucketsLock.RUnlock()
	if m.isEvacuating.Load() {
		m.evacuationBucketsLock.RLock()
		for i := range m.evacuationBuckets {
			out += m.evacuationBuckets[i].len()
		}
		m.evacuationBucketsLock.RUnlock()
	}
	return out
	//return m.size.Load() + m.evacuatedSize.Load()
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
		ov := m.isEvacuating.Swap(true)
		if !ov {
			m.evacuationBucketsLock.Lock()
			m.evacuationBucketsCount = m.bucketsCount * growMul
			m.evacuationBuckets = make([]bucket, m.evacuationBucketsCount)
			m.evacuationBucketsLock.Unlock()
		}
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
	bucketNum := keyHash % m.evacuationBucketsCount
	m.evacuationBucketsLock.RLock()
	b := &m.evacuationBuckets[bucketNum]
	m.evacuationBucketsLock.RUnlock()
	_ = b.set(key, value)
	//if isNew {
	//	m.evacuatedSize.Add(1)
	//}
}

func (m *FastMap) set(key string, value interface{}, keyHash uint64) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered bucketNum %v and buckets count %v and len %v\n", keyHash%m.bucketsCount, m.bucketsCount, len(m.buckets))
			fmt.Printf("IS EVAC %v\n", m.isEvacuating.Load())
		}
	}()
	m.bucketsLock.RLock()
	bucketNum := keyHash % m.bucketsCount
	b := &m.buckets[bucketNum]
	m.bucketsLock.RUnlock()
	isNew := b.set(key, value)
	if isNew {
		//m.size.Add(1)
		m.loadCounter.Add(1)
	}
}

func (m *FastMap) delEv(key string, keyHash uint64) {
	bucketNum := keyHash % m.evacuationBucketsCount
	m.evacuationBucketsLock.RLock()
	b := &m.evacuationBuckets[bucketNum]
	m.evacuationBucketsLock.RUnlock()
	found := b.del(key)
	if !found {
		m.del(key, keyHash)
		return
	}
	//m.evacuatedSize.Add(decrement)
}

func (m *FastMap) del(key string, keyHash uint64) {
	m.bucketsLock.RLock()
	bucketNum := keyHash % m.bucketsCount
	b := &m.buckets[bucketNum]
	m.bucketsLock.RUnlock()
	found := b.del(key)
	if !found {
		return
	}
	//m.size.Add(decrement)
}

func (m *FastMap) getEv(key string, keyHash uint64) (value interface{}, found bool) {
	bucketNum := keyHash % m.evacuationBucketsCount
	m.evacuationBucketsLock.RLock()
	b := &m.evacuationBuckets[bucketNum]
	m.evacuationBucketsLock.RUnlock()
	value, found = b.get(key)
	if !found {
		value, found = m.get(key, keyHash)
	}
	return value, found
}

func (m *FastMap) get(key string, keyHash uint64) (interface{}, bool) {
	m.bucketsLock.RLock()
	bucketNum := keyHash % m.bucketsCount
	b := &m.buckets[bucketNum]
	m.bucketsLock.RUnlock()
	return b.get(key)
}

func (m *FastMap) shouldEvacuate() bool {
	//if float32(m.size.Load())/float32(m.bucketsCount.Load()) >= m.loadFactor {
	//	return true
	//}
	if float32(m.loadCounter.Load())/float32(m.bucketsCount) >= m.loadFactor {
		return true
	}
	return false
}

func (m *FastMap) evacuatePart() {
	var found bool
	var e entry
	m.bucketsLock.RLock()
	for i := range m.buckets {
		e, found = m.buckets[i].items.Any()
		if found {
			break
		}
	}
	m.bucketsLock.RUnlock()
	if !found {
		m.bucketsLock.Lock()
		m.buckets = m.evacuationBuckets
		if ov := m.isEvacuating.Swap(false); ov {
			m.bucketsCount = m.bucketsCount * growMul
		}
		m.bucketsLock.Unlock()
		//m.size.Swap(m.evacuatedSize.Load())
		//m.evacuatedSize.Store(0)
		return
	}
	keyHash := xxh3.HashStringSeed(e.key, m.seed)
	m.setEv(e.key, e.value, keyHash)
	m.del(e.key, keyHash)
}

func (b *bucket) set(key string, value interface{}) (isNew bool) {
	in := entry{key: key, value: value}
	ok := b.items.CompareAndSwap(&in)
	if ok {
		return false
	}
	b.items.Add(&in)
	return true
}

func (b *bucket) get(key string) (interface{}, bool) {
	out, ok := b.items.Get(key)
	return out.value, ok
}

func (b *bucket) del(key string) bool {
	return b.items.Delete(key)
}

func (b *bucket) len() uint64 {
	return b.items.Len()
}
