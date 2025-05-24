# 测试策略

本项目后端采用 Go 语言内置的 `testing` 包结合 `net/http/httptest` 进行 API 集成测试。

## 自动化 API 测试

-   **位置**: 测试代码位于被测试包的 `_test.go` 文件中，例如 `internal/router/router_test.go` 用于测试路由处理。
-   **框架**: 主要使用 `testing`、`net/http/httptest` 和 `github.com/stretchr/testify/assert` 断言库。
    -   对于**接口 mock**，项目统一使用 `go.uber.org/mock/gomock`。`gomock` 通过代码生成提供类型安全的 mock 对象，有助于编写清晰且可维护的单元测试和集成测试。通常结合 `//go:generate mockgen ...` 指令来生成 mock 实现。
    -   对于**数据库交互的 mock**，在需要精确控制 SQL 行为的单元测试中，会使用 `github.com/DATA-DOG/go-sqlmock` (`sqlmock`)。
-   **数据库测试方法**:
    -   **内存 SQLite 数据库**: 主要用于 API 集成测试和涉及 GORM 交互的单元测试。测试时使用 (`gorm.io/driver/sqlite` 配置为 `file::memory:?cache=shared`)，以确保测试的独立性和速度，并模拟真实数据库环境下的行为。
        -   在每个测试函数或测试设置 (`setupTestRouter` 辅助函数) 中，当使用内存 SQLite 时，会使用 GORM 的 `AutoMigrate` 功能自动创建所需的表结构。
        -   测试数据通过辅助函数（如 `createTestUser`）在测试开始前按需创建。
        -   测试设置函数 (例如 `setupTestRouter`) 负责初始化 Gin 引擎和所有必要的依赖项（Handler, Service, Repository），并将它们连接到测试数据库。
    -   **SQL Mocking (`sqlmock`)**: 对于需要精确控制和验证底层 SQL 语句的 Repository 层单元测试，推荐使用 `sqlmock` (如前文"框架"部分所述)。这允许开发者模拟数据库驱动的行为，而无需实际数据库连接。
-   **执行**:
    -   在包目录下（例如 `backend/internal/router`）运行 `go test` 或 `go test -v` 来执行测试。
    -   测试应该被集成到 CI/CD 流程中，以确保代码变更不会破坏现有功能。

## 手动 API 测试

-   可以使用任何 REST 客户端工具（如 Postman, Insomnia, Hoppscotch 或 `curl`）进行手动测试。
-   详细的 API 端点定义和 `curl` 使用示例可以在各自的组件设计文档中找到，例如 `docs/design/components/auth_api.md`。
-   手动测试时，请确保后端服务正在运行，并连接到正确的开发数据库 (`data/effiplat.db`)。
