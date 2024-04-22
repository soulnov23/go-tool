package netpoll

import (
	"runtime"
	"unsafe"
)

type operatorCache struct {
	first *FDOperator
	cache []*FDOperator
	// freelist store the freeable operator
	// to reduce GC pressure, we only store operator index here
	freelist []int32
}

func newOperatorCache() *operatorCache {
	cache := &operatorCache{
		cache:    make([]*FDOperator, 0, 1024),
		freelist: make([]int32, 0, 1024),
	}
	runtime.KeepAlive(cache)
	return cache
}

func (c *operatorCache) alloc() *FDOperator {
	if c.first == nil {
		const opSize = unsafe.Sizeof(FDOperator{})
		n := 4 * 1024 / opSize
		if n == 0 {
			n = 1
		}
		index := int32(len(c.cache))
		for i := uintptr(0); i < n; i++ {
			pd := &FDOperator{index: index}
			c.cache = append(c.cache, pd)
			pd.next = c.first
			c.first = pd
			index++
		}
	}
	operator := c.first
	c.first = operator.next
	return operator
}

// freeable mark the operator that could be freed
// only poller could do the real free action
func (c *operatorCache) freeable(operator *FDOperator) {
	// reset all state
	operator.reset()
	c.freelist = append(c.freelist, operator.index)
}

func (c *operatorCache) free() {
	if len(c.freelist) == 0 {
		return
	}

	for _, index := range c.freelist {
		operator := c.cache[index]
		operator.next = c.first
		c.first = operator
	}
	c.freelist = c.freelist[:0]
}
