# 后端服务设计

## 1. 技术选型

*   **语言**: 当前采用Go来搭建服务框架，主要提供web,数据库，配置相关服务，[待定，后期有高性能模块考虑用c++实现]
*   **Web 框架**: Gin
*   **ORM**: GORM/sqlite3
*   **依赖管理**: Go Modules
*   **日志库**: **`uber-go/zap`**
*   **配置管理**: **Viper**
*   **测试框架**: Go 标准库 `testing` + Testify (可选，提供 Assertions)。
*   **高性能模块 (未来考虑)**: 特定性能瓶颈可能考虑 C++ 通过 IPC/gRPC 集成。
  
## 2. 内部架构/分层

推荐采用分层架构，后续会向多进程框架演进，如：

*   **Handler/Controller 层**: 接收 HTTP 请求，解析参数，调用 Service 层，组装响应。
*   **Service 层**: 实现核心业务逻辑，处理数据校验、事务管理、调用 Repository。
*   **Repository/DAL 层**: 负责与数据库交互，封装数据访问逻辑。
*   **Model 层**: 定义数据结构 (对应数据库表)。

## 2.1 依赖注入与服务初始化方案（wire）

为提升后端服务的可维护性、类型安全和依赖管理效率，项目采用 [Google wire](https://github.com/google/wire) 作为依赖注入（DI）工具。  
wire 方案适用于所有 HTTP 接口、业务服务、数据库访问、第三方集成、定时任务等后端模块。

### 方案说明

- **依赖注入方式**：使用 wire 进行编译期依赖注入，自动生成依赖关系初始化代码，避免手写样板代码。
- **注入范围**：logger、db、各 repository、service、handler 及其他全局依赖均通过 wire 统一管理。
- **优势**：
  - 类型安全，依赖关系变更可编译期发现问题
  - 依赖链清晰，main.go 结构简洁
  - 无运行时性能损耗
  - 便于后续扩展和维护
- **后续扩展**：如项目规模进一步扩大，可平滑迁移到 fx 等更复杂的依赖注入框架。

### wire 配置示例

```go
// wire.go
func InitializeApp() (*App, error) {
    wire.Build(
        NewLogger,
        NewDB,
        NewUserRepo, NewUserService, NewUserHandler,
        // ...其他表/模块的 provider
        NewApp,
    )
    return &App{}, nil
}
```

### 适用范围

- 所有 HTTP handler、service、repository、定时任务、第三方服务集成等后端模块，均通过 wire 统一注入依赖。
- 未来如需更复杂的生命周期管理，可考虑平滑迁移到 fx。

## 3. 核心模块设计 (V1.0)

*   **用户认证与授权 (AuthN & AuthZ)**:
    *   实现登录接口，校验用户名密码，生成 JWT。
    *   设计中间件，拦截请求，校验 JWT，并将用户信息注入上下文。
    *   实现 RBAC 检查逻辑，根据用户角色和请求资源/操作进行权限判断。
    *   详细 API 设计见 [api_design.md](./api_design.md)
*   **环境管理 (Environment)**:
    *   实现 CRUD 操作对应的 Service 和 Repository 逻辑。
    *   处理环境与服务实例、资产、职责组的关联关系。
*   **Bug 管理 (Bug)**:
    *   实现 Bug 提交逻辑，注意关联环境、服务、业务等信息。
    *   实现 Bug 状态流转逻辑。
    *   实现 Bug 分配逻辑 (V1.0 可能先手动分配，V1.1 考虑基于规则自动分配)。
*   **审计日志 (AuditLog)**:
    *   设计 AOP (面向切面编程) 方式或在 Service 层关键操作后显式记录审计日志。
    *   异步写入日志以减少对主流程性能影响。

## 4. 配置管理

*   使用 `viper` 库管理服务配置（数据库连接、端口、日志级别等）。
*   使用YAML，良好地支持数组结构，对于复杂嵌套结构的可读性好。
*   同时支持通过环境变量覆盖配置文件中的值。
*   考虑使用配置中心（如 Nacos, Consul）进行动态配置管理（未来）。

## 5. 日志记录

*   **日志库**: 选用 **`uber-go/zap`**。
    *   *理由*: 极致的性能和低内存分配，非常适合高并发场景。原生支持结构化日志和**采样 (Sampling)** 功能。
*   **结构化**: `zap` 强制或强烈推荐使用强类型的字段进行结构化日志记录。
*   **包含信息**: 日志条目应包含以下关键信息：
    *   **时间戳 (Timestamp)**: `zap` 自动添加。
    *   **日志级别 (Level)**: `zap` 自动添加。
    *   **消息内容 (Message)**: `zap` 自动添加。
    *   **代码位置 (Caller)**: 可通过配置开启。
    *   **进程号 (PID)**: 可在 Logger 初始化时配置添加为固定字段。
    *   **请求/追踪 ID (Request/Trace ID)**: 应在请求处理链路上传递，并通过 `Logger.With` 添加到日志上下文中。
    *   **结构化属性 (Fields)**: 使用 `zap.String()`, `zap.Int()`, `zap.Any()` 等方法添加键值对。
*   **输出与轮转**: 
    *   日志输出到文件。
    *   集成 **`natefinch/lumberjack.v2`** 库实现日志文件的**按大小轮转 (Rotation)** 和**自动压缩 (Compression)** (通过 `zapcore.AddSync(logWriter)` 集成)。
    *   开发环境下可同时输出到控制台 (使用 `zapcore.NewTee`)。
*   **动态级别调整**: 使用 **`zap.AtomicLevel`** 实现运行时的日志级别动态调整。
*   **采样**: 利用 `zap` 内建的**采样器 (Sampler)** 配置，在高频日志点减少日志量。
*   **集中化 (未来)**: 考虑将日志推送到日志平台 (如 ELK, Loki)。

## 6. 错误处理

*   **基本原则**: 错误处理应保持一致性、可预测性，并向 API 消费者提供清晰有效的信息，同时保护内部实现细节。
*   **错误包装 (Error Wrapping)**: 遵循 Go 1.13+ 规范，使用 `fmt.Errorf` 的 `%w` 动词在错误向上传递时添加上下文，保留原始错误信息。例如: `fmt.Errorf("service: failed to process order %d: %w", orderID, err)`。
*   **自定义错误类型 (Custom Error Types)**: 定义特定的错误类型（实现 `