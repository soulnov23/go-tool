# Go 代码审查完整检查清单

> 参考标准：Go 官方 Code Review Comments、Google Go Style Guide、Uber Go Style Guide、Effective Go、OWASP Go-SCP
>
> 审查时先识别项目主流风格，以此为基准逐项检查。[严重] 标记为必须修复项。

---

## 一、命名一致性

### 检查项
- 文件名：小写 + 下划线 (`snake_case.go`)，测试文件 `_test.go` 后缀
- 包名：全小写、简短、单词不拆分，无下划线，避免 `common`/`util`/`misc`/`base`/`helper`/`api` 等模糊名称
- 包名与目录名一致，不使用复数形式
- 避免选择容易被局部变量遮蔽的包名（如 `count` → `usercount`）
- 导出标识符：大驼峰 `PascalCase`
- 未导出标识符：小驼峰 `camelCase`（非 `snake_case`，非 `ALL_CAPS`）
- 错误变量：导出用 `Err` 前缀 (`ErrNotFound`)，未导出用 `err` 前缀；错误类型用 `Error` 后缀
- 构造函数：`New` / `NewXxx` 前缀，不混用 `Create`/`Make`/`Build`
- Option 函数：`With` 前缀 (`WithTimeout`)
- 接口：单方法用 `-er` 后缀 (`Reader`)；多方法用描述性名词；避免 `Manager`/`Helper`
- 缩写词：全大写或全小写一致 (`ID` 非 `Id`，`HTTP` 非 `Http`，`URL` 非 `Url`)
- 多缩写词组合时每个缩写词内部一致（如 `XMLAPI` 或 `xmlAPI`）
- Getter 不加 `Get` 前缀 (`Name()` 优于 `GetName()`)；复杂计算用 `Compute`/`Fetch`
- 布尔函数：`Is`/`Has`/`Can` 前缀
- 接收者名称：1-2 字母缩写，不用 `this`/`self`/`me`，同一类型所有方法保持一致
- 变量名长度与作用域成正比：小作用域短名，大作用域描述性名
- 避免在变量名中重复类型信息（`users` 优于 `userSlice`）
- 避免在导出符号中重复包名（`widget.New` 优于 `widget.NewWidget`）
- 常量使用 MixedCaps（驼峰），不用 `ALL_CAPS` 或 `kConstName`
- 测试函数遵循 `TestXxx`/`BenchmarkXxx`/`ExampleXxx` 命名

### 常见问题
- 混用 `snake_case` 和 `camelCase`
- 同一概念在不同包中名称不同
- [严重] 包名与目录名不一致
- 接收者名称在同类型不同方法间不一致

---

## 二、代码风格与可读性

### 检查项

**格式化**：
- 代码通过 `gofmt` / `goimports` 格式化
- 避免过长的行（建议软限制 99 字符），基于语义换行而非强制截断
- `if` 语句条件不要换行

**导入分组**：
- 分组排序：标准库 → 第三方库 → 项目内部包（→ protobuf 包 → 副作用导入）
- 各组之间用空行隔开
- 避免不必要的导入重命名，仅在名称冲突时使用
- [严重] 禁止使用 `import .` 形式（除测试循环依赖例外）
- 副作用导入 (`import _`) 仅出现在 main 包或测试中

**缩进与错误流**：
- [严重] 先处理错误并返回，正常逻辑保持最小缩进
- 不使用 `else` 包裹正常代码路径
- 减少嵌套层级，提前 return/continue

**声明风格**：
- 显式赋值用短变量声明 (`:=`)
- 声明空切片用 `var t []string`（nil 切片），而非 `t := []string{}`
- `nil` 是有效的空切片，检查用 `len(s) == 0` 而非 `s == nil`
- 顶级变量声明使用 `var`，类型已明确时可省略类型
- 未导出的全局变量加 `_` 前缀防止误用（错误变量用 `err` 前缀除外）
- 结构体初始化使用字段名，省略零值字段
- 使用 `&T{}` 而非 `new(T)` 创建指针
- 常量/变量使用 `var ()` 和 `const ()` 分组声明
- 空 map 用 `make()`，固定元素用字面量 `{}`
- 枚举使用 `iota`，通常从非零值开始（`iota + 1`）除非零值是期望的默认值

