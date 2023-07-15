package bf

import (
	"hash"

	"github.com/bits-and-blooms/bitset"
	"github.com/spaolacci/murmur3"
)

type BloomFilter struct {
	set         *bitset.BitSet
	hashHelpers []hash.Hash32
}

func NewBloomFilter(hashNum uint32, bitSize uint32) *BloomFilter {
	bf := new(BloomFilter)
	bf.set = bitset.New(uint(bitSize))
	bf.hashHelpers = make([]hash.Hash32, 0, hashNum)
	var initSeed uint32 = 16
	for i := uint32(0); i < hashNum; i++ {
		bf.hashHelpers = append(bf.hashHelpers, murmur3.New32WithSeed(initSeed+i))
	}
	return bf
}

func (b *BloomFilter) Set(value string) {
	for _, helper := range b.hashHelpers {
		set(b.set, value, helper)
	}
}

func (b *BloomFilter) Check(value string) bool {
	for _, helper := range b.hashHelpers {
		if !check(b.set, value, helper) {
			return false
		}
	}
	return true
}

func (b *BloomFilter) Reset() {
	b.set.ClearAll()
}

func set(bits *bitset.BitSet, value string, helper hash.Hash32) {
	helper.Reset()
	if _, err := helper.Write([]byte(value)); err != nil {
		panic(err)
	}
	bits.Len()
	bits.Set(uint(helper.Sum32()) % bits.Len())
}

func check(bits *bitset.BitSet, value string, helper hash.Hash32) bool {
	helper.Reset()
	if _, err := helper.Write([]byte(value)); err != nil {
		panic(err)
	}
	return bits.Test(uint(helper.Sum32()) % bits.Len())
}
