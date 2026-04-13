# 常见修复模式与代码模板（生成器参考）

> 供生成器使用：按问题类型分类的标准修复模式。修复时直接使用或参考这些模板。

---

## 一、错误处理修复

### 1.1 未处理的错误返回值

```go
// ❌ 问题：用 _ 丢弃错误
result, _ := doSomething()

// ✅ 修复：正确处理错误
result, err := doSomething()
if err != nil {
    return fmt.Errorf("do something: %w", err)
}
```

### 1.2 错误既记录又返回

```go
// ❌ 问题：记录日志后又返回，导致日志噪音
if err != nil {
    logger.Error("query failed", zap.Error(err))
    return err
}

// ✅ 修复：只做一件事（返回错误，让调用者决定是否记录）
if err != nil {
    return fmt.Errorf("query user: %w", err)
}
```

### 1.3 错误字符串格式

```go
// ❌ 问题：首字母大写 + 标点结尾
return fmt.Errorf("Failed to query user: %w.", err)

// ✅ 修复：小写开头，无标点
return fmt.Errorf("query user: %w", err)
```

### 1.4 错误包装丢失链

```go
// ❌ 问题：%v 丢失错误链
return fmt.Errorf("query user: %v", err)

// ✅ 修复：%w 保留错误链
return fmt.Errorf("query user: %w", err)
```

---

## 二、资源管理修复

### 2.1 未关闭资源

```go
// ❌ 问题：忘记关闭文件
f, err := os.Open(name)
if err != nil {
    return err
}
// 直接使用 f，未 close

// ✅ 修复：获取成功后立即 defer Close
f, err := os.Open(name)
if err != nil {
    return err
}
defer f.Close()
```

### 2.2 循环中 defer

```go
// ❌ 问题：循环中 defer 导致资源延迟释放
for _, name := range files {
    f, err := os.Open(name)
    if err != nil {
        return err
    }
    defer f.Close() // 所有文件在函数退出时才关闭！
    process(f)
}

// ✅ 修复：提取为函数或循环内关闭
for _, name := range files {
    if err := processFile(name); err != nil {
        return err
    }
}

func processFile(name string) error {
    f, err := os.Open(name)
    if err != nil {
        return err
    }
    defer f.Close()
    return process(f)
}
```

### 2.3 Context 使用修复

```go
// ❌ 问题：context 不是第一个参数
func Query(db *sql.DB, ctx context.Context, id int) error

// ✅ 修复：context 作为第一个参数
func Query(ctx context.Context, db *sql.DB, id int) error
```

---

## 三、并发安全修复

### 3.1 Goroutine 泄漏

```go
// ❌ 问题：goroutine 无退出机制
func (s *Server) Start() {
    go s.serve()
    go s.watchConfig()
}

// ✅ 修复：使用 context + errgroup
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
```

### 3.2 WaitGroup 位置错误

```go
// ❌ 问题：Add 在 goroutine 内部
for i := 0; i < n; i++ {
    go func() {
        wg.Add(1) // 可能在 Wait 之后执行
        defer wg.Done()
        process()
    }()
}
wg.Wait()

// ✅ 修复：Add 在 goroutine 启动前
for i := 0; i < n; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        process()
    }()
}
wg.Wait()
```

### 3.3 锁范围过大

```go
// ❌ 问题：锁内执行网络请求
mu.Lock()
data := cache[key]
if data == nil {
    data, err = fetchFromDB(key) // 锁内 I/O！
    cache[key] = data
}
mu.Unlock()

// ✅ 修复：最小化锁范围
mu.RLock()
data := cache[key]
mu.RUnlock()

if data == nil {
    data, err = fetchFromDB(key)
    if err != nil {
        return err
    }
    mu.Lock()
    cache[key] = data
    mu.Unlock()
}
```

### 3.4 Goroutine 无限制创建

```go
// ❌ 问题：无限制创建 goroutine
for _, task := range tasks {
    go process(task)
}

// ✅ 修复：使用 semaphore 控制并发
sem := make(chan struct{}, maxWorkers)
var wg sync.WaitGroup
for _, task := range tasks {
    wg.Add(1)
    sem <- struct{}{}
    go func(t Task) {
        defer func() {
            <-sem
            wg.Done()
        }()
        process(t)
    }(task)
}
wg.Wait()
```

---

## 四、安全漏洞修复

### 4.1 math/rand 生成安全随机数

