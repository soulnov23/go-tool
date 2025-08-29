package buffer

import (
	"sync"
	"sync/atomic"

	"github.com/soulnov23/go-tool/pkg/cache"
)

var nodes sync.Pool

func init() {
	nodes.New = func() any {
		return &node{}
	}
}

type node struct {
	block       []byte
	readOffset  int
	writeOffset int
	referCount  atomic.Int32
	next        *node
}

func new(blockSize int) *node {
	node := nodes.Get().(*node)
	node.block = cache.New(blockSize)
	node.referCount.Store(1)
	return node
}

func (node *node) size() int {
	return node.writeOffset - node.readOffset
}

func (node *node) peek(size int) []byte {
	return node.block[node.readOffset : node.readOffset+size]
}

func (node *node) skip(size int) {
	node.readOffset += size
}

func (node *node) read(size int) []byte {
	offset := node.readOffset
	node.readOffset += size
	return node.block[offset:node.readOffset]
}

func (node *node) delete() {
	if node.referCount.CompareAndSwap(1, 0) {
		cache.Delete(node.block)
		node.block, node.readOffset, node.writeOffset, node.next = nil, 0, 0, nil
		nodes.Put(node)
	}
}
