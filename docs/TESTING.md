# 测试策略

本项目后端采用 Go 语言内置的 `testing` 包结合 `net/http/httptest` 进行 API 集成测试。

## 自动化 API 测试

-   **位置**: 测试代码位于被测试包的 `_test.go` 文件中，例如 `internal/router/router_test.go` 用于测试路由处理。
-   **框架**: 主要使用 `testing`、`net/http/httptest` 和 `github.com/stretchr/testify/assert` 断言库。
-   **数据库**:
    -   测试时使用**内存中的 SQLite 数据库** (`gorm.io/driver/sqlite` 配置为 `file::memory:?cache=shared`)，以确保测试的独立性和速度，并避免影响开发数据库。
    -   在每个测试函数或测试设置 (`setupTestRouter` 辅助函数) 中，使用 GORM 的 `AutoMigrate` 功能自动创建所需的表结构。
    -   测试数据通过辅助函数（如 `createTestUser`）在测试开始前按需创建。
-   **依赖**: 测试设置函数 (`setupTestRouter`) 负责初始化 Gin 引擎和所有必要的依赖项（Handler, Service, Repository），并将它们连接到测试数据库。
-   **执行**:
    -   在包目录下（例如 `backend/internal/router`）运行 `go test` 或 `go test -v` 来执行测试。
    -   测试应该被集成到 CI/CD 流程中，以确保代码变更不会破坏现有功能。

## 手动 API 测试

-   可以使用任何 REST 客户端工具（如 Postman, Insomnia, Hoppscotch 或 `curl`）进行手动测试。
-   详细的 API 端点定义和 `curl` 使用示例可以在各自的组件设计文档中找到，例如 `docs/design/components/auth_api.md`。
-   手动测试时，请确保后端服务正在运行，并连接到正确的开发数据库 (`data/effiplat.db`)。
