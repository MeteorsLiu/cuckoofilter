package cuckoo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
)

// maxCuckooKickouts is the maximum number of times reinsert
// is attempted.
const maxCuckooKickouts = 500

// Filter is a probabilistic counter.
type Filter interface {
	// Lookup returns true if data is in the filter.
	Lookup(data []byte) bool
	// Insert data into the filter. Returns false if insertion failed. In the resulting state, the filter
	// * Might return false negatives
	// * Deletes are not guaranteed to work
	// To increase success rate of inserts, create a larger filter.
	Insert(data []byte) bool
	// Delete data from the filter. Returns true if the data was found and deleted.
	Delete(data []byte) bool
	// Count returns the number of items in the filter.
	Count() uint

	// LoadFactor returns the fraction slots that are occupied.
	LoadFactor() float64
	// Reset removes all items from the filter, setting count to 0.
	Reset()
	// Encode returns a byte slice representing a Cuckoofilter.
	Encode() []byte
}

type filter[T fingerprintsize] struct {
	buckets        []bucket[T]
	getFingerprint func(hash uint64) T
	count          uint
	// Bit mask set to len(buckets) - 1. As len(buckets) is always a power of 2,
	// applying this mask mimics the operation x % len(buckets).
	bucketIndexMask uint
}

func numBuckets(numElements uint) uint {
	numBuckets := getNextPow2(uint64(numElements / bucketSize))
	if float64(numElements)/float64(numBuckets*bucketSize) > 0.96 {
		numBuckets <<= 1
	}
	if numBuckets == 0 {
		numBuckets = 1
	}
	return numBuckets
}

// NewFilter returns a new cuckoofilter suitable for the given number of elements.
// When inserting more elements, insertion speed will drop significantly and insertions might fail altogether.
// A capacity of 1000000 is a normal default, which allocates
// about ~2MB on 64-bit machines.
func NewFilter(cfg Config) Filter {
	numBuckets := numBuckets(cfg.NumElements)
	switch cfg.Precision {
	case Low:
		buckets := make([]bucket[uint8], numBuckets)
		return &filter[uint8]{
			buckets:         buckets,
			count:           0,
			bucketIndexMask: uint(len(buckets) - 1),
			getFingerprint:  getFinterprintUint8,
		}
	case High:
		buckets := make([]bucket[uint32], numBuckets)
		return &filter[uint32]{
			buckets:         buckets,
			count:           0,
			bucketIndexMask: uint(len(buckets) - 1),
			getFingerprint:  getFinterprintUint32,
		}
	default:
		buckets := make([]bucket[uint16], numBuckets)
		return &filter[uint16]{
			buckets:         buckets,
			count:           0,
			bucketIndexMask: uint(len(buckets) - 1),
			getFingerprint:  getFinterprintUint16,
		}
	}
}

func (cf *filter[T]) Lookup(data []byte) bool {
	i1, fp := getIndexAndFingerprint(data, cf.bucketIndexMask, cf.getFingerprint)
	if b := cf.buckets[i1]; b.contains(fp) {
		return true
	}
	i2 := getAltIndex(fp, i1, cf.bucketIndexMask)
	b := cf.buckets[i2]
	return b.contains(fp)
}

func (cf *filter[T]) Reset() {
	for i := range cf.buckets {
		cf.buckets[i].reset()
	}
	cf.count = 0
}

func (cf *filter[T]) Insert(data []byte) bool {
	i, fp := getIndexAndFingerprint(data, cf.bucketIndexMask, cf.getFingerprint)
	if cf.insertIntoBucket(fp, i) {
		return true
	}

	// Apply cuckoo kickouts until a free space is found.
	for k := 0; k < maxCuckooKickouts; k++ {
		j := rand.Intn(bucketSize)
		// Swap fingerprint with bucket entry.
		cf.buckets[i][j], fp = fp, cf.buckets[i][j]

		// Move kicked out fingerprint to alternate location.
		i = getAltIndex(fp, i, cf.bucketIndexMask)
		if cf.insertIntoBucket(fp, i) {
			return true
		}
	}
	return false
}

func (cf *filter[T]) insertIntoBucket(fp T, i uint) bool {
	if cf.buckets[i].insert(fp) {
		cf.count++
		return true
	}
	return false
}


func (cf *filter[T]) Delete(data []byte) bool {
	i1, fp := getIndexAndFingerprint(data, cf.bucketIndexMask, cf.getFingerprint)
	i2 := getAltIndex(fp, i1, cf.bucketIndexMask)
	return cf.delete(fp, i1) || cf.delete(fp, i2)
}

func (cf *filter[T]) delete(fp T, i uint) bool {
	if cf.buckets[i].delete(fp) {
		cf.count--
		return true
	}
	return false
}

func (cf *filter[T]) Count() uint {
	return cf.count
}

func (cf *filter[T]) LoadFactor() float64 {
	return float64(cf.count) / float64(len(cf.buckets)*bucketSize)
}

// TODO(panmari): Size of fingerprint needs to be derived from type. Currently hardcoded to 16 for uint16.
const bytesPerBucket = bucketSize * 16 / 8

func (cf *filter[T]) Encode() []byte {
	res := bytes.NewBuffer(nil)
	res.Grow(len(cf.buckets) * bytesPerBucket)
	for _, b := range cf.buckets {
		for _, fp := range b {
			binary.Write(res, binary.LittleEndian, fp)
		}
	}
	return res.Bytes()
}

// Decode returns a Cuckoofilter from a byte slice created using Encode.
// TODO(panmari): This only works for uint16 at this point.
func Decode(bytes []byte) (Filter, error) {
	if len(bytes)%bucketSize != 0 {
		return nil, fmt.Errorf("bytes must to be multiple of %d, got %d", bucketSize, len(bytes))
	}
	numBuckets := len(bytes) / bytesPerBucket
	if numBuckets < 1 {
		return nil, fmt.Errorf("bytes can not be smaller than %d, size in bytes is %d", bytesPerBucket, len(bytes))
	}
	if getNextPow2(uint64(numBuckets)) != uint(numBuckets) {
		return nil, fmt.Errorf("numBuckets must to be a power of 2, got %d", numBuckets)
	}
	var count uint
	buckets := make([]bucket[uint16], numBuckets)
	for i, b := range buckets {
		for j := range b {
			var next []byte
			next, bytes = bytes[:2], bytes[2:]

			if fp := binary.LittleEndian.Uint16(next); fp != 0 {
				buckets[i][j] = fp
				count++
			}
		}
	}
	return &filter[uint16]{
		buckets:         buckets,
		count:           count,
		bucketIndexMask: uint(len(buckets) - 1),
		getFingerprint:  getFinterprintUint16,
	}, nil
}
