# Go 代码风格与命名规范

本文档汇总 Go 语言代码风格和命名的详细规范，基于 Go 官方 Code Review Comments、
Google Go Style Guide 和 Uber Go Style Guide。

---

## 一、格式化

### 1.1 必须使用 gofmt

所有 Go 源文件必须通过 `gofmt`（或 `goimports`）格式化。这是 Go 社区的强制要求，
没有任何商量余地。推荐使用 `goimports` 作为 `gofmt` 的超集，自动管理导入。

### 1.2 行长度

Go 没有严格的行长度限制。建议参考值：

- Uber 建议软性限制为 99 个字符
- Google 建议在 80-100 列处换行

不要仅为了凑行数而拆行。语义完整的一行通常比强行拆断的行更易读。
如果行太长，优先考虑重构（提取变量/函数）而非机械拆行。

---

## 二、命名规范

### 2.1 文件名

- 文件名使用小写 + 下划线（`snake_case.go`）
- 测试文件以 `_test.go` 结尾
- 文件名应反映文件中的主要内容

```go
// ✅ 正确
user_service.go
user_service_test.go
http_handler.go

// ❌ 错误
UserService.go
userservice.GO
```

### 2.2 总体原则

Go 命名必须使用 **MixedCaps**（驼峰式），禁止使用 snake_case（下划线分隔）。

- 导出名称：`MaxLength`、`ServeHTTP`
- 未导出名称：`maxLength`、`serveHTTP`
- 常量：`MaxPacketSize`（不是 `MAX_PACKET_SIZE` 或 `kMaxPacketSize`）

### 2.2 缩略词大小写

缩略词必须保持大小写一致，不要混合。

| ✅ 正确 | ❌ 错误 |
|---------|---------|
| `URL` | `Url` |
| `ID` | `Id` |
| `HTTP` | `Http` |
| `appID` | `appId` |
| `ServeHTTP` | `ServeHttp` |
| `xmlHTTPRequest` | `XmlHttpRequest` |

### 2.3 包名

- 全小写，单个单词，不含下划线或大写字母
- 简短且有意义
- 避免使用 `util`、`common`、`base`、`misc`、`helper` 等无意义名称
- 包内标识符不重复包名：`widget.New`（不是 `widget.NewWidget`）

```go
// ✅ 好的包名
package config
package transport
package httputil

// ❌ 差的包名
package common_utils
package myHelpers
package base
```

### 2.4 变量名

- 变量名长度与作用域大小成正比
- 局部短作用域变量：简短（`i`、`c`、`r`、`w`）
- 包级变量或长作用域变量：描述性更强（`lineCount`、`errorHandler`）
- 不在变量名中包含类型信息：`users`（不是 `userSlice`）

```go
// ✅ 短作用域
for i, v := range items {
    // ...
}

// ✅ 包级变量
var defaultTimeout = 30 * time.Second
```

### 2.6 函数名

- 返回值的函数使用名词短语，不加 `Get` 前缀
- 执行操作的函数使用动词短语
- Getter 不要 `Get` 前缀：`Counts()`（不是 `GetCounts()`），除非涉及远程调用或复杂计算（此时用 `Compute`/`Fetch`）
- 构造函数：`New` / `NewXxx` 前缀，不混用 `Create`/`Make`/`Build`
- Option 函数：`With` 前缀（`WithTimeout`、`WithPort`）
- 布尔函数/方法：`Is`/`Has`/`Can`/`Should` 前缀
- Printf 风格函数名以 `f` 结尾（如 `Wrapf`、`Errorf`）
- Must 函数（启动时 panic）：`MustXxx` 前缀

```go
// ✅ 正确
func (u *User) Name() string     // Getter，不加 Get
func (u *User) SetName(n string) // Setter，加 Set
func ParseConfig(data []byte) (*Config, error) // 动词短语
func NewServer(opts ...Option) *Server          // 构造函数
func WithPort(port int) Option                  // Option 函数
func (u *User) IsActive() bool                  // 布尔方法
func MustParse(s string) *Config                // Must 函数

// ❌ 错误
func (u *User) GetName() string
func CreateServer() *Server     // 应用 New
func BuildConfig() *Config      // 应用 New
```