**注释与文档**：
- [严重] 所有导出的类型、函数、常量必须有 godoc 注释
- 注释为完整句子，以被描述标识符名称开头，以句号结尾
- 行为不明显的未导出类型/函数也应有注释
- 包注释紧邻 `package` 子句上方，无空行
- 复杂逻辑有行内注释解释意图
- 避免过时注释与代码不一致
- 函数调用中含义不明显的参数（特别是 bool）添加注释 `/* paramName */`
- 使用原始字符串字面量（反引号）避免转义

**命名结果参数与裸返回**：
- 命名结果参数仅在返回多个同类型参数或文档需要时使用
- 不要仅为省略 `var` 声明而命名结果参数
- 裸返回仅用于极短函数，中等及以上函数明确返回值

**代码组织**：
- 函数按大致调用顺序排列，按接收者分组
- 类似声明（import/const/var/type）分组，逻辑分组间空行
- Printf 风格函数名以 `f` 结尾（如 `Wrapf`）
- 格式字符串声明为 `const` 以便 `go vet` 检查

### 常见问题
- `if ... { return } else { ... }` 应简化为 `if ... { return }` + 正常代码
- 使用 `fmt.Errorf("Something bad.")` — 错误字符串不应首字母大写或以标点结尾
- [严重] 导出函数缺少 godoc 注释

---

## 三、错误处理

### 检查项

**错误创建**：
- 无需匹配的静态错误：`errors.New("...")`
- 需要匹配的静态错误：导出的 `var ErrXxx = errors.New("...")`
- 无需匹配的动态错误：`fmt.Errorf("...: %v", err)`
- 需要匹配的动态错误：自定义错误类型 + `Error` 后缀

**错误包装**：
- [严重] 需保留错误链用 `%w`，不需要时用 `%v`
- 错误消息添加有意义的上下文（操作名、关键参数）
- 合理使用 `errors.Is()` / `errors.As()` 判断错误，不用字符串比较

**错误字符串格式**（Go 官方标准）：
- [严重] 错误字符串不应首字母大写（除专有名词/缩写）
- [严重] 错误字符串不应以标点符号结尾
- 原因：错误信息常被包装在其他上下文中打印

**错误处理原则**：
- [严重] 不要用 `_` 丢弃错误，除非有注释解释为何安全
- [严重] 错误一次处理：不要既记录日志又返回错误（导致日志噪音）
- 初始化失败 panic，运行时错误 return error
- 不要使用返回值（-1, null, ""）表示错误，用额外的 `error` 或 `bool` 返回值
- `os.Exit` / `log.Fatal` 仅在 `main()` 中调用，其他函数返回 error

**Must 函数**：
- 仅用于程序启动初始化或测试
- 命名约定 `MustXxx`

### 常见问题
- 错误包装用 `%v` 丢失错误链
- 记录日志后又返回同一个错误
- [严重] 用 `_` 忽略关键函数的错误返回值

---

## 四、资源与生命周期管理

### 检查项

**defer 与资源释放**：
- [严重] 文件/连接/事务等获取成功后立即 `defer Close()`
- 错误分支正确清理已获取的资源
- 循环中创建的资源在循环内清理（不要在循环中 defer）
- 注意 `defer f.Close()` 的错误被忽略问题
- `defer` 中的变量在声明时求值（非执行时）

```go
// ✅ 正确的资源管理模式
f, err := os.Open(name)
if err != nil {
    return err
}
defer f.Close()
```

**context 使用**：
- [严重] 作为函数第一个参数 `func Foo(ctx context.Context, ...)`
- 不存储在 struct 中，每次调用时传递
- 不创建自定义 Context 类型
- 不传 nil，用 `context.TODO()` 或 `context.Background()`
- 不用 `context.WithValue` 传业务数据
- 使用 `context.WithTimeout` / `context.WithCancel` 控制超时和取消

**goroutine 生命周期**（Go 官方重点）：
- [严重] 必须明确每个 goroutine 何时/是否退出
- 每个 goroutine 有退出机制（context cancel / channel close / done signal）
- 即使阻塞的 channel 不可达，GC 也不会终止 goroutine — 会导致泄漏
- 优先使用同步函数，让调用者决定是否启动 goroutine
- 保持并发代码简单，使生命周期显而易见

**channel 安全**：
- [严重] 避免 send on closed channel（panic）
- 明确谁负责关闭 channel（通常由发送方关闭）
- channel 通常大小为 0（无缓冲）或 1，其他大小需严格审查（Uber 标准）
- 无缓冲 channel 确保有消费者，避免 deadlock

### 常见问题
- [严重] 忘记 `defer Close()` 导致资源泄漏
- goroutine 无退出机制导致泄漏
- `context.TODO()` 遗留在生产代码中
- 循环中 `defer` 导致文件句柄耗尽

