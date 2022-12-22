package buffer

import (
	"sync"
	"sync/atomic"

	"github.com/SoulNov23/go-tool/pkg/cache"
)

var nodes sync.Pool

func init() {
	nodes.New = func() any {
		return &linkedBufferNode{
			referCount: 1,
		}
	}
}

type linkedBufferNode struct {
	block       []byte
	readOffset  int
	writeOffset int
	referCount  int32
	next        *linkedBufferNode
}

func NewNode(blockSize int) *linkedBufferNode {
	node := nodes.Get().(*linkedBufferNode)
	node.block = cache.New(blockSize)
	return node
}

func DeleteNode(node *linkedBufferNode) {
	if node == nil {
		return
	}
	if atomic.AddInt32(&node.referCount, -1) == 0 {
		cache.Delete(node.block)
		node.block, node.readOffset, node.writeOffset, node.next = nil, 0, 0, nil
		nodes.Put(node)
	}
}

func (node *linkedBufferNode) Len() int {
	return node.writeOffset - node.readOffset
}

func (node *linkedBufferNode) Peek(size int) []byte {
	return node.block[node.readOffset : node.readOffset+size]
}

func (node *linkedBufferNode) Skip(size int) {
	node.readOffset += size
}

func (node *linkedBufferNode) Next(size int) []byte {
	offset := node.readOffset
	node.readOffset += size
	return node.block[offset:node.readOffset]
}
