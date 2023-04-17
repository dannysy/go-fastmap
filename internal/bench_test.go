package internal

import (
	"testing"

	"github.com/alphadose/haxmap"

	"github.com/dannysy/go-fastmap"
	"github.com/dannysy/go-fastmap/internal/mutexmap"
)

const (
	Alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Numerals = "0123456789"
)

func getInStrings(n int) (out []string) {
	out = make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = string(Alphabet[i%len(Alphabet)]) + string(Numerals[i%len(Numerals)]) + Alphabet
	}
	return out
}

func BenchmarkBuiltinMapSet(b *testing.B) {
	b.StopTimer()
	inStrings := getInStrings(b.N)
	m := map[string]interface{}{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m[inStrings[i]] = i
	}
	b.ReportAllocs()
}

func BenchmarkHaxMapSet(b *testing.B) {
	b.StopTimer()
	inStrings := getInStrings(b.N)
	m := haxmap.New[string, interface{}]()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.Set(inStrings[i], i)
	}
	b.ReportAllocs()
}

func BenchmarkMutexMapSet(b *testing.B) {
	b.StopTimer()
	inStrings := getInStrings(b.N)
	m := mutexmap.New()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.Set(inStrings[i], i)
	}
	b.ReportAllocs()
}

func BenchmarkFastStringMapSet(b *testing.B) {
	b.StopTimer()
	m := fastmap.New()
	inStrings := getInStrings(b.N)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.Set(inStrings[i], i)
	}
	b.ReportAllocs()
}

func BenchmarkBuiltinMapGet(b *testing.B) {
	b.StopTimer()
	inStrings := getInStrings(b.N)
	m := map[string]interface{}{}
	for i := 0; i < b.N; i++ {
		m[inStrings[i]] = i
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m[inStrings[i]]
	}
	b.ReportAllocs()
}

func BenchmarkHaxMapGet(b *testing.B) {
	b.StopTimer()
	inStrings := getInStrings(b.N)
	m := haxmap.New[string, interface{}]()
	for i := 0; i < b.N; i++ {
		m.Set(inStrings[i], i)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Get(inStrings[i])
	}
	b.ReportAllocs()
}

func BenchmarkMutexMapGet(b *testing.B) {
	b.StopTimer()
	inStrings := getInStrings(b.N)
	m := mutexmap.New()
	for i := 0; i < b.N; i++ {
		m.Set(inStrings[i], i)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Get(inStrings[i])
	}
	b.ReportAllocs()
}

func BenchmarkFastStringMapGet(b *testing.B) {
	b.StopTimer()
	m := fastmap.New()
	inStrings := getInStrings(b.N)
	for i := 0; i < b.N; i++ {
		m.Set(inStrings[i], i)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Get(inStrings[i])
	}
	b.ReportAllocs()
}
