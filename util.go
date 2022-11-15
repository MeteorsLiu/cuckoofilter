package cuckoo

import (
	"github.com/zeebo/wyhash"
	"github.com/zeebo/xxh3"
)

var (
	altHash = [256]uint{}
	masks   = [65]uint{}
)

func init() {
	for i := 0; i < 256; i++ {
		altHash[i] = (uint(xxh3.Hash([]byte{byte(i)})))
	}
	for i := uint(0); i <= 64; i++ {
		masks[i] = (1 << i) - 1
	}
}

var rng wyhash.SRNG

// randi returns either i1 or i2 randomly.
func randi(i1, i2 uint) uint {
	// it's faster than mod, but the result is almost same.
	if uint32(uint64(uint32(rng.Uint64()))*uint64(2)>>32) == 0 {
		return i1
	}
	return i2
}

func getAltIndex(fp fingerprint, i uint, bucketPow uint) uint {
	mask := masks[bucketPow]
	hash := altHash[fp] & mask
	return (i & mask) ^ hash
}

func getFingerprint(hash uint64) byte {
	// Use least significant bits for fingerprint.
	fp := byte(hash%255 + 1)
	return fp
}

// getIndicesAndFingerprint returns the 2 bucket indices and fingerprint to be used
func getIndexAndFingerprint(data []byte, bucketPow uint) (uint, fingerprint) {
	hash := xxh3.Hash(data)
	fp := getFingerprint(hash)
	// Use most significant bits for deriving index.
	i1 := uint(hash>>32) & masks[bucketPow]
	return i1, fingerprint(fp)
}

func getNextPow2(n uint64) uint {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	n++
	return uint(n)
}