### 2.7 接收者名称

- 简短（通常 1-2 个字母），能反映类型
- 同一类型的所有方法使用一致的接收者名称
- 不使用 `this`、`self`、`me`

```go
// ✅ 正确
func (s *Server) Start() error
func (s *Server) Stop() error
func (c *Client) Do(req *Request) (*Response, error)

// ❌ 错误
func (this *Server) Start() error
func (self *Server) Stop() error
func (server *Server) Start() error // 过长
```

### 2.8 接口命名

- 单方法接口使用 `-er` 后缀（`Reader`、`Writer`、`Closer`）
- 多方法接口使用描述性名词
- 避免 `Manager`/`Helper`/`Handler`（除非语义明确如 `http.Handler`）

### 2.9 未导出全局变量

未导出的顶级变量和常量使用 `_` 前缀，以表明其全局作用域（错误值使用 `err` 前缀除外）。

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

---

## 三、注释与文档

### 3.1 文档注释

- 所有导出的顶层名称（类型、函数、常量、变量）必须有文档注释
- 注释以被描述实体的名称开头
- 注释是完整的句子，以句号结尾

```go
// Server represents an HTTP server that handles incoming requests.
type Server struct { ... }

// NewServer creates a new Server with the given options.
func NewServer(opts ...Option) *Server { ... }

// ErrNotFound indicates the requested resource was not found.
var ErrNotFound = errors.New("not found")
```

### 3.2 包注释

- 包注释紧邻 `package` 子句
- 一个包只有一个包注释（通常在 doc.go 或主文件中）
- `package main` 的注释以 "Binary ..." 或 "Command ..." 开头

```go
// Package config provides configuration loading and validation
// for the application.
package config
```

### 3.3 注释内容原则

- 注释应解释"**为什么**"（why），而非重复"**是什么**"（what）
- 复杂业务逻辑和性能优化必须注释说明原因
- 避免冗余注释（代码本身能说明的不需要注释）
- TODO 注释包含关联的 issue 编号或负责人
- 函数调用中含义不明显的参数（特别是 bool）添加注释 `/* paramName */`
- 使用原始字符串字面量（反引号 `` ` ``）避免转义
- 避免过时注释与代码不一致

```go
// ✅ 好的注释：解释为什么
// 使用 LRU 缓存而非简单 map，因为内存受限环境下需要控制缓存大小。
// 缓存大小 1000 基于线上 P99 请求的不同 key 数量统计。
cache := lru.New(1000)

// ❌ 差的注释：复述代码
// 创建一个大小为 1000 的新 LRU 缓存
cache := lru.New(1000)
```

---

## 四、代码组织

### 4.1 声明分组

相关的 `import`、`const`、`var`、`type` 声明用括号分组在一起。

```go
const (
    defaultPort    = 8080
    defaultTimeout = 30 * time.Second
)

var (
    ErrNotFound = errors.New("not found")
    ErrTimeout  = errors.New("timeout")
)
```

### 4.2 函数排序

文件内函数按以下顺序组织：

1. 类型定义
2. 构造函数 `NewXxx`
3. 导出方法（按接收者分组）
4. 未导出方法
5. 辅助函数

### 4.3 结构体字段排序

- 嵌入字段放在最前面，用空行分隔
- 按逻辑分组（如配置项、运行时状态、同步原语）

```go
type Server struct {
    // 嵌入字段
    http.Handler

    // 配置
    addr    string
    port    int
    timeout time.Duration

    // 运行时状态
    started bool
    conns   map[string]*Conn

    // 同步
    mu sync.Mutex
}
```

### 4.4 结构体初始化

- 跨包初始化必须指定字段名
- 零值字段可以省略
- 零值结构体使用 `var` 声明
- 初始化指针使用 `&T{}` 而非 `new(T)`

```go
// ✅ 正确
svc := &Server{
    addr:    "localhost",
    port:    8080,
    timeout: 30 * time.Second,
}

var cfg Config // 零值

// ❌ 错误：不指定字段名
svc := &Server{"localhost", 8080, 30 * time.Second}
```

---

## 五、控制流

### 5.1 减少嵌套

通过提前返回减少嵌套层级（"happy path" 左对齐）。

```go
// ✅ 正确
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

