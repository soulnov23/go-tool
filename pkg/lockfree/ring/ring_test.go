package ring

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"github.com/soulnov23/go-tool/pkg/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAnySize(t *testing.T) {
	var value any
	// any相当于两个指针大小，64为操作系统，指针大小是8，这里就是16
	t.Log(unsafe.Sizeof(value))
}

func TestRoundUpToPower2(t *testing.T) {
	testCases := []struct {
		input    uint64
		expected uint64
	}{
		{0, 2},
		{1, 1},
		{2, 2},
		{3, 4},
		{4, 4},
		{7, 8},
		{15, 16},
		{21, 32},
		{33, 64},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("RoundUpToPower2(%d)", tc.input), func(t *testing.T) {
			result := roundUpToPower2(tc.input)
			assert.Equal(t, tc.expected, result, "roundUpToPower2(%d) should return %d, got %d", tc.input, tc.expected, result)
		})
	}
}

// 基本操作测试
func TestBasicOperations(t *testing.T) {
	queue := New(16)

	// 测试初始状态
	assert.Equal(t, uint64(0), queue.Size(), "初始队列应该为空")
	assert.True(t, queue.IsEmpty(), "初始队列应该为空")
	assert.False(t, queue.IsFull(), "初始队列不应该是满的")

	// 测试入队和出队
	for i := 0; i < 10; i++ {
		err := queue.Enqueue(i)
		assert.NoError(t, err, "入队操作不应该出错")
	}

	assert.Equal(t, uint64(10), queue.Size(), "队列大小应该是10")
	assert.False(t, queue.IsEmpty(), "队列不应该为空")

	// 测试出队
	for i := 0; i < 5; i++ {
		val, err := queue.Dequeue()
		assert.NoError(t, err, "出队操作不应该出错")
		assert.Equal(t, i, val, "出队的值应该是按顺序的")
	}

	assert.Equal(t, uint64(5), queue.Size(), "出队后，队列大小应该是5")

	// 继续入队直到满
	for i := 0; i < 11; i++ {
		err := queue.Enqueue(i + 100)
		if i < 11 {
			assert.NoError(t, err, "入队操作不应该出错")
		}
	}

	// 测试已满队列
	assert.True(t, queue.IsFull(), "队列应该已满")
	err := queue.Enqueue(999)
	assert.Error(t, err, "向已满队列入队应该返回错误")
	assert.Equal(t, ErrQueueFull, err, "应该返回队列已满错误")

	// 测试全部出队
	for i := 0; i < 16; i++ {
		_, err := queue.Dequeue()
		assert.NoError(t, err, "出队操作不应该出错")
	}

	// 测试空队列
	assert.True(t, queue.IsEmpty(), "队列应该为空")
	_, err = queue.Dequeue()
	assert.Error(t, err, "从空队列出队应该返回错误")
	assert.Equal(t, ErrQueueEmpty, err, "应该返回队列已空错误")
}

// 并发测试
func TestConcurrentOperations(t *testing.T) {
	queue := New(1024)
	const (
		numProducers    = 4
		numConsumers    = 4
		opsPerGoroutine = 10000
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 记录已入队和已出队的数量
	var enqueued, dequeued atomic.Uint64

	// 启动生产者
	var wgProducers sync.WaitGroup
	for i := 0; i < numProducers; i++ {
		wgProducers.Add(1)
		go func(id int) {
			defer wgProducers.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					value := fmt.Sprintf("p%d-%d", id, j)
					for {
						err := queue.Enqueue(value)
						if err == nil {
							enqueued.Add(1)
							break
						} else if err == ErrQueueFull {
							// 队列满，让出CPU
							runtime.Gosched()
						} else {
							t.Errorf("意外的入队错误: %v", err)
							return
						}
					}
				}
			}
		}(i)
	}

	// 给生产者一点时间先填充队列
	time.Sleep(100 * time.Millisecond)

	// 启动消费者
	var wgConsumers sync.WaitGroup
	for i := 0; i < numConsumers; i++ {
		wgConsumers.Add(1)
		go func(id int) {
			defer wgConsumers.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					value, err := queue.Dequeue()
					if err == nil {
						// 成功出队
						if value != nil {
							dequeued.Add(1)
						}
					} else if err == ErrQueueEmpty {
						// 队列空，检查是否所有生产者已完成
						if enqueued.Load() >= numProducers*opsPerGoroutine &&
							dequeued.Load() >= enqueued.Load() {
							return
						}
						// 让出CPU
						runtime.Gosched()
					} else {
						t.Errorf("意外的出队错误: %v", err)
						return
					}
				}
			}
		}(i)
	}

	// 等待所有生产者完成
	wgProducers.Wait()

	// 等待所有消费者完成或超时
	done := make(chan struct{})
	go func() {
		wgConsumers.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 消费者正常完成
	case <-time.After(5 * time.Second):
		t.Log("消费者等待超时，可能有死锁")
		cancel() // 取消所有消费者
		wgConsumers.Wait()
	}

	// 验证结果
	t.Logf("入队: %d, 出队: %d", enqueued.Load(), dequeued.Load())
	assert.Equal(t, enqueued.Load(), dequeued.Load(), "入队和出队的数量应该相等")
	assert.True(t, queue.IsEmpty() || queue.Size() == 0, "测试结束后队列应该为空")
}

