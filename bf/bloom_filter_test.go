package bf

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// 假阳性
func TestBloomFilter_Check(t *testing.T) {
	filter := NewBloomFilter(1, 100)
	for i := 0; i < 1000; i++ {
		filter.Set(fmt.Sprintf("abc%dxxx", i))
	}
	println(filter.set.String())
	for i := 0; i < 1000; i++ {
		if !filter.Check(fmt.Sprintf("ooo%dxxx", i)) {
			t.Fatal()
		}
	}
}

// 假阳性
func TestBloomFilter(t *testing.T) {
	bf := NewBloomFilter(1, 1000) // 创建布隆过滤器

	// 添加元素
	for i := 0; i < 10000; i++ {
		value := generateRandomString(10) // 生成随机字符串
		bf.Set(value)
	}

	// 检查元素是否存在
	for i := 0; i < 10000; i++ {
		value := generateRandomString(20) // 生成随机字符串
		if !bf.Check(value) {
			t.Errorf("%s should exist in the Bloom filter", value)
		}
	}
}

func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
