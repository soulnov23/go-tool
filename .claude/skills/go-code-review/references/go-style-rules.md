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

### 2.10 全项目命名一致性

> 同一项目中，同一概念必须使用同一个名称，不允许不同文件/包中对同一概念使用不同的命名。

```go
// ✅ 全项目统一使用同一命名
userID   // 全项目统一，不要混用 userId / user_id / uid
cfg      // 统一缩写，不要某处用 config 某处用 cfg 某处用 conf

// ❌ 同一概念不同命名散落在各处
// file_a.go: userID
// file_b.go: userId
// file_c.go: uid
// file_d.go: user_id
```

**常见的必须统一的命名约定**：

| 概念 | 选定一种后全项目统一 | 禁止混用 |
|------|---------------------|---------|
| 上下文 | `ctx` | `c` / `context` / `reqCtx` |
| 错误 | `err` | `e` / `error` / `er` |
| HTTP 请求 | `req` 或 `r` | 选一种后不混用 |
| HTTP 响应 | `resp` 或 `w` | 选一种后不混用 |
| 数据库连接 | `db` | `conn` / `database` / `dbc`（除非语义不同）|
| 日志对象 | `logger` 或 `log` | 选一种后不混用 |
| 配置对象 | `cfg` 或 `config` | 选一种后不混用 |
| 互斥锁 | `mu` | `lock` / `mtx` / `mutex` |

**变量声明风格统一**：

```go
// ✅ 选定一种声明风格后全项目统一
// 风格 A：短变量声明
client := &http.Client{Timeout: 10 * time.Second}

// 风格 B：var 声明
var client = &http.Client{Timeout: 10 * time.Second}

// ❌ 同一场景下混用两种风格
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

### 4.5 函数调用与操作写法全项目统一

> 相同功能的函数调用在全项目中必须使用统一的写法和模式。
> 一旦项目中确立了某种写法，全项目必须统一遵循，不允许多种写法并存。

**错误创建与包装方式统一**：

```go
// ✅ 全项目统一使用一种错误创建方式
// 方式 A：fmt.Errorf + %w
return fmt.Errorf("query user: %w", err)

// 方式 B：errors.Wrap（如使用 pkg/errors）
return errors.Wrap(err, "query user")

// ❌ 项目中混用多种错误包装
// file_a.go: fmt.Errorf("failed: %w", err)
// file_b.go: errors.Wrap(err, "failed")
```

**错误消息格式统一**：

```go
// ✅ 选定一种错误消息格式后全项目统一
// 风格 A：动词短语（推荐）
return fmt.Errorf("query user by id: %w", err)
return fmt.Errorf("parse config file: %w", err)

// 风格 B：failed to 句式
return fmt.Errorf("failed to query user: %w", err)

// ❌ 错误消息格式不统一
// file_a.go: "query user failed: ..."     （后置 failed）
// file_b.go: "failed to query user: ..."  （前置 failed to）
// file_c.go: "query user: ..."            （无 failed）
// file_d.go: "Query user error: ..."      （首字母大写 + error）
```

**构造函数与初始化写法统一**：

```go
// ✅ 全项目统一选择一种构造方式
// 方式 A：函数式选项
srv := NewServer(WithPort(8080), WithTimeout(30*time.Second))

// 方式 B：配置结构体
srv := NewServer(ServerConfig{Port: 8080, Timeout: 30*time.Second})

// ❌ 同一项目中部分用选项模式，部分用配置结构体
```

**HTTP 响应写法统一**：

```go
// ✅ 全项目统一 HTTP 响应写法
// 方式 A：
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(resp)

// 方式 B：
data, _ := json.Marshal(resp)
w.Write(data)

// 选定一种后全项目统一，不要 A、B 混用
```

**类型转换写法统一**：

```go
// ✅ 全项目统一选择一种类型转换方式
s := strconv.Itoa(n)

// ❌ 混用 fmt.Sprintf 和 strconv
// file_a.go: s := strconv.Itoa(n)
// file_b.go: s := fmt.Sprintf("%d", n)
```

**通用一致性检查原则**：

1. **同一种操作只允许一种写法**：一旦出现两种写法做同一件事，标记为风格不一致
2. **先到先得**：以项目中最先/最多采用的写法为基准，偏离基准的应统一修正
3. **新代码跟随已有风格**：新增代码必须遵循项目已确立的风格
4. **重构时全量统一**：如需变更风格，必须全项目统一变更，不允许新旧风格并存

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

### 9.5 日志全项目一致性

> 日志是全项目代码风格一致性最容易失控的领域。以下规则强制统一日志写法。

**日志库与调用方式统一**：

```go
// ✅ 全项目统一使用同一种日志调用风格