// 测试超时操作
func TestTimeoutOperations(t *testing.T) {
	queue := New(8)

	// 填满队列
	for i := 0; i < 8; i++ {
		err := queue.Enqueue(i)
		assert.NoError(t, err)
	}

	// 测试入队超时
	startTime := time.Now()
	err := queue.EnqueueTimeout(100, 50*time.Millisecond)
	elapsed := time.Since(startTime)

	assert.Equal(t, ErrTimeout, err, "应该返回超时错误")
	assert.True(t, elapsed >= 50*time.Millisecond, "应该等待至少50ms")
	assert.True(t, elapsed < 500*time.Millisecond, "不应该等待太久")

	// 清空队列
	for i := 0; i < 8; i++ {
		_, err := queue.Dequeue()
		assert.NoError(t, err)
	}

	// 测试出队超时
	startTime = time.Now()
	_, err = queue.DequeueTimeout(50 * time.Millisecond)
	elapsed = time.Since(startTime)

	assert.Equal(t, ErrTimeout, err, "应该返回超时错误")
	assert.True(t, elapsed >= 50*time.Millisecond, "应该等待至少50ms")
	assert.True(t, elapsed < 500*time.Millisecond, "不应该等待太久")
}

// 测试非阻塞操作
func TestNonBlockingOperations(t *testing.T) {
	queue := New(4)

	// 测试TryEnqueue成功
	success := queue.TryEnqueue(1)
	assert.True(t, success, "TryEnqueue应该成功")
	assert.Equal(t, uint64(1), queue.Size())

	// 填满队列
	queue.TryEnqueue(2)
	queue.TryEnqueue(3)
	queue.TryEnqueue(4)

	// 测试TryEnqueue失败（队列已满）
	success = queue.TryEnqueue(5)
	assert.False(t, success, "TryEnqueue应该失败，因为队列已满")

	// 测试TryDequeue成功
	val, success := queue.TryDequeue()
	assert.True(t, success, "TryDequeue应该成功")
	assert.Equal(t, 1, val, "出队的第一个值应该是1")

	// 清空队列
	queue.TryDequeue()
	queue.TryDequeue()
	queue.TryDequeue()

	// 测试TryDequeue失败（队列为空）
	_, success = queue.TryDequeue()
	assert.False(t, success, "TryDequeue应该失败，因为队列为空")
}

// 测试扩展功能（多轮入队出队，检查序列号机制）
func TestSequenceNumberMechanism(t *testing.T) {
	// 使用较小的队列来快速测试序列号循环
	queue := New(4)

	// 进行多轮入队出队，检查序列号机制是否正常工作
	for round := 0; round < 10; round++ {
		// 完全填充队列
		for i := 0; i < 4; i++ {
			value := fmt.Sprintf("round%d-item%d", round, i)
			err := queue.Enqueue(value)
			assert.NoError(t, err, "入队操作不应该出错")
		}

		assert.True(t, queue.IsFull(), "队列应该已满")

		// 完全清空队列
		for i := 0; i < 4; i++ {
			expected := fmt.Sprintf("round%d-item%d", round, i)
			val, err := queue.Dequeue()
			assert.NoError(t, err, "出队操作不应该出错")
			assert.Equal(t, expected, val, "出队的值应该按顺序")
		}

		assert.True(t, queue.IsEmpty(), "队列应该为空")
	}

	// 测试序列号循环的边界情况 - 交错的入队出队
	for i := 0; i < 100; i++ {
		err := queue.Enqueue(i)
		assert.NoError(t, err, "入队操作不应该出错")

		val, err := queue.Dequeue()
		assert.NoError(t, err, "出队操作不应该出错")
		assert.Equal(t, i, val, "出队的值应该与入队的值相同")
	}
}

// 原始测试用例（保留但不使用）
func _TestRingBuffer(t *testing.T) {
	t.Skip("这是原始测试，已被新的测试用例替代")

	queue := New(512)

	timeout := 10 * time.Second
	ctx, cancel := context.WithCancel(context.Background())

	var enWait sync.WaitGroup
	var enCount atomic.Uint64
	for i := 0; i < 8; i++ {
		enWait.Add(1)
		go func(ctx context.Context, queue *Queue) {
			defer enWait.Done()
			for {
				select {
				case <-ctx.Done():
					log.DebugFields("ctx done")
					return
				default:
					if err := queue.Enqueue("ringbuffer"); err != nil {
						log.DebugFields("full", zap.Uint64("size", queue.Size()))
					}
					log.DebugFields("Enqueue", zap.Uint64("size", queue.Size()))
					enCount.Add(1)
				}
			}
		}(ctx, queue)
	}

	var deWait sync.WaitGroup
	var deCount atomic.Uint64
	for i := 0; i < 8; i++ {
		deWait.Add(1)
		go func(ctx context.Context, queue *Queue) {
			defer deWait.Done()
			for {
				select {
				case <-ctx.Done():
					log.DebugFields("ctx done")
					return
				default:
					if _, err := queue.Dequeue(); err != nil {
						log.DebugFields("empty", zap.Uint64("size", queue.Size()))
					}
					log.DebugFields("Dequeue", zap.Uint64("size", queue.Size()))
					deCount.Add(1)
				}
			}
		}(ctx, queue)
	}

	time.Sleep(timeout)

	cancel()
	enWait.Wait()
	deWait.Wait()

	log.DebugFields("", zap.Uint64("enCount", enCount.Load()), zap.Uint64("deCount", deCount.Load()))
}
