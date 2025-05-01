# 局域网一体化协作平台

<!-- Add badges here: build status, coverage, license, etc. -->
<!-- Example: [![Build Status](...)](...) -->

本项目旨在开发一个部署在公司内部局域网的一体化协作平台，解决当前跨部门协作中的信息孤岛、流程脱节、环境配置管理复杂、问题追溯困难等问题，打通从需求到运维的完整工作流，提升整体工作效率和质量。

## Table of Contents

- [局域网一体化协作平台](#局域网一体化协作平台)
  - [Table of Contents](#table-of-contents)
  - [Tech Stack](#tech-stack)
  - [目录结构](#目录结构)
  - [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Environment Variables](#environment-variables)
    - [Installation \& Running](#installation--running)
  - [Running Tests](#running-tests)
  - [Deployment](#deployment)
  - [文档](#文档)
  - [Contributing](#contributing)
  - [License](#license)

## Tech Stack

根据初步技术选型 ([docs/design/technology_stack.md](docs/design/technology_stack.md)):

*   **Frontend:** [Next.js](https://nextjs.org/) (using [TypeScript](https://www.typescriptlang.org/)), [React](https://reactjs.org/), [Tailwind CSS](https://tailwindcss.com/) (已配置), [Ant Design](https://ant.design/) / [MUI](https://mui.com/) (待定 UI 库)
*   **Backend:** [Go](https://golang.org/) (using [Gin](https://gin-gonic.com/) or [Echo](https://echo.labstack.com/) - 待定), [GORM](https://gorm.io/)
*   **Database:** [SQLite](https://www.sqlite.org/index.html) (根据 [docs/design/components/database_design.md](docs/design/components/database_design.md) 和任务清单初步选择) 
*   **Other:** Docker

## 目录结构

```
.
├── .cursor/            # Cursor AI 配置和规则
│   ├── rules/          # 项目特定的 AI 规则
│   │   ├── commit-message-format.md
│   │   ├── rule-git-standards.mdc
│   │   ├── rule-go-coding-standards.mdc
│   │   ├── rule-next-coding-standards.mdc
│   │   ├── rule-tailwind-coding-standards.mdc
│   │   └── rule-tailwind-v4-ext.mdc
│   └── settings.json
├── .cursorignore       # Cursor AI 忽略文件
├── .gitignore          # Git 忽略文件配置
├── README.md           # 项目根 README
├── backend/            # Go 后端服务
│   ├── cmd/
│   │   └── api/        # API 服务入口
│   ├── go.mod          # Go 模块依赖
│   └── internal/       # 内部包
│       ├── api/        # API 路由和处理器
│       ├── config/     # 配置加载
│       ├── domain/     # 业务领域模型
│       └── store/      # 数据存储
├── docs/               # 项目文档
│   ├── README.md       # 文档入口
│   ├── api/            # API 文档
│   │   └── api_design.md
│   ├── design/         # 系统设计文档
│   │   ├── README.md
│   │   ├── architecture.md
│   │   ├── components/ # 组件设计
│   │   │   ├── backend_server.md
│   │   │   ├── database_design.md
│   │   │   └── frontend_client.md
│   │   ├── deployment_strategy.md
│   │   ├── error_codes.md
│   │   └── security_design.md
│   ├── requirements/   # 需求文档
│   │   ├── execution_plan.md
│   │   └── requirements.md
│   └── task_checklist.md # 任务检查清单
└── frontend/           # Next.js 前端应用
    ├── app/            # App Router (Next.js 13+)
    │   └── layout.js   # 根布局
    ├── components/     # 共享 React 组件
    ├── next.config.js  # Next.js 配置文件
    ├── package.json    # Node.js 依赖
    ├── postcss.config.js # PostCSS 配置文件
    ├── public/         # 静态资源 (图片, 字体等)
    ├── styles/         # 样式文件
    │   └── globals.css # 全局 CSS
    └── tailwind.config.js # Tailwind CSS 配置文件
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

## 文档

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