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
  - [Deployment](#deployment)
  - [Documentation](#documentation)
  - [Contributing](#contributing)
  - [License](#license)

## Tech Stack

主要技术选型（详情请参考各组件设计文档，如 [`docs/design/components/backend_server.md`](docs/design/components/backend_server.md), [`docs/design/components/frontend_client.md`](docs/design/components/frontend_client.md) 等）：

*   **Frontend:** [Next.js](https://nextjs.org/) (using [TypeScript](https://www.typescriptlang.org/)), [React](https://reactjs.org/), [Tailwind CSS](https://tailwindcss.com/) (已配置), [Ant Design](https://ant.design/) / [MUI](https://mui.com/) (待定 UI 库)
*   **Backend:** [Go](https://golang.org/) (using [Gin](https://gin-gonic.com/) - 已选定), [GORM](https://gorm.io/)
*   **Database:** [SQLite](https://www.sqlite.org/index.html) (根据 [`docs/design/components/database_design.md`](docs/design/components/database_design.md) 和任务清单初步选择)
*   **Logging**: [Zap](https://github.com/uber-go/zap)
*   **Configuration**: [Viper](https://github.com/spf13/viper)
*   **Other:** Docker

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
├── frontend/           # Next.js Frontend Application
│   ├── app/
│   ├── components/
│   ├── public/
│   └── styles/
├── README.md           # Project Root README
└── ...                 # Other Config Files (e.g., .gitignore)
```

## Getting Started

本节介绍如何在本地设置和运行 EffiPlat 开发环境。

### Prerequisites

确保你已安装以下软件：

*   **Go:** 版本 >= 1.21 (根据 `backend/go.mod` 推断，请安装最新稳定版)。 [Go 下载地址](https://golang.org/dl/)
*   **Migrate CLI:** 用于数据库迁移。
    ```bash
    # 确保包含 SQLite 驱动
    go install -tags 'sqlite' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    ```
    *验证安装:* `migrate -version`
*   **C Compiler (GCC):** 后端使用的 SQLite 驱动需要 CGO，因此需要安装 C 编译器。
    *   **Windows:** 安装 [MinGW-w64](https://www.mingw-w64.org/downloads/) (推荐通过 [MSYS2](https://www.msys2.org/) 安装 `mingw-w64-x86_64-toolchain`) 并将其 `bin` 目录添加到系统 PATH。
    *   **macOS:** 安装 Xcode Command Line Tools (`xcode-select --install`).
    *   **Linux (Debian/Ubuntu):** `sudo apt update && sudo apt install build-essential`
    *   **Linux (Fedora):** `sudo dnf groupinstall "Development Tools"`
    *验证安装:* `gcc --version`

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

    # 运行数据库迁移 (确保数据库文件目录存在: data/)
    # 这会创建或更新 backend/data/effiplat.db 文件
    migrate -database "sqlite3://data/effiplat.db" -path internal/migrations up

    # 启动后端服务器
    go run cmd/api/main.go
    ```
    如果启动时遇到 CGO 相关错误，请确保 C 编译器已正确安装并配置在 PATH 中。
    服务器默认运行在 `http://localhost:<port>` (端口在配置文件中定义)。

3.  **前端设置:**
    

## Running Tests

(待补充: 提供运行前后端测试的命令)

## Deployment

(待补充: 描述部署流程或链接到相关文档，参考 [docs/design/deployment_strategy.md](docs/design/deployment_strategy.md))

## Documentation

详细的项目文档位于 [`/docs`](./docs/) 目录，主要包括：

*   **需求文档**: [`docs/requirements/`](./docs/requirements/)
*   **系统设计**: [`docs/design/`](./docs/design/)
*   **API 文档**: [`docs/api/`](./docs/api/)

项目特定的编码规范和 AI 协作指南定义在 `.cursor/rules/` 目录下的规则文件中，并由 Cursor AI 助手自动应用。

## Contributing

Information on how to contribute to the project. (e.g., contribution guidelines, code of conduct).

(Add details or link to CONTRIBUTING.md)

## License

Specify the project license.

(e.g., This project is licensed under the MIT License - see the LICENSE file for details.)