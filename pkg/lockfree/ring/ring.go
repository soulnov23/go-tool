// Package ring 实现了基于环形缓冲区的有界无锁(lock-free)多生产者多消费者队列
// 采用Vyukov的MPMC bounded queue方案：每个节点带一个序列号，操作时先校验序列号确认
// "是否轮到本票号"以及节点的空/满状态，确认通过后再CAS抢占全局票号。
// 因为校验在抢票之前完成，CAS成功的那一刻节点必定归本次操作独占，
// 所以整个流程不存在任何等待循环：任一goroutine被抢占都不会阻塞其他goroutine，满足lock-free的进展保证
package ring

import (
	"errors"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/cpu"
)

// 错误定义
var (
	// ErrQueueFull 队列已满错误
	ErrQueueFull = errors.New("queue is full")
	// ErrQueueEmpty 队列为空错误
	ErrQueueEmpty = errors.New("queue is empty")
)

const (
	// 缓存行大小
	cacheLinePadSize = unsafe.Sizeof(cpu.CacheLinePad{})
	// 最小容量
	minCapacity = 2
)

// node 表示队列中的节点
// seq同时编码了"轮次"和"空满"两件事，取值只有两种：
//
//	seq == pos           节点为空，正轮到入队票号pos写入
//	seq == pos+1         节点已写入，正轮到出队票号pos取走
//
// 入队完成后seq推进到pos+1，出队完成后seq推进到pos+capacity，
// 即下一个会落到本节点上的入队票号。节点每capacity个票号复用一次，
// seq始终等于"下一个有权操作本节点的票号"，据此可以区分轮次，
// 保证复用同一节点的不同票号之间严格串行
type node struct {
	/*----------------CacheLine----------------*/
	seq   atomic.Uint64               // 序列号，标识当前轮到哪个票号操作本节点
	value any                         // 节点存储的值
	_     [cacheLinePadSize - 24]byte // 填充至整缓存行，避免相邻节点伪共享
}

// Queue 为了获得高性能，使用缓存行填充在多线程环境下避免伪共享
// 伪共享(False Sharing)指多个CPU核心访问不同变量，但这些变量在同一缓存行，导致缓存失效
// 通过确保关键字段处于不同的缓存行，可以避免缓存一致性流量，提高并发性能
type Queue struct {
	/*----------------CacheLine----------------*/
	capacity uint64                      // 队列容量，必须是2的幂
	mask     uint64                      // 掩码，用于计算索引(等于capacity-1)
	_        [cacheLinePadSize - 16]byte // 填充至缓存行大小
	/*----------------CacheLine----------------*/
	head atomic.Uint64              // 出队票号发号器
	_    [cacheLinePadSize - 8]byte // 避免伪共享
	/*----------------CacheLine----------------*/
	tail atomic.Uint64              // 入队票号发号器
	_    [cacheLinePadSize - 8]byte // 避免伪共享
	/*----------------CacheLine----------------*/
	nodes []node // 节点数组
}

// New 创建指定容量的环形队列
// 容量会自动调整为大于等于输入值的最小2的幂
// 使用2的幂作为容量可以通过位运算快速计算索引(index & mask)
func New(capacity uint64) *Queue {
	// 确保容量至少为2
	if capacity < minCapacity {
		capacity = minCapacity
	}

	capacity = roundUpToPower2(capacity)
	queue := &Queue{
		capacity: capacity,
		mask:     capacity - 1,
		nodes:    make([]node, capacity),
	}
	for index := range queue.nodes {
		// 初始化时序列号设置为索引位置，索引位置就是第一个落到该节点上的入队票号
		queue.nodes[index].seq.Store(uint64(index))
	}
	return queue
}