---

## 五、逻辑正确性与 Bug 检测

### 检查项

**边界条件**：
- [严重] nil 检查：nil map 写入 panic、nil 指针解引用、nil slice 取下标
- 零值、空切片、空字符串、空 map 处理
- 整数溢出/下溢（特别是无符号整数减法）、除零
- 接口值可能为 nil（特别注意：非 nil 接口可能包含 nil 指针）

**切片与 Map 陷阱**：
- append 后可能影响原切片（共享底层数组）
- 子切片持有大数组引用导致内存无法释放
- [严重] 在边界处复制切片和 map（Uber 标准）：接收参数时复制防止外部修改，返回时复制防止暴露内部状态
- 向 nil map 写入 panic

**类型断言**：
- [严重] 使用 `value, ok := x.(Type)` 安全断言，不用单返回值形式（会 panic）

**defer 陷阱**：
- 循环中 `defer` 导致资源延迟释放
- defer 函数参数在声明时求值

**复制陷阱**：
- [严重] 不要复制包含 `sync.Mutex`、`sync.RWMutex`、`bytes.Buffer` 的结构体
- 如果类型 T 的方法关联指针 `*T`，不要复制 T 的值

**数据竞争**：
- [严重] 共享变量有正确同步保护
- `sync.WaitGroup.Add` 在 goroutine 启动前调用
- `sync.Pool.Get` 返回值必须重置旧状态

**时间处理**（Uber 标准）：
- 使用 `time.Time` 处理时间瞬间，`time.Duration` 处理时间段
- 不假设一天 24 小时、一年 365 天等
- 外部系统交互中字段名包含单位（如 `IntervalMillis`）

### 常见问题
- [严重] 向 nil map 写入导致 panic
- channel close 时机不当导致 panic
- 复制含 Mutex 的结构体导致未定义行为
- 未使用 `go test -race` 发现 data race

---

## 六、并发安全与死锁

### 检查项

**锁竞争**：
- 热路径可否用 `atomic` 替代 `Mutex`
- 读多写少用 `sync.RWMutex`
- [严重] 锁范围最小化，不在锁内执行 I/O / 网络请求
- 考虑分片减少竞争
- `sync.Mutex` 零值有效，不需要指针，不要嵌入导出结构体（Uber 标准）

**死锁风险**：
- [严重] 锁嵌套顺序一致（避免 A→B 与 B→A）
- channel 操作有超时或 default 分支
- 数据库事务避免互相等待

**并发原语选择**：
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

**goroutine 控制**：
- [严重] 控制 goroutine 创建数量（worker pool / semaphore）
- 高并发有限流/背压机制
- 不要 fire-and-forget goroutine，必须有等待退出机制

**伪共享**：高频 `atomic` 字段间有缓存行填充

**CAS 自旋**：失败后 `runtime.Gosched()` 让出 CPU，有退出条件

### 常见问题
- [严重] 在锁内执行网络/数据库请求导致锁持有过长
- [严重] goroutine 无限制创建导致 OOM
- 未使用 `go test -race` 导致 data race 上线

---

## 七、内存分配与 GC 压力

### 检查项

**避免不必要的堆分配**：
- 返回局部变量指针导致逃逸
- `any`/`interface{}` 参数导致值装箱逃逸
- 小结构体（≤4 字段）考虑传值而非传指针

**sync.Pool**：
- 高频创建/销毁的临时对象使用 `sync.Pool`
- 归还前正确重置状态

**预分配**：
- [严重] 已知大小的切片 `make([]T, 0, n)` 预分配
- Map `make(map[K]V, n)` 预分配
- `strings.Builder` 使用 `Grow()` 预分配

**字符串操作**：
- 多字符串拼接使用 `strings.Builder`
- 热路径用 `strconv` 替代 `fmt.Sprintf`
- 减少 `[]byte` ↔ `string` 不必要转换
- 不在循环中反复从固定字符串创建字节切片

**闭包**：避免捕获不必要的大对象

**热路径优化**：
- 用预定义哨兵错误替代 `fmt.Errorf`
- 避免热路径中频繁创建临时对象

### 常见问题
- 热路径 `fmt.Errorf`/`fmt.Sprintf` 创建大量临时对象
- 未预分配切片频繁 `append` 导致多次扩容

---

## 八、运行时性能

### 检查项

