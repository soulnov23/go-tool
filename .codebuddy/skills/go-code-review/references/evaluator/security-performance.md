# Go 安全性、并发安全、内存优化与运行时性能规范（评估器参考）

> 供评估器使用：安全、并发、内存、性能维度的详细审查规则与判定标准。
>
> 基于 Go 官方规范、Google Go Style Guide、Uber Go Style Guide、OWASP Go-SCP。

---

## 一、安全性

### 1.1 加密随机数

**规则**：生成密钥、令牌、session ID 必须使用 `crypto/rand`，禁止 `math/rand`。

```go
// ✅ 通过
import "crypto/rand"
token := make([]byte, 32)
if _, err := rand.Read(token); err != nil {
    return err
}

// ❌ 失败（🔴 严重）
import "math/rand"
token := fmt.Sprintf("%d", rand.Int())
```

**判定标准**：
- 🔴 严重：任何安全上下文中使用 `math/rand`
- 搜索 `math/rand` 导入，检查所有使用场景

### 1.2 密码哈希

- 🔴 使用 MD5/SHA 系列做密码哈希
- ✅ 使用 `bcrypt` / `argon2`

### 1.3 API 边界防御

**规则**：切片和 map 在 API 边界处必须复制。

```go
// ✅ 通过：返回副本
func (s *Store) Items() []string {
    result := make([]string, len(s.items))
    copy(result, s.items)
    return result
}

// ❌ 失败（🔴 严重）：直接返回内部状态
func (s *Store) Items() []string {
    return s.items
}
```

**判定**：检查导出方法返回的切片/map 是否为内部状态的直接引用。

### 1.4 敏感信息保护

- 🔴 日志中打印密码、Token、身份证号、银行卡号
- 🔴 源码中硬编码密钥
- 🔴 错误信息暴露内部实现

**搜索模式**：`password`、`token`、`secret`、`key` 在日志调用中的出现。

### 1.5 输入验证

- 🔴 外部输入未经验证直接使用
- 验证类型、长度、范围、格式
- 白名单优于黑名单

### 1.6 SQL 注入

```go
// ✅ 通过
db.Query("SELECT * FROM users WHERE id = ?", userID)

// ❌ 失败（🔴 严重）
db.Query("SELECT * FROM users WHERE id = " + userID)
```

**搜索模式**：`db.Query`、`db.Exec` 等调用中包含字符串拼接（`+`）。

### 1.7 文件路径安全

- 🔴 用户输入直接拼接文件路径未经 `filepath.Clean` 和前缀验证
- 搜索 `filepath.Join` 调用，检查参数来源

### 1.8 HTTP 安全

```go
// ✅ 设置超时
srv := &http.Server{
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}

// ❌ 使用默认超时（无限制）
srv := &http.Server{}
```

- 🟡 未设置请求超时
- 🟡 未限制请求体大小

### 1.9 依赖安全

- 🟡 未配置 `govulncheck`
- 🟡 Go 版本过旧

---

## 二、并发安全

### 2.1 Goroutine 生命周期

**核心规则**：每个 goroutine 必须有可预测的停止时间。

```go
// ✅ 通过：context + errgroup
func (s *Server) Start(ctx context.Context) error {
    g, ctx := errgroup.WithContext(ctx)
    g.Go(func() error { return s.serve(ctx) })
    return g.Wait()
}

// ❌ 失败（🔴 严重）：即发即弃
func (s *Server) Start() {
    go s.serve()
}
```

**判定**：每个 `go` 关键字处检查是否有退出机制。

### 2.2 Goroutine 数量控制

- 🔴 `for` 循环中无限制 `go func()`
- ✅ 使用 semaphore / worker pool / errgroup

**搜索模式**：`for` 循环内的 `go func` 或 `go ` 调用。

### 2.3 Context 使用

- 🔴 context 不是第一个参数
- 🔴 context 存入结构体
- 🟡 传递 nil context
- 🟡 `context.WithValue` 传业务数据

### 2.4 Mutex 使用

```go
// ✅ 通过：未导出字段
type Cache struct {
    mu    sync.Mutex
    items map[string]string
}

// ❌ 失败（🔴 严重）：嵌入导致导出
type Cache struct {
    sync.Mutex
    items map[string]string
}
```

- 🔴 锁内执行 I/O / 网络请求
- 🔴 复制含 Mutex 的结构体
- 🟡 读多写少场景用 Mutex 而非 RWMutex

### 2.5 死锁预防

- 🔴 锁嵌套顺序不一致（A→B 与 B→A）
- 🟡 channel 操作无超时
- 搜索嵌套的 `Lock()` 调用

