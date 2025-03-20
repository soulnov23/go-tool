// Package linkedlist 使用链表实现
package linkedlist

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

// 解决ABA问题有两个思路：
// 1. 不要重用队列中的元素，DeQueue出来的A不要直接EnQueue进队列，应该new一个新的元素A出来然后在EnQueue进队列中。当然new一个新的元素也不绝对安全，如果是A先被delete了，接着调用new来new一个新的元素有可能会返回A的地址，这样还是存在ABA的风险。一般对于无锁编程中的内存回收采用延迟回收的方式，在确保被回收内存没有被其他线程使用的情况下安全回收内存。
// 2. 允许内存重用，对指向的内存采用标签指针(Tagged Pointers)的方式，标签作为一个版本号，随着标签指针上的每一次cas运算而增加，并且只增不减。
// 对于go：即使有GC，在高并发场景下仍需注意ABA问题，通过合理的内存屏障和延迟回收策略可以降低风险

type node struct {
	value any
	next  unsafe.Pointer
}

func load(addr *unsafe.Pointer) *node {
	return (*node)(atomic.LoadPointer(addr))
}

func cas(addr *unsafe.Pointer, old, new *node) bool {
	return atomic.CompareAndSwapPointer(addr, unsafe.Pointer(old), unsafe.Pointer(new))
}

// Queue
type Queue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
	size *atomic.Uint64
}

// New 创建无锁队列
func New() *Queue {
	// 分配一个空节点dummy头指针head来解决队列中如果只有一个元素，head和tail都指向同一个节点的问题
	p := &node{
		value: nil,
		next:  nil,
	}
	return &Queue{
		head: unsafe.Pointer(p),
		tail: unsafe.Pointer(p),
		size: &atomic.Uint64{},
	}
}

// Enqueue 入队列尾
func (queue *Queue) Enqueue(value any) {
	p := &node{
		value: value,
		next:  nil,
	}
	var tail, tailNext *node
	backoff := uint(0) // 添加退避策略
	for {
		// 执行cas前先把上一刻的tail和tail.next保存
		tail = load(&queue.tail)
		// 添加内存屏障确保load操作按顺序执行
		// 确保之前的读操作（load(&queue.tail)）在之后的读操作（load(&tail.next)）之前完成
		// 防止CPU或编译器的指令重排导致的不一致性
		atomic.CompareAndSwapPointer(&tail.next, nil, nil) // 轻量级内存屏障
		tailNext = load(&tail.next)

		// 如果tail已经被其它线程移动了，重新开始
		if tail != load(&queue.tail) {
			backoff = 0 // 重置退避
			continue
		}

		// 如果tail.next不为nil，往下遍历到尾位置
		if tailNext != nil {
			cas(&queue.tail, tail, tailNext)
			backoff = 0 // 重置退避
			continue
		}

		// 尝试把p连接到tail.next
		if cas(&tail.next, tailNext, p) {
			// 入列成功，尝试把tail移到next新位置，失败了没关系不需要判断返回值，下次EnQueue/DeQueue时会遍历
			cas(&queue.tail, tail, p)
			queue.size.Add(1)
			return
		}

		// 入列失败实现指数退避策略
		if backoff < 5 {
			backoff++
		}
		for i := uint(0); i < (1 << backoff); i++ {
			runtime.Gosched()
		}
	}
}

// Dequeue 出队列头
func (queue *Queue) Dequeue() any {
	var head, tail, headNext *node
	backoff := uint(0) // 添加退避策略
	for {
		// 执行cas前先把上一刻的head，tail和head.next保存
		head = load(&queue.head)
		// 添加内存屏障确保load操作按顺序执行
		// 确保之前的读操作（load(&queue.head)）在之后的读操作（load(&queue.tail)和load(&head.next)）之前完成
		// 防止CPU或编译器的指令重排导致的不一致性
		atomic.CompareAndSwapPointer(&head.next, nil, nil) // 轻量级内存屏障
		tail = load(&queue.tail)
		headNext = load(&head.next)

		// 如果head已经被其它线程移动了，重新开始
		if head != load(&queue.head) {
			backoff = 0 // 重置退避
			continue
		}

		// 因为引入了dummy节点，当队列中只有一个元素时，head!=tail，所以当队列中没有元素时，head==tail，分两种情况
		// 1. head==tail且head.next==nil，队列为空，返回nil
		if head == tail && headNext == nil {
			return nil
		}

		// 2. 如果其它线程EnQueue做了一半导致head.next!=nil，但是tail还没有移到新位置
		if head == tail && headNext != nil {
			// 尝试把tail移到next新位置，失败了没关系不需要判断返回值，下次EnQueue/DeQueue时会遍历
			cas(&queue.tail, tail, headNext)
			backoff = 0 // 重置退避
			continue
		}

		// 因为引入了dummy节点，所以每次操作的都是head.next的值
		// 执行cas前先把head.next的值保存下来，避免cas刚执行完那一刻，其它线程也同时DeQueue把head移动了，那么cas后再取值可能就是head.next.next的值
		value := headNext.value
		if cas(&queue.head, head, headNext) {
			// 使用-1代替^uint64(0)提高可读性
			queue.size.Add(^uint64(0)) // 相当于-1

			// 释放旧head的引用，帮助GC回收
			atomic.StorePointer(&head.next, nil)
			head.value = nil

			return value
		}

		// 出列失败实现指数退避策略
		if backoff < 5 {
			backoff++
		}
		for i := uint(0); i < (1 << backoff); i++ {
			runtime.Gosched()
		}
	}
}

// Size 实际大小，链表没有容量限制，这里不需要Capacity方法
func (queue *Queue) Size() uint64 {
	return queue.size.Load()
}

// IsEmpty 检查队列是否为空
func (queue *Queue) IsEmpty() bool {
	head := load(&queue.head)
	tail := load(&queue.tail)
	return head == tail && load(&head.next) == nil
}
