package internal

import (
	"math/rand"
	"time"
)

const cmDepth = 4

type cmSketch struct {
	rows [cmDepth]cmRow
	seed [cmDepth]uint64
	mask uint64
}

//numCounter - 1 = next2Power(n) = 0111111(n个1）

//0000,0000|0000,0000|0000,0000
//0000,0000|0000,0000|0000,0000
//0000,0000|0000,0000|0000,0000
//0000,0000|0000,0000|0000,0000

func newCmSketch(n int64) *cmSketch {
	if n == 0 {
		panic("cmSketch: bad numCounters")
	}

	numCounters := next2Power(n)
	sketch := &cmSketch{mask: uint64(numCounters - 1)}
	// Initialize rows of counters and seeds.
	source := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < cmDepth; i++ {
		sketch.seed[i] = source.Uint64()
		sketch.rows[i] = newCmRow(numCounters)
	}
	return sketch
}

func (s *cmSketch) increment(hashed uint64) {
	for i := range s.rows {
		s.rows[i].increment((hashed ^ s.seed[i]) & s.mask)
	}
}

// 找到最小的计数值
func (s *cmSketch) estimate(hashed uint64) int64 {
	min := byte(255)
	for i := range s.rows {
		val := s.rows[i].get((hashed ^ s.seed[i]) & s.mask)
		if val < min {
			min = val
		}
	}
	return int64(min)
}

// 让所有计数器都减半，保鲜机制
func (s *cmSketch) reset() {
	for _, r := range s.rows {
		r.reset()
	}
}

// 清空所有计数器
func (s *cmSketch) clear() {
	for _, r := range s.rows {
		r.clear()
	}
}

//快速计算最接近x的二次幂的算法
//比如x=5，返回8
//x = 110，返回128

//2^n
//1000000 (n个0）
//01111111（n个1） + 1
// x = 1001010 = 1111111 + 1 =10000000 
func next2Power(x int64) int64 {
	x--
	x |= x >> 1 
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x |= x >> 32
	x++
	return x
 }