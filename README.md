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
│   └── internal/
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

(待补充: 如何设置和运行本地开发环境)

### Prerequisites

(待补充: 列出所需软件和版本，如 Node.js, Go, Docker, 数据库等)

### Environment Variables

(待补充: 说明前后端所需环境变量及 `.env.example` 文件)

### Installation & Running

(待补充: 提供详细的克隆、安装依赖、启动前后端服务的步骤)

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