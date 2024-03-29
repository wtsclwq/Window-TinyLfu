package internal

import (
	"math"
)

// bloomFilter is a small bloom-filter-based cache admission policy
type bloomFilter struct {
	m      uint32    // size of bit vector in bits
	k      uint32    // distinct hash functions needed
	filter bitvector // our filter bit vector
}

func newBloomFilter(capacity int, falsePositiveRate float64) *bloomFilter {
	bits := float64(capacity) * -math.Log(falsePositiveRate) / (math.Log(2.0) * math.Log(2.0)) // in bits
	m := nextPowerOfTwo(uint32(bits))

	if m < 1024 {
		m = 1024
	}

	k := uint32(0.7 * float64(m) / float64(capacity))
	if k < 2 {
		k = 2
	}

	return &bloomFilter{
		m:      m,
		filter: newBitVec(m),
		k:      k,
	}
}

// insert inserts the byte array b into the bloom filter.  Returns true if the value
// was already considered to be in the bloom filter.
func (d *bloomFilter) insert(h uint64) bool {
	h1, h2 := uint32(h), uint32(h>>32)
	var o uint = 1
	for i := uint32(0); i < d.k; i++ {
		o &= d.filter.getAndSet((h1 + (i * h2)) & (d.m - 1))
	}
	return o == 1
}

// Reset clears the bloom filter
func (d *bloomFilter) reset() {
	if d == nil {
		return
	}
	for i := range d.filter {
		d.filter[i] = 0
	}
}

// Internal routines for the bit vector
type bitvector []uint64

func newBitVec(size uint32) bitvector {
	return make([]uint64, uint(size+63)/64)
}

// set bit 'bit' in the bitvector d and return previous value
func (b bitvector) getAndSet(bit uint32) uint {
	shift := bit % 64
	idx := bit / 64
	bb := b[idx]
	m := uint64(1) << shift
	b[idx] |= m
	return uint((bb & m) >> shift)
}

// return the integer >= i which is a power of two
func nextPowerOfTwo(i uint32) uint32 {
	n := i - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}
