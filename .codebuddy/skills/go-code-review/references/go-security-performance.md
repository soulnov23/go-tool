# Go 安全性、并发安全、内存优化与运行时性能规范

本文档涵盖 Go 代码审查中安全性、并发安全、内存/GC 优化和运行时性能的详细规则，
基于 Go 官方规范、Google Go Style Guide、Uber Go Style Guide 和 OWASP Go-SCP。

---

## 一、安全性

### 1.1 加密随机数

**强制规则**：生成密钥、令牌、session ID 等安全敏感的随机值时，
**必须**使用 `crypto/rand`，**禁止**使用 `math/rand`。

```go
// ✅ 正确：加密安全的随机数
import "crypto/rand"

token := make([]byte, 32)
if _, err := rand.Read(token); err != nil {
    return err
}
// 或使用 Go 1.22+ 的 rand.Text()

// ❌ 严重安全漏洞
import "math/rand"

token := fmt.Sprintf("%d", rand.Int()) // 可预测！
```

### 1.2 密码哈希

- 密码哈希使用 `bcrypt` / `argon2`，**禁用** MD5/SHA 系列
- 使用标准库加密实现，不自行实现加密算法

### 1.3 API 边界防御

切片和 map 在 API 边界处（函数入参/返回值被外部代码持有时），
必须进行复制，防止外部代码意外修改内部状态。

```go
// ✅ 正确：返回内部切片的副本
type Store struct {
    mu    sync.Mutex
    items []string
}

func (s *Store) Items() []string {
    s.mu.Lock()
    defer s.mu.Unlock()
    result := make([]string, len(s.items))
    copy(result, s.items)
    return result
}

// ❌ 危险：外部代码可修改内部状态
func (s *Store) Items() []string {
    return s.items
}
```

同理，接收切片/map 参数时也应在需要时复制：

```go
// ✅ 正确：存储前复制
func (s *Store) SetItems(items []string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.items = make([]string, len(items))
    copy(s.items, items)
}
```

### 1.4 敏感信息保护

- 不将密码、密钥、token 等敏感信息写入日志
- 不将用户个人隐私信息（PII）以明文形式持久化或传输
- 错误信息中不泄露内部实现细节（文件路径、SQL 语句等）
- 配置中的密钥不硬编码在源码中

### 1.5 输入验证

- 对来自外部的输入（用户输入、API 参数、配置文件）进行合法性验证
- 验证数据类型、长度、范围、格式
- 白名单验证优于黑名单
- 防范整数溢出

### 1.6 SQL 注入防护

- [严重] 使用参数化查询/预编译语句，禁止字符串拼接 SQL
- ORM 查询也需审查动态拼接部分

```go
// ✅ 参数化查询
db.Query("SELECT * FROM users WHERE id = ?", userID)

// ❌ 字符串拼接（SQL 注入风险）
db.Query("SELECT * FROM users WHERE id = " + userID)
```

### 1.7 文件路径安全

- [严重] 防止路径遍历攻击（`../` 注入）
- 使用 `filepath.Clean` 和白名单验证路径

### 1.8 HTTP 安全

- 设置合理的请求超时（读/写/空闲超时）
- 限制请求体大小（`http.MaxBytesReader`）
- 使用 HTTPS
- 设置安全相关 HTTP 头

```go
// ✅ 设置 HTTP 超时
srv := &http.Server{
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}
```

### 1.9 依赖安全

- 定期运行 `govulncheck` 扫描依赖漏洞
- 保持 Go 版本和依赖为最新（注意审查更新）
- 使用 `go.sum` 确保依赖完整性

---

## 二、并发安全

### 2.1 Goroutine 生命周期管理

**核心原则**：每个 goroutine 都必须有可预测的停止时间或停止信号。

```go
// ✅ 正确：使用 context + errgroup 控制 goroutine 生命周期
func (s *Server) Start(ctx context.Context) error {
    g, ctx := errgroup.WithContext(ctx)
    g.Go(func() error {
        return s.serve(ctx)
    })
    g.Go(func() error {
        return s.watchConfig(ctx)
    })
    return g.Wait()
}

// ❌ 错误：即发即弃，无法控制
func (s *Server) Start() {
    go s.serve()       // 如何停止？
    go s.watchConfig() // 如何等待退出？
}
```

### 2.2 Goroutine 数量控制

- [严重] 控制 goroutine 创建数量（worker pool / semaphore）
- 高并发场景需要限流/背压机制
- 不要 fire-and-forget goroutine

```go
// ✅ 使用 semaphore 控制并发数
sem := make(chan struct{}, maxWorkers)
for _, task := range tasks {
    sem <- struct{}{} // 获取信号量
    go func(t Task) {
        defer func() { <-sem }() // 释放信号量
        process(t)
    }(task)
}
```

