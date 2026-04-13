# Go 代码风格与命名规范（生成器参考）

> 供生成器使用：修复代码时遵循的风格规范。每条规范附带正反示例，修复时严格按项目风格基线选择。

---

## 一、格式化

### 1.1 必须使用 gofmt

所有修复后的代码必须通过 `gofmt`（或 `goimports`）格式化。

### 1.2 行长度

- 软性限制 99 字符（Uber 建议）
- 过长行优先重构（提取变量/函数），非机械拆行
- `if` 条件不换行

---

## 二、命名规范

### 2.1 文件名

```go
// ✅ 正确
user_service.go
user_service_test.go

// ❌ 错误
UserService.go
userservice.GO
```

### 2.2 总体原则：MixedCaps

- 导出：`MaxLength`、`ServeHTTP`
- 未导出：`maxLength`、`serveHTTP`
- 常量：`MaxPacketSize`（不用 `MAX_PACKET_SIZE` 或 `kMaxPacketSize`）

### 2.3 缩略词

| ✅ 正确 | ❌ 错误 |
|---------|---------|
| `URL` | `Url` |
| `ID` | `Id` |
| `HTTP` | `Http` |
| `appID` | `appId` |
| `ServeHTTP` | `ServeHttp` |

### 2.4 包名

- 全小写、单词不拆分、无下划线
- 避免 `util`/`common`/`base`/`misc`/`helper`
- 包内标识符不重复包名：`widget.New`（非 `widget.NewWidget`）

### 2.5 变量名

- 长度与作用域成正比
- 不含类型信息：`users`（非 `userSlice`）
- 短作用域：`i`/`c`/`r`/`w`
- 包级变量：`defaultTimeout`/`errorHandler`

### 2.6 函数名

```go
// ✅ 正确
func (u *User) Name() string     // Getter 不加 Get
func (u *User) SetName(n string) // Setter 加 Set
func NewServer(opts ...Option) *Server // 构造函数 New 前缀
func WithPort(port int) Option         // Option 函数 With 前缀
func (u *User) IsActive() bool        // 布尔方法 Is 前缀
func MustParse(s string) *Config       // Must 函数

// ❌ 错误
func (u *User) GetName() string
func CreateServer() *Server    // 应用 New
```

### 2.7 接收者名称

```go
// ✅ 1-2 字母缩写，同类型所有方法一致
func (s *Server) Start() error
func (s *Server) Stop() error

// ❌ 禁止 this/self/me 或过长
func (this *Server) Start() error
func (server *Server) Start() error
```

### 2.8 接口命名

- 单方法：`-er` 后缀（`Reader`/`Writer`/`Closer`）
- 多方法：描述性名词
- 避免 `Manager`/`Helper`

### 2.9 未导出全局变量

```go
var (
    _defaultPort = 8080
    _maxRetries  = 3
)
var (
    errNotFound = errors.New("not found")
    errTimeout  = errors.New("timeout")
)
```

### 2.10 全项目命名一致性

修复命名一致性问题时，统一到项目主流风格：

| 概念 | 选定后全项目统一 | 禁止混用 |
|------|----------------|---------|
| 上下文 | `ctx` | `c` / `context` / `reqCtx` |
| 错误 | `err` | `e` / `error` / `er` |
| HTTP 请求 | `req` 或 `r` | 混用 |
| HTTP 响应 | `resp` 或 `w` | 混用 |
| 数据库连接 | `db` | `conn` / `database` |
| 日志 | `logger` 或 `log` | 混用 |
| 配置 | `cfg` 或 `config` | 混用 |
| 互斥锁 | `mu` | `lock` / `mtx` |

---

## 三、注释与文档

### 修复时遵循

```go
// ✅ 导出标识符必须有 godoc 注释
// Server represents an HTTP server that handles incoming requests.
type Server struct { ... }

// NewServer creates a new Server with the given options.
func NewServer(opts ...Option) *Server { ... }
```

- 注释以标识符名称开头，完整句子，句号结尾
- 解释"为什么"而非重复"是什么"
- 含义不明的参数（特别是 bool）添加 `/* paramName */` 注释

---

## 四、代码组织

### 4.1 导入分组