```go
// ❌ 问题：可预测的随机数
import "math/rand"
token := fmt.Sprintf("%d", rand.Int())

// ✅ 修复：加密安全的随机数
import "crypto/rand"
import "encoding/hex"
b := make([]byte, 32)
if _, err := rand.Read(b); err != nil {
    return err
}
token := hex.EncodeToString(b)
```

### 4.2 SQL 注入

```go
// ❌ 问题：字符串拼接 SQL
query := "SELECT * FROM users WHERE id = " + userID
rows, err := db.Query(query)

// ✅ 修复：参数化查询
rows, err := db.Query("SELECT * FROM users WHERE id = ?", userID)
```

### 4.3 路径遍历

```go
// ❌ 问题：未验证路径
filePath := filepath.Join(baseDir, userInput)
data, err := os.ReadFile(filePath)

// ✅ 修复：清理并验证路径
cleanPath := filepath.Clean(userInput)
if strings.Contains(cleanPath, "..") {
    return fmt.Errorf("invalid path: %s", userInput)
}
filePath := filepath.Join(baseDir, cleanPath)
if !strings.HasPrefix(filePath, baseDir) {
    return fmt.Errorf("path traversal detected: %s", userInput)
}
data, err := os.ReadFile(filePath)
```

---

## 五、Nil 安全修复

### 5.1 Nil map 写入

```go
// ❌ 问题：向 nil map 写入（panic）
var m map[string]int
m["key"] = 1

// ✅ 修复：初始化 map
m := make(map[string]int)
m["key"] = 1
```

### 5.2 不安全的类型断言

```go
// ❌ 问题：单返回值类型断言（panic）
val := x.(string)

// ✅ 修复：安全类型断言
val, ok := x.(string)
if !ok {
    return fmt.Errorf("expected string, got %T", x)
}
```

### 5.3 API 边界切片/Map 防护

```go
// ❌ 问题：返回内部切片（外部可修改内部状态）
func (s *Store) Items() []string {
    return s.items
}

// ✅ 修复：返回副本
func (s *Store) Items() []string {
    result := make([]string, len(s.items))
    copy(result, s.items)
    return result
}
```

---

## 六、内存优化修复

### 6.1 未预分配切片

```go
// ❌ 问题：频繁 append 扩容
var result []int
for _, v := range input {
    result = append(result, v)
}

// ✅ 修复：预分配容量
result := make([]int, 0, len(input))
for _, v := range input {
    result = append(result, v)
}
```

### 6.2 热路径 fmt.Sprintf

```go
// ❌ 问题：热路径用 fmt
s := fmt.Sprintf("%d", n)

// ✅ 修复：用 strconv
s := strconv.Itoa(n)
```

### 6.3 循环中字符串拼接

```go
// ❌ 问题：循环中 += 拼接
var result string
for _, s := range items {
    result += s + ","
}

// ✅ 修复：使用 strings.Builder
var b strings.Builder
for _, s := range items {
    b.WriteString(s)
    b.WriteByte(',')
}
result := b.String()
```

---

## 七、一致性修复

### 7.1 命名统一

修复策略：统计项目中各变体的出现次数，统一到出现最多的写法。

```go
// 如果项目主流用 userID（出现 20 次），有 3 处用了 userId
// → 将 3 处 userId 全部改为 userID
```

### 7.2 错误包装统一

```go
// 如果项目主流用 fmt.Errorf + %w，有少数用 errors.Wrap
// → 将 errors.Wrap 统一改为 fmt.Errorf + %w
// 反之亦然
```

### 7.3 日志字段统一

```go
// 如果项目主流日志字段用 snake_case
// → 将 camelCase 的字段 key 全部改为 snake_case
// 如 "userId" → "user_id"
```

---

## 八、结构体与接口修复

### 8.1 接口定义位置

```go
// ❌ 问题：在实现方定义接口
// package store
type UserStore interface {
    GetUser(id int) (*User, error)
}
type userStore struct { ... }

// ✅ 修复：在使用方定义接口
// package handler
type UserGetter interface {
    GetUser(id int) (*User, error)
}
// package store（不定义接口，只返回具体类型）
type Store struct { ... }
```

### 8.2 嵌入 Mutex

```go
// ❌ 问题：嵌入导致 Lock/Unlock 被导出
type Cache struct {
    sync.Mutex
    items map[string]string
}

// ✅ 修复：作为未导出字段
type Cache struct {
    mu    sync.Mutex
    items map[string]string
}
```

### 8.3 复制含 Mutex 的结构体

```go
// ❌ 问题：值传递含 Mutex 的结构体
func process(c Cache) { ... }

// ✅ 修复：传指针
func process(c *Cache) { ... }
```