### 2.3 Context 使用规范

- `context.Context` 作为函数的**第一个**参数传递
- 参数名统一为 `ctx`
- 不把 Context 存入结构体成员
- 不创建自定义 Context 类型
- 即使当前不需要，也应传递 Context（为未来扩展预留）
- 不传递 `nil` Context，不确定时使用 `context.TODO()`
- 不用 `context.WithValue` 传递业务数据
- 使用 `context.WithTimeout` / `context.WithCancel` 控制超时和取消

```go
// ✅ 正确
func (s *Server) Handle(ctx context.Context, req *Request) error

// ❌ 错误
func (s *Server) Handle(req *Request, ctx context.Context) error
type Server struct {
    ctx context.Context // 不应存入结构体
}
```

### 2.4 同步原语使用

#### Mutex

- `sync.Mutex` 零值可用，不需要初始化
- 不要复制包含 Mutex 的结构体
- 不要在导出结构体中嵌入 Mutex（避免泄露锁接口）
- Mutex 应作为未导出字段，靠近它保护的数据
- [严重] 锁范围最小化，不在锁内执行 I/O / 网络请求

```go
// ✅ 正确
type Cache struct {
    mu    sync.Mutex
    items map[string]string
}

// ❌ 错误：嵌入导致 Lock/Unlock 被导出
type Cache struct {
    sync.Mutex
    items map[string]string
}
```

#### RWMutex

- 读多写少场景用 `sync.RWMutex` 替代 `sync.Mutex`
- 读锁和写锁不要混用

#### Atomic

- 基本类型的原子操作优先使用 `sync/atomic` 包
- 考虑使用 `atomic.Bool`、`atomic.Int64` 等类型化包装（Go 1.19+）
- 热路径计数器/标志位可用 `atomic` 替代 `Mutex`

#### WaitGroup

- 使用 `sync.WaitGroup` 等待一组 goroutine 完成
- `wg.Add()` 必须在启动 goroutine **之前**调用

```go
var wg sync.WaitGroup
for i := 0; i < n; i++ {
    wg.Add(1) // 在 go 之前
    go func() {
        defer wg.Done()
        // 工作...
    }()
}
wg.Wait()
```

#### 并发原语选择表

| 场景 | 推荐 |
|------|------|
| 计数器/标志位 | `atomic` |
| 读多写少 | `sync.RWMutex` |
| 一般互斥 | `sync.Mutex` |
| 一次性初始化 | `sync.Once` |
| 并发安全 Map（读多写少） | `sync.Map` |
| 等待多协程完成 | `sync.WaitGroup` / `errgroup` |
| 信号通知/退出 | `channel` / `context` |
| 限制并发数 | 带缓冲 channel / semaphore |

### 2.5 死锁预防

- [严重] 锁嵌套顺序全局一致（避免 A→B 与 B→A）
- channel 操作有超时或 default 分支
- 数据库事务避免互相等待

### 2.6 Channel 使用

- Channel 大小通常为 0（无缓冲）或 1（缓冲）
- 其他大小需要明确的设计依据和审慎评估
- 优先使用同步函数而非 channel 通信
- 如需并发，由调用者决定是否启动 goroutine
- [严重] 避免 send on closed channel（panic）
- 明确谁负责关闭 channel（通常由发送方关闭）

### 2.7 init() 函数

- `init()` 中不启动 goroutine
- `init()` 中不做耗时操作或 I/O
- 如需后台 goroutine，导出管理其生命周期的对象

### 2.8 避免可变全局变量

- 库代码不使用可变的包级变量
- 使用依赖注入替代全局状态
- 全局变量导致测试困难和并发安全问题

### 2.9 高级并发话题

**伪共享**：高频 `atomic` 字段间应有缓存行填充，避免同一缓存行上的竞争。

**CAS 自旋**：失败后 `runtime.Gosched()` 让出 CPU，并设有退出条件。

---

## 三、内存分配与 GC 压力

### 3.1 逃逸分析

- 返回局部变量指针会导致变量逃逸到堆
- `any`/`interface{}` 参数导致值装箱逃逸
- 小结构体（≤4 字段）考虑传值而非传指针（减少堆分配）
- 使用 `go build -gcflags="-m"` 分析逃逸情况

```go
// ✅ 传值不逃逸
type Point struct{ X, Y int }
func Distance(a, b Point) float64 { ... }

// ❌ 不必要的指针导致堆分配
func Distance(a, b *Point) float64 { ... }
```

### 3.2 sync.Pool

