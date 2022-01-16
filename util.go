package cuckoo

import (
	metro "github.com/dgryski/go-metro"
)

func getAltIndex[T fingerprintsize](fp T, i uint, bucketIndexMask uint) uint {
	// NOTE(panmari): hash was originally computed as uint(metro.Hash64(fp, 1337)).
	// Multiplying with a constant has a similar effect and is cheaper.
	// 0x5bd1e995 is the hash constant from MurmurHash2
	const murmurConstant = 0x5bd1e995
	hash := uint(fp) * murmurConstant
	return (i ^ hash) & bucketIndexMask
}

func getFinterprintUint8(hash uint64) uint8 {
	const fingerprintSizeBits = 8
	const maxFingerprint = (1 << fingerprintSizeBits) - 1
	// Use most significant bits for fingerprint.
	shifted := hash >> (64 - fingerprintSizeBits)
	// Valid fingerprints are in range [1, maxFingerprint], leaving 0 as the special empty state.
	fp := shifted%(maxFingerprint-1) + 1
	return uint8(fp)
}

func getFinterprintUint16(hash uint64) uint16 {
	const fingerprintSizeBits = 16
	const maxFingerprint = (1 << fingerprintSizeBits) - 1
	// Use most significant bits for fingerprint.
	shifted := hash >> (64 - fingerprintSizeBits)
	// Valid fingerprints are in range [1, maxFingerprint], leaving 0 as the special empty state.
	fp := shifted%(maxFingerprint-1) + 1
	return uint16(fp)
}

func getFinterprintUint32(hash uint64) uint32 {
	const fingerprintSizeBits = 32
	const maxFingerprint = (1 << fingerprintSizeBits) - 1
	// Use most significant bits for fingerprint.
	shifted := hash >> (64 - fingerprintSizeBits)
	// Valid fingerprints are in range [1, maxFingerprint], leaving 0 as the special empty state.
	fp := shifted%(maxFingerprint-1) + 1
	return uint32(fp)
}

// getIndexAndFingerprint returns the primary bucket index and fingerprint to be used
func getIndexAndFingerprint[T fingerprintsize](data []byte, bucketIndexMask uint, getFingerprint func(uint64) T) (uint, T) {
	hash := metro.Hash64(data, 1337)
	f := getFingerprint(hash)
	// Use least significant bits for deriving index.
	i1 := uint(hash) & bucketIndexMask
	return i1, f
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
