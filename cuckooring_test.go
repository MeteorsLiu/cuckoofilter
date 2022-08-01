package cuckoo

import (
	"crypto/rand"
	"io"
	"sync"
	"testing"
)

func BenchmarkCuckooRing(b *testing.B) {
	ring := NewCuckooRing(uint(2), uint(2e6))

	keys := benchmarkKeys(b, 1500000)
	for _, k := range keys {
		ring.Add(k)

	}
	k1 := make([]byte, 32)
	io.ReadFull(rand.Reader, k1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ring.Test(k1)
	}
}

func BenchmarkCuckooRing1(b *testing.B) {
	ring := NewCuckooRing1(uint(2e6))

	keys := benchmarkKeys(b, 1500000)
	for _, k := range keys {
		ring.Add(k)
	}
	k1 := make([]byte, 32)
	io.ReadFull(rand.Reader, k1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ring.Test(k1)
	}
}
func BenchmarkCuckooFilter(b *testing.B) {
	filter := NewFilter(2e6)
	keys := benchmarkKeys(b, 1500000)
	for _, k := range keys {
		filter.Insert(k)

	}
	var lock sync.RWMutex
	k1 := make([]byte, 32)
	io.ReadFull(rand.Reader, k1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		lock.RLock()
		filter.Lookup(k1)
		lock.RUnlock()
	}
}