// 风格 A：zap 的强类型 Field
logger.Info("user login",
    zap.String("user_id", userID),
    zap.String("ip", clientIP),
    zap.Int("status", statusCode),
)

// 风格 B：slog 的键值对
slog.Info("user login",
    "user_id", userID,
    "ip", clientIP,
    "status", statusCode,
)

// 风格 C：logrus 的 WithFields
log.WithFields(log.Fields{
    "user_id": userID,
    "ip":      clientIP,
    "status":  statusCode,
}).Info("user login")

// ❌ 同一项目中混用多种日志库或调用风格
```

**日志消息风格统一**：

```go
// ✅ 选定一种日志消息风格后全项目统一
// 风格 A：小写开头，无标点（推荐）
logger.Info("user login success")
logger.Error("query database failed")

// 风格 B：大写开头，无标点
logger.Info("User login success")

// ❌ 消息风格不统一
// logger.Info("User login success.")   // 有句号
// logger.Info("user login success")    // 小写无句号
// logger.Info("USER_LOGIN")            // 全大写下划线
```

**日志字段（Field）命名统一**：

同一语义的日志字段在全项目中必须使用同一个 key 名。

```go
// ✅ 全项目统一字段命名（选定一套命名规则）
// 规则：全部使用 snake_case
zap.String("user_id", userID)
zap.String("request_id", reqID)
zap.String("trace_id", traceID)
zap.Int("status_code", code)
zap.String("method", method)
zap.String("path", path)
zap.Duration("latency", dur)
zap.Error(err)  // error 字段统一用 zap.Error

// ❌ 同一字段不同 key 名散落在各处
// file_a.go: zap.String("userId", id)       // camelCase
// file_b.go: zap.String("user_id", id)      // snake_case
// file_c.go: zap.String("UserID", id)       // PascalCase
// file_d.go: zap.String("user-id", id)      // kebab-case
```

**建议统一的常见日志字段名**：

| 语义 | 推荐 key（snake_case） | 禁止混用 |
|------|----------------------|---------|
| 用户标识 | `user_id` | `userId` / `UserID` / `uid` |
| 请求标识 | `request_id` | `requestId` / `reqId` / `req_id` |
| 链路追踪 | `trace_id` | `traceId` / `TraceID` |
| HTTP 方法 | `method` | `http_method` / `httpMethod` |
| 请求路径 | `path` | `url` / `uri` / `endpoint`（除非语义不同）|
| 状态码 | `status_code` | `statusCode` / `status` / `code` |
| 耗时 | `latency` 或 `duration` | 选一种后不混用 |
| 错误信息 | `error`（使用日志库内置 Error 方法）| `err` / `errmsg` / `error_msg` |
| 调用方 | `caller` | `source` / `from`（除非语义不同）|
| 模块/组件 | `component` | `module` / `service`（除非语义不同）|

**同类操作日志必备字段统一**：

对于同一类操作，日志中携带的上下文字段必须保持一致。

```go
// ✅ 全项目统一：所有 HTTP 请求日志必须包含以下字段
logger.Info("request completed",
    zap.String("request_id", reqID),
    zap.String("method", r.Method),
    zap.String("path", r.URL.Path),
    zap.Int("status_code", statusCode),
    zap.Duration("latency", duration),
)

// ✅ 全项目统一：所有数据库操作日志必须包含以下字段
logger.Debug("db query",
    zap.String("request_id", reqID),
    zap.String("query", queryName),
    zap.Duration("latency", duration),
)

// ✅ 全项目统一：所有错误日志必须包含以下字段
logger.Error("query user failed",
    zap.String("request_id", reqID),
    zap.String("user_id", userID),
    zap.Error(err),
)

// ❌ 同类操作日志字段不一致
// handler_a.go: 打印了 request_id, method, path, latency
// handler_b.go: 只打印了 path（缺少 request_id 和 latency）
// handler_c.go: 打印了 request_id, url, time_cost（字段名不同）
```

---

## 十、格式字符串

Printf 风格的格式字符串声明为 `const`，以便 `go vet` 进行静态分析。

```go
// ✅ 可被 go vet 检查
const fmtStr = "invalid input: %s"
return fmt.Errorf(fmtStr, input)
```

Printf 风格函数名以 `f` 结尾（如 `Wrapf`、`Errorf`），以便工具检测格式字符串匹配。
