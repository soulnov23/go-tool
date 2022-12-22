package buffer

import (
	"errors"
	"sync/atomic"

	"github.com/SoulNov23/go-tool/pkg/cache"
)

const (
	Block1k = 1 << 10
	Block2k = 1 << 11
	Block4k = 1 << 12
	Block8k = 1 << 13
)

var (
	ErrInvalidParam  = errors.New("param is invalid")
	ErrNotEnoughData = errors.New("data is not enough")
)

type LinkedBuffer struct {
	head      *LinkedBufferNode
	readNode  *LinkedBufferNode
	writeNode *LinkedBufferNode // tail node

	len int32
}

func NewBuffer() *LinkedBuffer {
	return &LinkedBuffer{
		head:      nil,
		readNode:  nil,
		writeNode: nil,
		len:       0,
	}
}

func DeleteBuffer(buffer *LinkedBuffer) {
	if buffer == nil {
		return
	}
	atomic.StoreInt32(&buffer.len, 0)
	for node := buffer.head; node != nil; {
		next := node.next
		DeleteNode(node)
		node = next
	}
	buffer.head, buffer.readNode, buffer.writeNode = nil, nil, nil
}

func (buffer *LinkedBuffer) Len() int {
	return int(atomic.LoadInt32(&buffer.len))
}

func (buffer *LinkedBuffer) Peek(size int) ([]byte, error) {
	if size <= 0 {
		return nil, ErrInvalidParam
	}
	if buffer.Len() < size {
		return nil, ErrNotEnoughData
	}
	// 遍历前面已经读取过的节点
	for buffer.readNode.Len() == 0 {
		buffer.readNode = buffer.readNode.next
	}
	// size只需要读取一个节点的buf就足够了
	node := buffer.readNode
	if node.Len() >= size {
		return node.Peek(size), nil
	}
	// size需要读取多个节点的buf
	buf := cache.New(size)
	// ack记录遍历一个节点后的累积值，最终得到ack==size
	for ack := 0; ack < size; {
		// 遇到空节点直接跳过
		if node.Len() == 0 {
			node = node.next
			continue
		}
		offset := node.Len()
		if ack+offset > size {
			offset = size - ack
		}
		tempBuf := node.Peek(offset)
		copy(buf[ack:ack+offset], tempBuf)
		ack += offset
		node = node.next
	}
	return buf[:size], nil
}

func (buffer *LinkedBuffer) Skip(size int) error {
	if size <= 0 {
		return ErrInvalidParam
	}
	if buffer.Len() < size {
		return ErrNotEnoughData
	}
	// 遍历前面已经读取过的节点
	for buffer.readNode.Len() == 0 {
		buffer.readNode = buffer.readNode.next
	}
	// size只需要读取一个节点的buf就足够了
	node := buffer.readNode
	if node.Len() >= size {
		atomic.AddInt32(&buffer.len, -int32(size))
		node.Skip(size)
		return nil
	}
	// size需要读取多个节点的buf
	// ack记录遍历一个节点后的累积值，最终得到ack==size
	for ack := 0; ack < size; {
		if node.Len() == 0 {
			node = node.next
			continue
		}
		offset := node.Len()
		if ack+offset > size {
			offset = size - ack
		}
		node.Skip(offset)
		ack += offset
		node = node.next
	}
	buffer.readNode = node
	atomic.AddInt32(&buffer.len, -int32(size))
	return nil
}

func (buffer *LinkedBuffer) Next(size int) ([]byte, error) {
	if size <= 0 {
		return nil, ErrInvalidParam
	}
	if buffer.Len() < size {
		return nil, ErrNotEnoughData
	}
	// 遍历前面已经读取过的节点
	for buffer.readNode.Len() == 0 {
		buffer.readNode = buffer.readNode.next
	}
	// size只需要读取一个节点的buf就足够了
	node := buffer.readNode
	if node.Len() >= size {
		atomic.AddInt32(&buffer.len, -int32(size))
		return node.Peek(size), nil
	}
	// size需要读取多个节点的buf
	buf := cache.New(size)
	// ack记录遍历一个节点后的累积值，最终得到ack==size
	for ack := 0; ack < size; {
		// 遇到空节点直接跳过
		if node.Len() == 0 {
			node = node.next
			continue
		}
		offset := node.Len()
		if ack+offset > size {
			offset = size - ack
		}
		tempBuf := node.Next(offset)
		copy(buf[ack:ack+offset], tempBuf)
		ack += offset
		node = node.next
	}
	buffer.readNode = node
	atomic.AddInt32(&buffer.len, -int32(size))
	return buf[:size], nil
}

func (buffer *LinkedBuffer) OptimizeMemory() {
	node := buffer.head
	for node != nil {
		if node.Len() > 0 {
			break
		}
		next := node.next
		DeleteNode(node)
		node = next
	}
	buffer.head = node
}

func (buffer *LinkedBuffer) Append(buf []byte) {
	size := cap(buf)
	if size == 0 {
		return
	}
	node := NewNode(size)
	if buffer.writeNode == nil {
		buffer.head, buffer.readNode, buffer.writeNode = node, node, node
	} else {
		buffer.writeNode.next = NewNode(size)
		buffer.writeNode = buffer.writeNode.next
	}
	buffer.writeNode.block, buffer.writeNode.writeOffset = buf[:size], size
	atomic.AddInt32(&buffer.len, int32(size))
}