- 高频创建/销毁的临时对象使用 `sync.Pool`
- 归还前**必须**正确重置状态

```go
var bufPool = sync.Pool{
    New: func() any {
        return new(bytes.Buffer)
    },
}

func process() {
    buf := bufPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset() // 重置状态
        bufPool.Put(buf)
    }()
    // 使用 buf...
}
```

### 3.3 预分配

- [严重] 已知大小的切片 `make([]T, 0, n)` 预分配
- Map `make(map[K]V, n)` 预分配
- `strings.Builder` 使用 `Grow()` 预分配

```go
// ✅ 正确：预分配
result := make([]int, 0, len(input))
for _, v := range input {
    if v > 0 {
        result = append(result, v)
    }
}

m := make(map[string]int, len(keys))

var b strings.Builder
b.Grow(estimatedSize)

// ❌ 未预分配（频繁扩容）
var result []int
for _, v := range input {
    result = append(result, v)
}
```

### 3.4 字符串处理

#### strconv vs fmt

基本类型与字符串转换时，`strconv` 比 `fmt` 快。

```go
// ✅ 更快
s := strconv.Itoa(42)
s := strconv.FormatFloat(3.14, 'f', -1, 64)

// ❌ 较慢
s := fmt.Sprintf("%d", 42)
s := fmt.Sprintf("%f", 3.14)
```

#### 字符串拼接

| 场景 | 推荐方式 |
|------|----------|
| 简单拼接 | `+` 运算符 |
| 格式化 | `fmt.Sprintf` |
| 循环中逐步构建 | `strings.Builder` |

```go
// ✅ 循环中构建字符串
var b strings.Builder
for _, s := range items {
    b.WriteString(s)
    b.WriteByte(',')
}
result := b.String()

// ❌ 循环中字符串拼接（每次分配新内存）
var result string
for _, s := range items {
    result += s + ","
}
```

### 3.5 避免不必要的转换

不要在循环中重复将固定字符串转换为 `[]byte`，应预先转换并重用。

```go
// ✅ 预先转换
data := []byte("fixed string")
for i := 0; i < n; i++ {
    process(data)
}

// ❌ 循环中重复转换
for i := 0; i < n; i++ {
    process([]byte("fixed string"))
}
```

### 3.6 闭包优化

- 避免闭包捕获不必要的大对象
- 闭包引用外部变量会导致变量逃逸

### 3.7 热路径优化

- 用预定义哨兵错误替代热路径中的 `fmt.Errorf`
- 避免热路径中频繁创建临时对象

---

## 四、运行时性能

### 4.1 值传递 vs 指针传递

| 场景 | 推荐 |
|------|------|
| 需要修改接收者 | 指针接收者 |
| 包含 sync.Mutex 等不可复制字段 | 指针接收者 |
| 大型结构体（多字段/大于 64 字节） | 指针接收者 |
| 会增长的小结构体 | 指针接收者 |
| 小型、不可变的值类型（如 time.Time） | 值接收者 |
| 基本类型、string、interface | 值传递 |
| map、func、chan | 值传递（已是引用类型，不要传指针） |
| 不确定时 | 指针接收者 |

不要仅为了"节省几个字节"而使用指针。小对象的值拷贝通常比指针引用更高效（减少 GC 压力、提高缓存命中率）。

**同一类型不要混用值接收者和指针接收者。**

### 4.2 同步 vs 异步

- 优先使用同步函数，让调用者决定是否并发
- 同步函数更易推理、测试、调用

### 4.3 其他性能技巧

- 新代码用 `any` 替代 `interface{}`
- 使用 `%q` 打印带引号字符串
- Printf 格式字符串声明为 `const`，以便 `go vet` 静态分析

---

## 五、工具链

### 5.1 推荐的 Lint 工具集

审查时建议确认项目是否配置了以下 lint 工具：

| 工具 | 用途 |
|------|------|
| `gofmt` / `goimports` | 代码格式化和导入管理 |
| `go vet` | 常见错误静态分析 |
| `staticcheck` | 高级静态分析 |
| `errcheck` | 确保错误被处理 |
| `revive` | 常见风格问题 |
| `golangci-lint` | 聚合 lint 运行器 |

### 5.2 常用检测命令

```bash
# 数据竞争检测
go test -race ./...

# 逃逸分析
go build -gcflags="-m" ./...

# 漏洞扫描
govulncheck ./...

# 综合静态分析
golangci-lint run ./...

# 基准测试（含内存分配）
go test -bench=. -benchmem ./...
```

### 5.3 基准测试

性能敏感的代码应有基准测试：

```go
func BenchmarkProcess(b *testing.B) {
    data := prepareTestData()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        Process(data)
    }
}
```
