package buffer

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/soulnov23/go-tool/pkg/cache"
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

type Buffer struct {
	head      *node
	readNode  *node
	writeNode *node // tail node

	readLock  sync.Mutex
	writeLock sync.Mutex

	size *atomic.Uint64
}

func New() *Buffer {
	return &Buffer{
		head:      nil,
		readNode:  nil,
		writeNode: nil,
		size:      &atomic.Uint64{},
	}
}

func (buffer *Buffer) Size() uint64 {
	return buffer.size.Load()
}

func (buffer *Buffer) Peek(size int) ([]byte, error) {
	if size <= 0 {
		return nil, ErrInvalidParam
	}
	buffer.readLock.Lock()
	defer buffer.readLock.Unlock()
	if buffer.Size() < uint64(size) {
		return nil, ErrNotEnoughData
	}
	// 遍历前面已经读取过的节点
	for buffer.readNode.size() == 0 {
		buffer.readNode = buffer.readNode.next
	}
	// size只需要读取一个节点的buf就足够了
	node := buffer.readNode
	if node.size() >= size {
		return node.peek(size), nil
	}
	// size需要读取多个节点的buf
	buf := cache.New(size)
	// ack记录遍历一个节点后的累积值，最终得到ack==size
	ack := 0
	for ack < size && node != nil {
		// 遇到空节点跳下一个节点
		if node.size() == 0 {
			node = node.next
			continue
		}
		offset := node.size()
		if ack+offset > size {
			offset = size - ack
		}
		tempBuf := node.peek(offset)
		copy(buf[ack:ack+offset], tempBuf)
		ack += offset
		// Peek不会修改readOffset需要手动跳下一个节点
		node = node.next
	}
	return buf[:ack], nil
}

func (buffer *Buffer) Skip(size int) error {
	if size <= 0 {
		return ErrInvalidParam
	}
	buffer.readLock.Lock()
	defer buffer.readLock.Unlock()
	if buffer.Size() < uint64(size) {
		return ErrNotEnoughData
	}
	// 遍历前面已经读取过的节点
	for buffer.readNode.size() == 0 {
		buffer.readNode = buffer.readNode.next
	}
	// size只需要读取一个节点的buf就足够了
	node := buffer.readNode
	if node.size() >= size {
		buffer.size.Add(^uint64(size - 1))
		node.skip(size)
		return nil
	}
	// size需要读取多个节点的buf
	// ack记录遍历一个节点后的累积值，最终得到ack==size
	ack := 0
	for ack < size && node != nil {
		// 节点内容被读完了跳下一个节点
		if node.size() == 0 {
			node = node.next
			continue
		}
		offset := node.size()
		if ack+offset > size {
			offset = size - ack
		}
		node.skip(offset)
		ack += offset
	}
	buffer.readNode = node
	buffer.size.Add(^uint64(ack - 1))
	return nil
}

func (buffer *Buffer) Read(size int) ([]byte, error) {
	if size <= 0 {
		return nil, ErrInvalidParam
	}
	buffer.readLock.Lock()
	defer buffer.readLock.Unlock()
	if buffer.Size() < uint64(size) {
		return nil, ErrNotEnoughData
	}
	// 遍历前面已经读取过的节点
	for buffer.readNode.size() == 0 {
		buffer.readNode = buffer.readNode.next
	}
	// size只需要读取一个节点的buf就足够了
	node := buffer.readNode
	if node.size() >= size {
		buffer.size.Add(^uint64(size - 1))
		return node.read(size), nil
	}
	// size需要读取多个节点的buf
	buf := cache.New(size)
	// ack记录遍历一个节点后的累积值，最终得到ack==size
	ack := 0
	for ack < size && node != nil {
		// 节点内容被读完了跳下一个节点
		if node.size() == 0 {
			node = node.next
			continue
		}
		offset := node.size()
		if ack+offset > size {
			offset = size - ack
		}
		tempBuf := node.read(offset)
		copy(buf[ack:ack+offset], tempBuf)
		ack += offset
	}
	buffer.readNode = node
	buffer.size.Add(^uint64(ack - 1))
	return buf[:ack], nil
}

func (buffer *Buffer) GC() {
	buffer.readLock.Lock()
	defer buffer.readLock.Unlock()
	buffer.writeLock.Lock()
	defer buffer.writeLock.Unlock()

	// 如果缓冲区为空，释放所有节点
	if buffer.Size() == 0 {
		for node := buffer.head; node != nil; {
			next := node.next
			node.delete()
			node = next
		}
		buffer.head, buffer.readNode, buffer.writeNode = nil, nil, nil
		return
	}

	// 如果有未读的数据，尝试回收已读完的节点
	if buffer.head != buffer.readNode && buffer.head != nil {
		// 回收head到readNode之间已经读完的节点
		for node := buffer.head; node != nil && node != buffer.readNode; {
			if node.size() > 0 {
				// 跳过还有未读数据的节点
				break
			}
			next := node.next
			node.delete()
			node = next
			buffer.head = node
		}
	}
}

func (buffer *Buffer) Write(buf []byte) {
	size := len(buf)
	if size == 0 {
		return
	}
	buffer.readLock.Lock()
	defer buffer.readLock.Unlock()
	buffer.writeLock.Lock()
	defer buffer.writeLock.Unlock()
	node := new(size)
	node.block, node.writeOffset = buf[:size], size
	if buffer.writeNode == nil {
		buffer.head, buffer.readNode, buffer.writeNode = node, node, node
	} else {
		buffer.writeNode.next = node
		buffer.writeNode = node
	}
	buffer.size.Add(uint64(size))
}

func (buffer *Buffer) Delete() {
	buffer.readLock.Lock()
	defer buffer.readLock.Unlock()
	buffer.writeLock.Lock()
	defer buffer.writeLock.Unlock()
	buffer.size.Store(0)
	for node := buffer.head; node != nil; {
		next := node.next
		node.delete()
		node = next
	}
	buffer.head, buffer.readNode, buffer.writeNode = nil, nil, nil
}
