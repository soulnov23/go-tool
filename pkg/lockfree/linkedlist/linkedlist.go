// Package linkedlist 使用链表实现
package linkedlist

import (
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
	}
}

// Enqueue 入队列尾
func (queue *Queue) Enqueue(value any) {
	p := &node{
		value: value,
		next:  nil,
	}
	var tail, tailNext *node
	for {
		// 执行cas前先把上一刻的tail和tail.next保存
		tail = load(&queue.tail)
		tailNext = load(&tail.next)

		// 如果tail已经被其它线程移动了，重新开始
		if tail != load(&queue.tail) {
			continue
		}

		// 如果tail.next不为nil，往下遍历到尾位置
		if tailNext != nil {
			cas(&queue.tail, tail, tailNext)
			continue
		}

		// 尝试把p连接到tail.next
		if cas(&tail.next, tailNext, p) {
			// 入列成功，尝试把tail移到next新位置，失败了没关系不需要判断返回值，下次EnQueue/DeQueue时会遍历
			cas(&queue.tail, tail, p)
			return
		}
		// 入列失败继续try
	}
}

// Dequeue 出队列头
func (queue *Queue) Dequeue() any {
	var head, tail, headNext *node
	for {
		// 执行cas前先把上一刻的head，tail和head.next保存
		head = load(&queue.head)
		tail = load(&queue.tail)
		headNext = load(&head.next)

		// 如果head已经被其它线程移动了，重新开始
		if head != load(&queue.head) {
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
			continue
		}

		// 因为引入了dummy节点，所以每次操作的都是head.next的值
		if cas(&queue.head, head, headNext) {
			// 取值和清空都放在cas成功之后：head从head移到headNext这次cas全局只会成功一次，
			// 清空headNext.value的只可能是这次cas的赢家，所以赢家同时也是它唯一的读者，
			// 竞争失败者压根不会碰这个字段，value用普通字段读写不存在data race。
			// 此刻其它线程继续DeQueue把queue.head推到更后面也没关系，headNext是栈上的局部指针，
			// 指向的始终是本次摘下的那个节点，只有cas之后重新从queue.head出发取值才会读到后面节点的值
			value := headNext.value
			// headNext变成新的dummy头会一直留在队列里，不清空的话最后一个出队的元素会被它一直强引用无法GC
			headNext.value = nil
			// 注意：这里不能清空head.next。Enqueue的cas(&tail.next, nil, p)依赖
			// "next一旦非nil就不再变回nil"这个单调性来确认tail快照仍是真队尾，
			// 一旦把已摘下节点的next改回nil，持有过期tail快照的Enqueue会cas成功，
			// 把节点挂到已脱链的节点上导致元素丢失。旧head自身已不可达，会被GC整体回收
			// atomic.StorePointer(&head.next, nil)
			return value
		}
		// 出列失败继续try
	}
}
