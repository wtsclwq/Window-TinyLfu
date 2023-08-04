package internal

// 一个BitMap的实现
type cmRow []byte //byte = uint8 = 0000，0000 = COUNTER 4BIT = 2 dkCounter

// 64 dkCounter
// 1 uint8 =  2counter
// 32 uint8 = 64 dkCounter
func newCmRow(numCounters int64) cmRow {
	return make(cmRow, numCounters/2)
}

func (r cmRow) get(n uint64) byte {
	return byte(r[n/2]>>((n&1)*4)) & 0x0f
}

func (r cmRow) increment(n uint64) {
	//定位到第i个Counter
	i := n / 2 //r[i]
	//右移距离，偶数为0，奇数为4
	s := (n & 1) * 4
	//取前4Bit还是后4Bit
	v := (r[i] >> s) & 0x0f //0000, 1111
	//没有超出最大计数时，计数+1
	if v < 15 {
		r[i] += 1 << s
	}
}

// cmRow 100,
// 保鲜
func (r cmRow) reset() {
	// 计数减半
	for i := range r {
		r[i] = (r[i] >> 1) & 0x77 //0111，0111
	}
}

func (r cmRow) clear() {
	// 清空计数
	for i := range r {
		r[i] = 0
	}
}
