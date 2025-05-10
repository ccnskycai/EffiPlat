# EffiPlat

<!-- Add badges here: build status, coverage, license, etc. -->
<!-- Example: [![Build Status](...)](...) -->

本项目（EffiPlat）旨在开发一个部署在公司内部局域网的一体化协作平台，解决当前跨部门协作中的信息孤岛、流程脱节、环境配置管理复杂、问题追溯困难等问题，打通从需求到运维的完整工作流，提升整体工作效率和质量。

## Table of Contents

- [EffiPlat](#effiplat)
  - [Table of Contents](#table-of-contents)
  - [Tech Stack](#tech-stack)
  - [Directory Structure](#directory-structure)
  - [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Environment Variables](#environment-variables)
    - [Installation \& Running](#installation--running)
  - [Running Tests](#running-tests)
    - [1. 运行所有后端测试](#1-运行所有后端测试)
    - [2. 仅运行数据库迁移相关测试](#2-仅运行数据库迁移相关测试)
    - [3. 编写和运行单元测试](#3-编写和运行单元测试)
    - [4. 测试覆盖率](#4-测试覆盖率)
    - [5. 注意事项](#5-注意事项)
  - [Deployment](#deployment)
  - [Documentation](#documentation)
  - [Contributing](#contributing)
  - [License](#license)

## Tech Stack

主要技术选型（详情请参考各组件设计文档，如 [`docs/design/components/backend_server.md`](docs/design/components/backend_server.md), [`docs/design/components/frontend_design.md`](docs/design/components/frontend_design.md) 等）：

- **Frontend:** [Vite](https://vitejs.dev/) + [React](https://reactjs.org/) (using [TypeScript](https://www.typescriptlang.org/)), [Ant Design](https://ant.design/)
- **Backend:** [Go](https://golang.org/) (using [Gin](https://gin-gonic.com/) - 已选定), [GORM](https://gorm.io/)
- **Database:** [SQLite](https://www.sqlite.org/index.html) (根据 [`docs/design/components/database_design.md`](docs/design/components/database_design.md) 和任务清单初步选择)
- **Logging**: [Zap](https://github.com/uber-go/zap)
- **Configuration**: [Viper](https://github.com/spf13/viper)
- **Other:** Docker

## Directory Structure

```
.
├── backend/            # Go Backend Service
│   ├── cmd/
│   ├── data/           # Database files (e.g., SQLite)
│   ├── internal/
│   └── logs/           # Log files
├── configs/            # Configuration Files
├── docs/               # Project Documentation
│   ├── api/
│   ├── design/
│   └── requirements/
├── frontend/           # Vite + React Frontend Application
│   ├── public/
│   └── src/
├── README.md           # Project Root README
└── ...                 # Other Config Files (e.g., .gitignore)
```

## Getting Started

本节介绍如何在本地设置和运行 EffiPlat 开发环境。

### Prerequisites

确保你已安装以下软件：

- **Go:** 版本 >= 1.21 (根据 `backend/go.mod` 推断，请安装最新稳定版)。 [Go 下载地址](https://golang.org/dl/)
- **Migrate CLI:** 用于数据库迁移。
  ```bash
  # 确保包含 SQLite 驱动
  go install -tags 'sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```
  _验证安装:_ `migrate -version`
- **C Compiler (GCC):** 后端使用的 SQLite 驱动需要 CGO，因此需要安装 C 编译器。
  - **Windows:** 安装 [MinGW-w64](https://www.mingw-w64.org/downloads/) (推荐通过 [MSYS2](https://www.msys2.org/) 安装 `mingw-w64-x86_64-toolchain`) 并将其 `bin` 目录添加到系统 PATH。
  - **macOS:** 安装 Xcode Command Line Tools (`xcode-select --install`).
  - **Linux (Debian/Ubuntu):** `sudo apt update && sudo apt install build-essential`
  - **Linux (Fedora):** `sudo dnf groupinstall "Development Tools"`
    _验证安装:_ `gcc --version`

### Environment Variables

后端配置通过 `configs/` 目录下的 `yaml` 文件管理 (使用 Viper)。通常会有一个 `config.yaml` 或类似文件用于本地开发，并可能引用环境变量。请参考 `internal/pkg/config` 包了解具体配置加载逻辑。

### Installation & Running

1.  **克隆仓库:**

    ```bash
    git clone <repository-url>
    cd EffiPlat
    ```

2.  **后端设置:**

    ```bash
    # 进入后端目录
    cd backend

    # 安装 Go 依赖
    go mod tidy

    # 运行迁移 (应用最新结构):
    migrate -database "sqlite3://data/effiplat.db" -path internal/migrations up

    # 启动后端服务器
    go run cmd/api/main.go
    ```

    如果启动时遇到 CGO 相关错误，请确保 C 编译器已正确安装并配置在 PATH 中。
    服务器默认运行在 `http://localhost:<port>` (端口在配置文件中定义)。

3.  **回滚迁移 (撤销更改):**
    在 `backend` 目录下执行：

    - 回滚最后一次应用的迁移：
      ```bash
      migrate -database "sqlite3://data/effiplat.db" -path internal/migrations down 1
      ```
    - 回滚所有迁移 (回到空数据库状态)：
      ```bash
      migrate -database "sqlite3://data/effiplat.db" -path internal/migrations down -all
      ```
    - 迁移到指定的版本号 (例如，版本 2)：
      ```bash
      migrate -database "sqlite3://data/effiplat.db" -path internal/migrations goto 2
      ```

4.  **强制设置特定迁移版本 (修复错误状态):**
    如果迁移状态出错 (例如 `dirty` 状态)，可以强制设置当前数据库的迁移版本号。
    _注意：这不会执行 SQL，仅修改 `schema_migrations` 表。谨慎使用！_

    ```bash
    # 强制认为数据库当前是版本 3 且状态干净
    migrate -database "sqlite3://data/effiplat.db" -path internal/migrations force 3
    ```

5.  **运行数据填充 (Seeding):**
    在数据库结构迁移完成后，可以使用 Seeder 命令填充初始数据或测试数据：

    ```bash
    # 在 backend 目录下运行
    go run cmd/seeder/main.go
    ```

    _注意：Seeder 使用的数据库连接信息来自配置文件 (例如 `configs/config.dev.yaml`)，确保它指向你想要填充的数据库文件。Seeder 应该在迁移 (`up`) 完成后运行。_

6.  **创建新迁移 (开发过程中):**
    可以使用 `migrate` CLI 工具创建新的迁移文件框架：
    ```bash
    migrate create -ext sql -dir internal/migrations -seq init
    ```

## Running Tests

本项目后端采用 Go 的标准测试框架，所有测试文件以 `_test.go` 结尾。

### 1. 运行所有后端测试

在 `backend` 目录下执行：

```bash
go test ./...
```

这会递归运行所有包下的测试，包括单元测试、集成测试和迁移测试。

### 2. 仅运行数据库迁移相关测试

迁移测试文件位于 `backend/internal/migration_test.go`，可单独运行：

```bash
go test ./internal/migration_test.go
```

或

```bash
go test ./internal/
```

### 3. 编写和运行单元测试

- 推荐在每个包（如 `internal/api/`, `internal/service/` 等）下为每个功能模块编写对应的 `_test.go` 文件。
- 运行方式同上，或指定具体包/文件。

### 4. 测试覆盖率

可选：查看测试覆盖率报告

```bash
go test ./... -cover
```

### 5. 注意事项

- 迁移测试会在内存数据库中运行，不影响实际开发/生产数据库。
- 集成测试建议使用测试专用数据库或内存数据库，避免污染业务数据。
- 前端测试相关内容将在前端开发启动后补充。

## Deployment

(待补充: 描述部署流程或链接到相关文档，参考 [docs/design/deployment_strategy.md](docs/design/deployment_strategy.md))

## Documentation

详细的项目文档位于 [`/docs`](./docs/) 目录，主要包括：

- **需求文档**: [`docs/requirements/`](./docs/requirements/)
- **系统设计**: [`docs/design/`](./docs/design/)
- **API 文档**: [`docs/api/`](./docs/api/)

项目特定的编码规范和 AI 协作指南定义在 `.cursor/rules/` 目录下的规则文件中，并由 Cursor AI 助手自动应用。

## Contributing

Information on how to contribute to the project. (e.g., contribution guidelines, code of conduct).

(Add details or link to CONTRIBUTING.md)

## License

Specify the project license.

(e.g., This project is licensed under the MIT License - see the LICENSE file for details.)