### 2.6 Channel

- 🔴 send on closed channel
- 🟡 channel 大小非 0 或 1（需设计依据）
- 检查 channel 关闭者是否为发送方

### 2.7 WaitGroup

```go
// ✅ Add 在 go 之前
wg.Add(1)
go func() {
    defer wg.Done()
}()

// ❌ 失败（🔴 严重）：Add 在 goroutine 内
go func() {
    wg.Add(1)
    defer wg.Done()
}()
```

### 2.8 init() 函数

- 🔴 init() 中执行 I/O、网络、启动 goroutine
- ✅ init() 仅用于简单注册

### 2.9 全局变量

- 🔴 可变的包级变量（非只读）
- ✅ 使用依赖注入

### 2.10 高级并发

**并发原语选择参考**：

| 场景 | 推荐 | 审查关注 |
|------|------|---------|
| 计数器/标志位 | `atomic` | 热路径是否有更好选择 |
| 读多写少 | `sync.RWMutex` | 是否误用 Mutex |
| 一般互斥 | `sync.Mutex` | 锁范围是否最小 |
| 一次性初始化 | `sync.Once` | 是否有 init 可替代 |
| 等待多协程 | `errgroup` | 是否有错误传播 |
| 限制并发数 | 带缓冲 channel | 是否有上限 |

**伪共享**：高频 atomic 字段间应有缓存行填充。
**CAS 自旋**：失败后 `runtime.Gosched()`，有退出条件。

---

## 三、内存分配与 GC 压力

### 3.1 逃逸分析

- 返回局部变量指针 → 逃逸到堆
- `any`/`interface{}` 参数 → 值装箱逃逸
- 小结构体（≤4 字段）考虑传值

```go
// ✅ 传值不逃逸
type Point struct{ X, Y int }
func Distance(a, b Point) float64 { ... }

// ❌ 不必要指针导致堆分配
func Distance(a, b *Point) float64 { ... }
```

**工具**：`go build -gcflags="-m"` 分析逃逸。

### 3.2 sync.Pool

- 高频临时对象使用 `sync.Pool`
- 🔴 归还前未重置状态

### 3.3 预分配

```go
// ✅ 通过
result := make([]int, 0, len(input))
m := make(map[string]int, len(keys))
var b strings.Builder
b.Grow(estimatedSize)

// ❌ 失败（🔴 严重，当集合较大时）
var result []int // 未预分配
```

**判定**：已知大小（`len(input)` 可用）但未预分配 → 标记。

### 3.4 字符串处理

| 场景 | ✅ 推荐 | ❌ 避免 |
|------|---------|---------|
| 基本类型转字符串 | `strconv.Itoa(42)` | `fmt.Sprintf("%d", 42)` |
| 循环拼接 | `strings.Builder` | `+=` 拼接 |
| 固定字节切片 | 预先转换重用 | 循环中重复 `[]byte("...")` |

### 3.5 闭包

- 🟡 闭包捕获不必要的大对象

### 3.6 热路径

- 🟢 预定义哨兵错误替代热路径 `fmt.Errorf`
- 🟢 避免热路径频繁创建临时对象

---

## 四、运行时性能

### 4.1 值传递 vs 指针传递

| 场景 | 推荐 | 判定 |
|------|------|------|
| 需要修改接收者 | 指针 | ✅ |
| 含 sync.Mutex | 指针 | 值传递 → 🔴 |
| 大结构体（>64字节） | 指针 | ✅ |
| 小型不可变结构体 | 值 | 指针 → 🟢 |
| map/func/chan | 值（已是引用） | 指针 → 🟡 |

**一致性**：同一类型不混用值/指针接收者 → 混用为 🟡。

### 4.2 同步 vs 异步

- 优先同步函数
- 让调用者决定并发

### 4.3 其他

- `any` 替代 `interface{}`（Go 1.18+）
- Printf 格式字符串声明为 `const`

---

## 五、工具链检查

审查时确认项目是否配置以下工具：

| 工具 | 用途 | 缺失严重程度 |
|------|------|------------|
| `gofmt` / `goimports` | 格式化 | 🟡 |
| `go vet` | 静态分析 | 🟡 |
| `golangci-lint` | 聚合 lint | 🟡 |
| `go test -race` | 竞态检测 | 🔴（有并发代码时）|
| `govulncheck` | 漏洞扫描 | 🟡 |

**推荐的验证命令**：

```bash
go build ./...
go vet ./...
go test -race ./...
golangci-lint run ./...
govulncheck ./...
go build -gcflags="-m" ./...  # 逃逸分析
go test -bench=. -benchmem ./...  # 基准测试
```