**值传递 vs 指针传递**（Go 官方标准）：
- 不要仅为节省字节传递指针；如果函数只用 `*x`，参数不应是指针
- 大结构体或会增长的小结构体传指针
- map、func、chan 不要传指针

**接收者类型选择**：
- 需要修改接收者 → 指针接收者
- 包含 `sync.Mutex` 等不可复制字段 → 指针接收者
- 小型不可变结构体或基本类型 → 值接收者
- 不确定时 → 指针接收者
- 同一类型不要混用值接收者和指针接收者

**热路径优化**：
- 优先 `strconv` 而非 `fmt` 进行类型转换
- 使用 `%q` 打印带引号字符串
- 新代码用 `any` 替代 `interface{}`

**同步 vs 异步**：
- 优先使用同步函数，让调用者决定是否并发
- 同步函数更易推理、测试、调用

### 常见问题
- 小结构体不必要地传指针
- 接收者类型在同一类型不同方法间不一致

---

## 九、安全编码

### 检查项

**输入验证**：
- [严重] 所有外部输入（HTTP 请求参数、用户输入、文件内容）在使用前验证
- 验证数据类型、长度、范围、格式
- 白名单验证优于黑名单

**SQL 注入防护**：
- [严重] 使用参数化查询 / 预编译语句，禁止字符串拼接 SQL
- ORM 查询也需审查动态拼接部分

**加密与随机数**（Go 官方标准）：
- [严重] 生成密钥/token 必须使用 `crypto/rand`，禁止 `math/rand`
- 使用标准库加密实现，不自行实现加密算法
- 密码哈希使用 `bcrypt` / `argon2`，不用 MD5/SHA

**敏感信息保护**：
- [严重] 日志中不打印密码、Token、身份证号、银行卡号等敏感数据
- 配置中的密钥不硬编码在源码中
- 错误信息不暴露内部实现细节给外部用户

**依赖安全**：
- 定期运行 `govulncheck` 扫描依赖漏洞
- 保持 Go 版本和依赖为最新（注意审查更新）
- 使用 `go.sum` 确保依赖完整性

**文件路径安全**：
- [严重] 防止路径遍历攻击（`../` 注入）
- 使用 `filepath.Clean` 和白名单验证路径

**HTTP 安全**：
- 设置合理的请求超时（读/写/空闲超时）
- 限制请求体大小
- 使用 HTTPS
- 设置安全相关 HTTP 头

### 常见问题
- [严重] 使用 `math/rand` 生成安全相关随机数
- SQL 拼接导致注入风险
- 敏感信息泄露到日志或错误响应中

---

## 十、类型系统与接口设计

### 检查项

**接口定义**（Go 官方核心原则）：
- [严重] 接口由使用方（消费者）定义，而非实现方
- 实现包返回具体类型（指针或结构体），不要为 mock 在实现方定义接口
- 函数接受接口参数，返回具体类型
- 避免过早创建接口；只有一个实现时通常不需要接口

**接口大小**：
- 接口尽量小，遵循接口隔离原则
- 单方法接口最灵活、最易组合

**编译时接口验证**（Uber 标准）：
- 导出类型实现特定接口时，使用 `var _ Interface = (*Type)(nil)` 编译时断言

**结构体嵌入**：
- 嵌入类型放在字段列表顶部，空行隔开
- 提供切实的语义收益，不破坏零值
- [严重] 互斥锁不嵌入，作为非导出字段
- 不要通过嵌入暴露不需要的方法

**指针与接口**：
- 几乎永远不要使用指向接口的指针
- 接口作为值传递，底层数据可以是指针

**泛型使用**：
- 允许使用但不滥用
- 如果只有一种类型实例化，先不用泛型
- 优先使用接口而非泛型实现多态

**Option 模式**：
- 可选配置遵循 `Options` 结构体 + `Option` 函数类型 + `WithXxx` 设置函数
- 默认值合理设置

