package buffer

import (
	"sync"
	"sync/atomic"

	"github.com/soulnov23/go-tool/pkg/cache"
)

var nodes sync.Pool

func init() {
	nodes.New = func() any {
		return &LinkedBufferNode{
			referCount: 1,
		}
	}
}

type LinkedBufferNode struct {
	block       []byte
	readOffset  int
	writeOffset int
	referCount  int32
	next        *LinkedBufferNode
}

func NewNode(blockSize int) *LinkedBufferNode {
	node := nodes.Get().(*LinkedBufferNode)
	node.block = cache.New(blockSize)
	return node
}

func (node *LinkedBufferNode) Len() int {
	return node.writeOffset - node.readOffset
}

func (node *LinkedBufferNode) Peek(size int) []byte {
	return node.block[node.readOffset : node.readOffset+size]
}

func (node *LinkedBufferNode) Skip(size int) {
	node.readOffset += size
}

func (node *LinkedBufferNode) Next(size int) []byte {
	offset := node.readOffset
	node.readOffset += size
	return node.block[offset:node.readOffset]
}

func (node *LinkedBufferNode) Close() {
	if atomic.AddInt32(&node.referCount, -1) == 0 {
		cache.Delete(node.block)
		node.block, node.readOffset, node.writeOffset, node.next = nil, 0, 0, nil
		nodes.Put(node)
	}
}
