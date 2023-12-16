package cache

import (
	"sync"
)

// 内存池默认只管理2**24=16M以内的
const maxSize = 25

// 字节分片内存池
var caches [maxSize]sync.Pool

// 顺序初始化cache，第一个size为2**0，最后一个size为2**24
func init() {
	for i := 0; i < maxSize; i++ {
		size := 1 << i
		caches[i].New = func() any {
			buf := make([]byte, 0, size)
			return &buf
		}
	}
}

// size取值范围2**0~2**24，如果>=2**24会走普通的make不归sync.Pool管理
func New(size int) []byte {
	index := calcIndex(size)
	if index >= maxSize {
		return make([]byte, size)
	}
	buf := caches[index].Get().(*[]byte)
	return (*buf)[:size]
}

func calcIndex(size int) int {
	if size == 0 {
		return 0
	}
	index := log2(size)
	if (size & (size - 1)) == 0 {
		return index
	}
	// 非2的幂向右移，保证大于指定的size
	return index + 1
}

// 计算size对应的内存池位置
func log2(size int) int {
	index := 0
	for size != 0 {
		size = size >> 1
		index++
	}
	return index - 1
}

func Delete(buf []byte) {
	size := cap(buf)
	// 不是cache.New()管理的buf
	if (size & (size - 1)) != 0 {
		return
	}
	index := calcIndex(size)
	// 不是cache.New()管理的buf
	if index >= maxSize {
		return
	}
	buf = buf[:0]
	caches[index].Put(&buf)
}
