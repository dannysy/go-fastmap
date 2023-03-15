package fastmap

import (
	"hash/fnv"
	"testing"

	"github.com/zeebo/xxh3"
)

func BenchmarkSimpleHash(b *testing.B) {
	b.StopTimer()
	strs := getInStrings(b.N)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Hash(strs[i])
	}
	b.ReportAllocs()
}

func BenchmarkXxHash(b *testing.B) {
	b.StopTimer()
	strs := getInStrings(b.N)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		xxh3.HashStringSeed(strs[i], 0)
	}
	b.ReportAllocs()
}

func BenchmarkFnva1(b *testing.B) {
	b.StopTimer()
	strs := getInStrings(b.N)
	hashFn := fnv.New128a()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hashFn.Write([]byte(strs[i]))
		hashFn.Reset()
	}
	b.ReportAllocs()
}