```go
import (
    // 标准库
    "context"
    "fmt"

    // 第三方库
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    // 项目内部包
    "myproject/internal/config"
)
```

### 4.2 函数排序

1. 类型定义 → 2. 构造函数 → 3. 导出方法 → 4. 未导出方法 → 5. 辅助函数

### 4.3 结构体初始化

```go
// ✅ 使用字段名，省略零值
svc := &Server{
    addr:    "localhost",
    port:    8080,
    timeout: 30 * time.Second,
}
```

### 4.4 声明风格

- 显式赋值用 `:=`
- 空切片用 `var t []string`（nil 切片）
- 检查空切片用 `len(s) == 0`（非 `s == nil`）
- 空 map 用 `make()`，固定元素用字面量
- 枚举用 `iota`，通常从非零值开始

### 4.5 函数调用写法统一

修复时必须遵循项目已确立的写法：

**错误包装**——选一种：
```go
// 方式 A：fmt.Errorf + %w
return fmt.Errorf("query user: %w", err)
// 方式 B：errors.Wrap
return errors.Wrap(err, "query user")
```

**错误消息格式**——选一种：
```go
// 风格 A：动词短语（推荐）
return fmt.Errorf("query user by id: %w", err)
// 风格 B：failed to 句式
return fmt.Errorf("failed to query user: %w", err)
```

**HTTP 响应**——选一种：
```go
// 方式 A：
json.NewEncoder(w).Encode(resp)
// 方式 B：
data, _ := json.Marshal(resp)
w.Write(data)
```

**类型转换**——选一种：
```go
s := strconv.Itoa(n)        // 方式 A
s := fmt.Sprintf("%d", n)   // 方式 B
```

---

## 五、控制流

### 修复时遵循

```go
// ✅ 提前返回，减少嵌套
func process(data []byte) error {
    if len(data) == 0 {
        return errors.New("empty data")
    }
    if !isValid(data) {
        return errors.New("invalid data")
    }
    // 正常处理...
    return nil
}
```

- 先处理错误并返回，正常逻辑保持最小缩进
- 不用 `else` 包裹正常代码路径
- 减少变量作用域：`if err := doSomething(); err != nil {`

---

## 六、切片与 Map

```go
// ✅ 空切片声明
var t []string

// ✅ 检查空切片
if len(s) == 0 { ... }

// ✅ Map 初始化
m := make(map[string]int, 100)  // 已知容量
m := map[string]int{"a": 1}     // 固定元素
```

---

## 七、枚举与常量

```go
// ✅ iota 从非零值开始
type Status int
const (
    _ Status = iota
    StatusActive
    StatusInactive
)
```

---

## 八、序列化标签

```go
// ✅ 所有序列化字段必须有标签
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age,omitempty"`
}
```

---

## 九、日志规范

### 9.1 统一日志库调用

修复时按项目使用的日志库风格：

```go
// zap 风格
logger.Info("user login",
    zap.String("user_id", userID),
    zap.String("ip", clientIP),
)

// slog 风格
slog.Info("user login",
    "user_id", userID,
    "ip", clientIP,
)
```

### 9.2 日志字段命名统一

| 语义 | 推荐 key（snake_case）| 禁止混用 |
|------|---------------------|---------|
| 用户标识 | `user_id` | `userId`/`UserID`/`uid` |
| 请求标识 | `request_id` | `requestId`/`reqId` |
| HTTP 方法 | `method` | `http_method`/`httpMethod` |
| 请求路径 | `path` | `url`/`uri`/`endpoint` |
| 状态码 | `status_code` | `statusCode`/`status` |
| 耗时 | `latency` 或 `duration` | 混用 |
| 错误 | `error`（日志库内置）| `err`/`errmsg` |

### 9.3 同类操作日志必备字段

```go
// HTTP 请求日志：request_id, method, path, status_code, latency
// 数据库操作日志：request_id, query, latency
// 错误日志：request_id, [业务ID], error
```

---

## 十、格式字符串

```go
// ✅ Printf 格式字符串声明为 const
const fmtStr = "invalid input: %s"
return fmt.Errorf(fmtStr, input)
```

Printf 风格函数名以 `f` 结尾（`Wrapf`、`Errorf`）。