// ❌ 过度嵌套
func process(data []byte) error {
    if len(data) > 0 {
        if isValid(data) {
            // 正常处理...
            return nil
        } else {
            return errors.New("invalid data")
        }
    } else {
        return errors.New("empty data")
    }
}
```

### 5.2 不必要的 else

如果 if 的两个分支都设置同一个变量，考虑简化。

```go
// ✅ 简洁
a := 10
if condition {
    a = 20
}

// ❌ 不必要的 else
var a int
if condition {
    a = 20
} else {
    a = 10
}
```

### 5.3 减少变量作用域

尽量在最小作用域内声明变量。

```go
// ✅ 限制作用域
if err := doSomething(); err != nil {
    return err
}
```

---

## 六、切片与 Map

### 6.1 空切片声明

声明空切片优先使用 `var`（nil 切片），而非字面量（非 nil 的空切片）。

```go
// ✅ 首选（nil 切片）
var t []string

// ❌ 通常不需要（非 nil 的空切片）
t := []string{}
```

注意：JSON 编码时 nil 切片输出 `null`，空切片输出 `[]`。如果 API 要求返回 `[]`，
则需要使用 `[]string{}`。

### 6.2 检查空切片

使用 `len(s) == 0` 检查切片是否为空，不要与 nil 比较。

```go
// ✅ 正确
if len(s) == 0 {
    // 处理空切片
}

// ❌ 错误（无法处理非 nil 的空切片）
if s == nil {
    // ...
}
```

### 6.3 Map 初始化

- 空 map 使用 `make`
- 固定元素使用字面量
- 尽量提供容量提示

```go
// 空 map
m := make(map[string]int)

// 已知容量
m := make(map[string]int, 100)

// 固定元素
m := map[string]int{
    "a": 1,
    "b": 2,
}
```

---

## 七、枚举与常量

### 7.1 枚举起始值

使用 `iota` 定义枚举时，通常从 1 开始（或使用 `_` 跳过 0），
因为变量的零值为 0，从 1 开始可以区分"未设置"和"第一个枚举值"。

```go
type Status int

const (
    _ Status = iota
    StatusActive
    StatusInactive
    StatusDeleted
)
```

除非 0 代表合理的默认值：

```go
type LogLevel int

const (
    LogInfo LogLevel = iota // 0 作为默认日志级别是合理的
    LogWarn
    LogError
)
```

---

## 八、序列化标签

任何需要序列化/反序列化的结构体字段都应使用对应的标签（Uber 标准）。

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age,omitempty"`
}
```

未加标签的导出字段在 JSON 序列化时使用字段名，可能导致 API 不稳定。

---

## 九、日志规范

### 9.1 统一日志库

- 项目内统一使用同一日志库，不混用（如统一使用 `zap`、`slog`、`logr` 等）
- 避免直接使用 `fmt.Println` 或 `log.Printf` 替代结构化日志

### 9.2 结构化日志

```go
// ✅ 结构化日志（Field 方式）
logger.Info("user login",
    zap.String("user_id", userID),
    zap.String("ip", clientIP),
)

// ❌ 字符串拼接日志
logger.Info(fmt.Sprintf("user %s login from %s", userID, clientIP))
```

### 9.3 日志级别

- Debug：调试信息，生产环境关闭
- Info：正常业务事件
- Warn：异常但不影响功能
- Error：需要关注的错误
- Fatal/Panic：仅在 main 或不可恢复时使用

### 9.4 注意事项

- 错误日志包含足够上下文（请求 ID、用户 ID、关键参数）
- 日志消息简短明确，风格全项目统一
- 热路径/循环中避免无限制日志输出
- [严重] 不在日志中打印敏感信息（密码、Token、PII）

---

## 十、格式字符串

Printf 风格的格式字符串声明为 `const`，以便 `go vet` 进行静态分析。

```go
// ✅ 可被 go vet 检查
const fmtStr = "invalid input: %s"
return fmt.Errorf(fmtStr, input)
```

Printf 风格函数名以 `f` 结尾（如 `Wrapf`、`Errorf`），以便工具检测格式字符串匹配。
