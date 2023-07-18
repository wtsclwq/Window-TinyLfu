package util

import (
	"unsafe"
)

type HashKey[K comparable] struct {
	size  int
	isStr bool
}

func NewHash[K comparable]() *HashKey[K] {
	h := &HashKey[K]{}
	var k K
	switch (any(k)).(type) {
	case string:
		h.isStr = true
	default:
		h.size = int(unsafe.Sizeof(k))
	}
	return h
}

func (h *HashKey[K]) Hash(key K) string {
	var strKey string
	if h.isStr {
		strKey = *(*string)(unsafe.Pointer(&key))
	} else {
		strKey = *(*string)(unsafe.Pointer(&struct {
			data unsafe.Pointer
			len  int
		}{unsafe.Pointer(&key), h.size}))
	}
	return strKey
}