```go
// ✅ 函数式选项模式
type Option func(*Server)

func WithPort(port int) Option {
    return func(s *Server) { s.port = port }
}

func NewServer(opts ...Option) *Server {
    s := &Server{port: 8080}
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

**序列化标签**（Uber 标准）：
- 任何序列化的结构体字段使用相关标签（`json:"field_name"`）

### 常见问题
- 在实现方定义接口以便 mock — 应在使用方定义
- 嵌入 `sync.Mutex` 导致零值不安全
- 接口过度设计，只有一个实现

---

## 十一、测试质量

### 检查项

**表驱动测试**：
- 使用表驱动测试减少重复代码
- 测试用例结构体切片命名 `tests`，用例命名 `tt`
- 字段用 `give`/`want` 或 `input`/`expected` 前缀
- 避免在子测试中包含复杂分支逻辑

```go
// ✅ 表驱动测试模板
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {name: "positive", a: 1, b: 2, want: 3},
        {name: "zero", a: 0, b: 0, want: 0},
        {name: "negative", a: -1, b: 1, want: 0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

**测试失败信息**（Go 官方标准）：
- [严重] 失败信息包含：函数名、输入、实际结果、期望结果
- 格式：`t.Errorf("Foo(%q) = %d; want %d", input, got, want)`
- 顺序：先 `got` 后 `want`
- 测试失败后尽可能继续（用 `t.Error`），仅后续无意义时用 `t.Fatal`

**子测试**：
- 使用 `t.Run()` 组织子测试
- 子测试名应可读、可输入，避免使用 `/`
- 子测试不应依赖其他测试的执行状态

**结构比较**：
- 使用 `cmp.Equal` / `cmp.Diff` 进行深度比较，避免手写字段比较
- 比较语义结果而非不稳定的输出（如 JSON 字符串序列化）

**边界与错误路径覆盖**：
- 核心逻辑有单元测试
- 边界条件和错误路径有测试覆盖
- 不只覆盖 happy path

**测试注意事项**：
- 不在单独的 goroutine 中调用 `t.Fatal`（不会终止测试）
- `t.Fatal` 仅用于测试设置失败或必须终止测试的情况
- 测试辅助函数使用 `t.Helper()` 标记
- 优先使用真实的测试服务器（httptest）而非手写 mock

**基准测试**：
- 性能关键路径有 `Benchmark` 测试
- 使用 `-benchmem` 验证内存分配

**模糊测试**：
- 解析/反序列化等输入处理使用 Go 原生模糊测试 (`testing.F`)
- 可发现 SQL 注入、缓冲区溢出、DoS 等边缘漏洞

**竞态检测**：
- [严重] 并发代码使用 `go test -race` 检测
- CI/CD 中集成 race detector

**示例函数**：
- 新包提供可运行的 `Example` 函数演示用法

### 常见问题
- 测试失败信息仅 `t.Fail()` 无具体信息，排查困难
- 只测试 happy path，缺少错误路径
- [严重] 未使用 `-race` 运行测试

---

## 十二、项目工程规范

### 检查项

**包设计与职责**：
- 每个包职责单一
- 避免循环依赖
- 内部实现使用 `internal` 包保护
- 导出 API 保持稳定，变更需谨慎

**init() 使用**：
- [严重] 避免 `init()` 中执行 I/O、网络调用、复杂逻辑
- `init()` 仅用于简单注册（如 database driver、codec）
- 保证 `init()` 行为完全确定性
- `init()` 中不启动 goroutine

**全局变量**：
- [严重] 避免可变全局变量，使用依赖注入
- 全局只读配置可以接受，但可变状态应通过参数传递
- flag 变量仅在 `main` 包中定义

**依赖管理**：
- `go.mod` 中 Go 版本与实际使用一致
- 定期运行 `go mod tidy` 清理未使用依赖
- 依赖版本锁定，`go.sum` 纳入版本控制
- 定期运行 `govulncheck` 扫描依赖漏洞

**构建与工具**：
- 使用 `go vet ./...` 检查可疑构造
- 使用 `golangci-lint` 进行综合静态分析
- 逃逸分析：`go build -gcflags="-m" ./...`
- 竞态检测：`go test -race ./...`
- 漏洞扫描：`govulncheck ./...`

**日志规范**：
- [严重] 项目内统一使用同一日志库，不混用
- 结构化日志：使用 Field 方式而非 `fmt.Sprintf` 拼接
- 日志消息简短明确，风格全项目统一
- 错误日志包含足够上下文（请求ID、用户ID、关键参数）
- 日志级别使用恰当：Debug/Info/Warn/Error/Fatal
- 热路径/循环中避免无限制日志输出
- panic 恢复：关键 goroutine 有 `recover()` 并记录堆栈

**API 设计**：
- 函数参数 > 5 个考虑 Options 模式
- 遵循 `(result, error)` 返回值惯例
- 避免裸 `bool` 返回值，考虑自定义类型或命名返回值

### 常见问题
- `init()` 中执行网络请求导致启动失败难以定位
- [严重] 可变全局变量导致并发问题和测试不稳定
- 循环依赖导致编译错误
- 日志混用多个库，风格不统一