// roundUpToPower2 将输入值调整为大于等于它的最小2的幂
func roundUpToPower2(v uint64) uint64 {
	if v == 0 {
		return minCapacity
	}
	// 非2的幂
	if v&(v-1) != 0 {
		// 依次将最高位1右边的第1位、第2~3位，第4~7位，第8~15位，第16~31位，第32~63位置为1
		v |= v >> 1
		v |= v >> 2
		v |= v >> 4
		v |= v >> 8
		v |= v >> 16
		v |= v >> 32
		// 进一位，将最右边所有的1都置为0，只保留最高位为1，就是2的幂
		v += 1
	}
	return v
}

// Enqueue 将元素添加到队列尾部
// 队列满时返回ErrQueueFull。该判断是保守的：当已发出的入队票号填满一整圈但数据还没写完时，
// 同样按已满处理，不会退化成等待
func (queue *Queue) Enqueue(value any) error {
	// tail是本次尝试使用的入队票号
	tail := queue.tail.Load()
	for {
		node := &queue.nodes[tail&queue.mask]
		// 转成有符号数比较，利用二进制补码天然处理uint64回绕
		diff := int64(node.seq.Load() - tail)
		if diff < 0 {
			// 节点还压着上一轮(票号tail-capacity)的值没被取走，队列已满
			return ErrQueueFull
		}
		// 节点为空且正轮到本票号，此时才去抢票
		// CAS成功说明票号tail只属于本次调用，而序列号推进到tail+1只可能由票号tail的持有者完成，
		// 因此这一刻节点已被独占，可以直接写入，无需任何等待
		if diff == 0 && queue.tail.CompareAndSwap(tail, tail+1) {
			node.value = value
			// 先写数据，后原子更新序列号
			// atomic更新seq提供release语义，保证value的写入对后续观察到seq变化的goroutine可见
			node.seq.Store(tail + 1)
			return nil
		}
		// diff>0说明其他生产者已经推进了tail，CAS失败说明票号被抢走，两种情况都要重新取号
		tail = queue.tail.Load()
	}
}

// Dequeue 从队列头部取出元素
// 队列空时返回ErrQueueEmpty。该判断是保守的：当入队票号已发出但数据还没写完时，
// 该元素尚不可见，同样按空处理，不会退化成等待
func (queue *Queue) Dequeue() (any, error) {
	// head是本次尝试使用的出队票号
	head := queue.head.Load()
	for {
		node := &queue.nodes[head&queue.mask]
		// 节点已写入的标志是seq推进到了head+1
		diff := int64(node.seq.Load() - (head + 1))
		if diff < 0 {
			// 票号head对应的入队还没写完(seq仍为head)，队列为空
			return nil, ErrQueueEmpty
		}
		// 节点已写入且正轮到本票号，此时才去抢票，同理CAS成功即独占该节点
		if diff == 0 && queue.head.CompareAndSwap(head, head+1) {
			value := node.value
			// 清除引用帮助GC
			node.value = nil
			// 先读数据并清零，后原子更新序列号
			// 序列号推进到head+capacity，即下一个落到本节点上的入队票号，
			// atomic更新seq提供release语义，保证value的清零对后续观察到seq变化的goroutine可见
			node.seq.Store(head + queue.capacity)
			return value, nil
		}
		// diff>0说明其他消费者已经推进了head，CAS失败说明票号被抢走，两种情况都要重新取号
		head = queue.head.Load()
	}
}

// Size 返回队列当前大小（近似值）
// 在并发环境下tail和head的读取不是原子的，结果可能不完全精确
func (queue *Queue) Size() uint64 {
	tail := queue.tail.Load()
	head := queue.head.Load()
	if tail >= head {
		return tail - head
	}
	return 0
}

// Capacity 返回队列最大容量
func (queue *Queue) Capacity() uint64 {
	return queue.capacity
}

// IsEmpty 检查队列是否为空（近似值）
func (queue *Queue) IsEmpty() bool {
	return queue.tail.Load() == queue.head.Load()
}

// IsFull 检查队列是否已满（近似值）
func (queue *Queue) IsFull() bool {
	return queue.tail.Load()-queue.head.Load() >= queue.capacity
}
