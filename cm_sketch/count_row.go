package cm_sketch

type cmRow []byte

// 为什么要➗2？
// 通常认为一个技术器占4bit，那么如果一行有numCounters个技术器，就只需要numCounters/2个bytes
func newCmRow(numCounters uint64) cmRow {
	return make(cmRow, numCounters/2)
}

// [0001 0010],[0111 1111]
func (r cmRow) get(n uint64) byte {
	// 如果n是偶数，n&1=0, 否则n&1=1
	return byte(r[n/2]>>((n&1)*4)) & 0x0f
}

func (r cmRow) increase(n uint64) {
	// 定位到第i个byte
	i := n / 2
	// 如果是n是奇数，我们就要左边4位，否则要右边4位
	// [0001 0010],[0111 1111]
	// n=0,得到的是0010，n=1得到的是0001
	s := (n & 1) * 4
	v := (r[i] >> s) & 0x0f
	// 如果没有超出最大计数，则增加计数
	if v < 15 {
		r[i] += 1 << s // 在第0位+1还是第4位+1
	}
}

// 保鲜机制，定期将所有的计数/2
func (r cmRow) reset() {
	for i := range r {
		// 将每一个字节都右移一位，&0x77表示保留每4位的后3位，表示/2，
		// 后面4位的首位变成了前面4位的末位，不会有影响吗？不会，因为&0x77会把那一位丢弃
		// 比如 0100 0111 -> 0010 0011 - >0010 0011
		r[i] = (r[i] >> 1) & 0x77 // 0x77 = 0111,0111
	}
}

func (r cmRow) clear() {
	for i := range r {
		r[i] = 0
	}
}

func next2Power(x uint64) uint64 {
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
