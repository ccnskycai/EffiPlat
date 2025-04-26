# 局域网一体化协作平台

<!-- Add badges here: build status, coverage, license, etc. -->
<!-- Example: [![Build Status](...)](...) -->

本项目旨在开发一个部署在公司内部局域网的一体化协作平台，解决当前跨部门协作中的信息孤岛、流程脱节、环境配置管理复杂、问题追溯困难等问题，打通从需求到运维的完整工作流，提升整体工作效率和质量。

## Table of Contents

- [局域网一体化协作平台](#局域网一体化协作平台)
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

根据初步技术选型 ([design/technology_stack.md](design/technology_stack.md)):

*   **Frontend:** [Next.js](https://nextjs.org/) (using [TypeScript](https://www.typescriptlang.org/)), [React](https://reactjs.org/), [Ant Design](https://ant.design/) / [MUI](https://mui.com/) (待定)
*   **Backend:** [Go](https://golang.org/) (using [Gin](https://gin-gonic.com/) or [Echo](https://echo.labstack.com/) - 待定), [GORM](https://gorm.io/) (初步选定)
*   **Database:** [PostgreSQL](https://www.postgresql.org/) (初步倾向) / MySQL
*   **Other:** Docker

## Directory Structure

```
/
├── .github/          # CI/CD workflows (可选)
├── .cursor/          # Cursor AI configuration and rules
│   └── rules/        # Project-specific AI rules (.mdc files)
├── backend/          # Go 后端应用
│   ├── cmd/api/main.go # 入口点
│   ├── internal/     # 内部应用代码
│   ├── go.mod        # 依赖
│   └── ...
├── docs/             # 项目文档 (需求、设计等)
│   ├── requirements/ # 需求相关文档
│   │   ├── requirements.md
│   │   └── execution_plan.md
│   └── design/       # 设计相关文档
│       ├── README.md
│       └── ...
├── frontend/         # Next.js 前端应用
│   ├── app/          # App Router (Next.js 13+)
│   ├── components/   # 共享组件
│   ├── public/       # 静态资源
│   ├── styles/       # 全局样式 & Tailwind (如果使用)
│   ├── package.json  # 依赖
│   └── ...
├── .gitignore
└── README.md
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

(待补充: 描述部署流程或链接到相关文档，参考 [design/deployment_strategy.md](design/deployment_strategy.md))

## Documentation

详细的项目文档位于 [`/docs`](./docs/) 目录，主要包括：

*   **需求文档**: [docs/requirements/requirements.md](docs/requirements/requirements.md)
*   **执行计划**: [docs/requirements/execution_plan.md](docs/requirements/execution_plan.md)
*   **系统设计**: [design/README.md](design/README.md) (包含详细设计文档入口)

项目特定的编码规范和 AI 协作指南定义在 `.cursor/rules/` 目录下的规则文件中，并由 Cursor AI 助手自动应用。

## Contributing

Information on how to contribute to the project. (e.g., contribution guidelines, code of conduct).

(Add details or link to CONTRIBUTING.md)

## License

Specify the project license.

(e.g., This project is licensed under the MIT License - see the LICENSE file for details.) 