package internal

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alphadose/haxmap"

	"github.com/dannysy/go-fastmap"
	"github.com/dannysy/go-fastmap/internal/mutexmap"
)

const (
	epochs  int = 4096
	mapSize     = 8
)

func setupHaxMap() *haxmap.Map[string, interface{}] {
	m := haxmap.New[string, interface{}](mapSize)
	in := getInStrings(epochs)
	for i := 0; i < epochs; i++ {
		m.Set(in[i], i)
	}
	return m
}

func setupFastStringMap() *fastmap.FastMap {
	m := fastmap.New()
	in := getInStrings(epochs)
	for i := 0; i < epochs; i++ {
		m.Set(in[i], i)
	}
	return m
}

func setupMutexMap() *mutexmap.MutexMap {
	m := mutexmap.New()
	in := getInStrings(epochs)
	for i := 0; i < epochs; i++ {
		m.Set(in[i], i)
	}
	return m
}

func setupSyncMap() *sync.Map {
	m := sync.Map{}
	in := getInStrings(epochs)
	for i := 0; i < epochs; i++ {
		m.Store(in[i], i)
	}
	return &m
}

func BenchmarkHaxMapReadsOnly(b *testing.B) {
	m := setupHaxMap()
	in := getInStrings(epochs)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < epochs; i++ {
				_, _ = m.Get(in[i])
			}
		}
	})
	b.ReportAllocs()
}

func BenchmarkHaxMapReadsWithWrites(b *testing.B) {
	m := setupHaxMap()
	in := getInStrings(epochs)
	var writer uintptr
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// use 1 thread as writer
		if atomic.CompareAndSwapUintptr(&writer, 0, 1) {
			for pb.Next() {
				for i := 0; i < epochs; i++ {
					m.Set(in[i], i)
				}
			}
		} else {
			for pb.Next() {
				for i := 0; i < epochs; i++ {
					_, _ = m.Get(in[i])
				}
			}
		}
	})
	b.ReportAllocs()
}
func BenchmarkFastStringMapReadsOnly(b *testing.B) {
	m := setupFastStringMap()
	in := getInStrings(epochs)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < epochs; i++ {
				_, _ = m.Get(in[i])
			}
		}
	})
	b.ReportAllocs()
}

func BenchmarkFastStringMapReadsWithWrites(b *testing.B) {
	m := setupFastStringMap()
	in := getInStrings(epochs)
	var writer uintptr
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// use 1 thread as writer
		if atomic.CompareAndSwapUintptr(&writer, 0, 1) {
			for pb.Next() {
				for i := 0; i < epochs; i++ {
					m.Set(in[i], i)
				}
			}
		} else {
			for pb.Next() {
				for i := 0; i < epochs; i++ {
					_, _ = m.Get(in[i])
				}
			}
		}
	})
	b.ReportAllocs()
}

func BenchmarkMutexMapReadsOnly(b *testing.B) {
	m := setupMutexMap()
	in := getInStrings(int(epochs))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < epochs; i++ {
				_, _ = m.Get(in[i])
			}
		}
	})
	b.ReportAllocs()
}

func BenchmarkMutexMapReadsWithWrites(b *testing.B) {
	m := setupMutexMap()
	in := getInStrings(int(epochs))
	var writer uintptr
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// use 1 thread as writer
		if atomic.CompareAndSwapUintptr(&writer, 0, 1) {
			for pb.Next() {
				for i := 0; i < epochs; i++ {
					m.Set(in[i], i)
				}
			}
		} else {
			for pb.Next() {
				for i := 0; i < epochs; i++ {
					_, _ = m.Get(in[i])
				}
			}
		}
	})
	b.ReportAllocs()
}

func BenchmarkSyncMapReadsOnly(b *testing.B) {
	m := setupSyncMap()
	in := getInStrings(int(epochs))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < epochs; i++ {
				_, _ = m.Load(in[i])
			}
		}
	})
	b.ReportAllocs()
}

func BenchmarkSyncMapReadsWithWrites(b *testing.B) {
	m := setupSyncMap()
	in := getInStrings(int(epochs))
	var writer uintptr
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		// use 1 thread as writer
		if atomic.CompareAndSwapUintptr(&writer, 0, 1) {
			for pb.Next() {
				for i := 0; i < epochs; i++ {
					m.Store(in[i], i)
				}
			}
		} else {
			for pb.Next() {
				for i := 0; i < epochs; i++ {
					_, _ = m.Load(in[i])
				}
			}
		}
	})
	b.ReportAllocs()
}
