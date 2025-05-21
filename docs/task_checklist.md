# 项目任务检查表

## 阶段一：项目初始化与核心架构

- [x] 初始化后端 Go 项目 (`/backend`)
- [x] 初始化前端 Vite + React 项目 (`/frontend`)
- [x] 确定并配置数据库 (sqlite3)
- [x] 完成数据库核心表结构设计 (`docs/design/components/database_design.md`)
- [x] 设计 V1.0 API 接口 (`docs/api/api_design.md`) - 完成
  - [x] 定义基础原则、认证、响应结构
  - [x] 定义用户认证、用户管理、环境管理、资产管理(服务器)、服务管理、服务实例、Bug 管理、审计日志核心端点及结构
  - [x] 细化角色/权限、职责/职责组、业务管理 API 结构
  - [x] 添加更多字段和关联关系到响应体

## 阶段二：核心功能开发 (V1.0 - MVP)

### 后端 API

- [x] 配置日志库 (`zap`) 并提供基础配置 (dev/prod)
- [x] **集成数据库迁移工具 (`golang-migrate/migrate`)**
- [x] **编写初始数据库迁移脚本 (创建核心表)**
- [x] **(数据填充) 创建 Seeder 包和 Factory 包结构** (`internal/seed`, `internal/factories`)
- [x] **(数据填充) 为核心模型实现 Factory 模式** (e.g., User)
- [x] **(数据填充) 编写基础 Seeding 函数** (e.g., SeedUsers, SeedAll)
- [x] **(数据填充) 提供 Seeding 执行入口** (e.g., cmd/seeder)
- [x] **(推荐) 编写自动化数据库迁移测试** (基础结构完成)
- [x] 设置日志实例的依赖注入机制
- [x] 实现用户认证 API (`/auth`) - 登录、获取用户、登出 完成
- [x] 实现用户管理 API (`/users`) - 后端已完成
- [x] 实现角色与权限管理基础 API (`/roles`, `/permissions`, 角色权限关联) - 后端已完成 (基本路由测试已覆盖)
- [x] 实现职责与职责组管理 API (`/responsibilities`, `/responsibility-groups`) - 后端已完成 (基本路由测试已覆盖)
- [x] 实现环境管理 API (`/environments`) - 基本完成
  - [x] 设计 Environment 模型 (models/environment.go)
  - [x] 检查数据库迁移 (已在初始迁移中包含 environments 表)
  - [x] 实现 Service 层 (service/environment_service.go) - 基本完成
  - [x] 实现 Repository 层 (repository/environment_repository.go) - 基本完成
  - [x] 实现 Handler 层 (handlers/environment_handler.go) - 结构和占位符方法完成
  - [x] 注册路由 (router.go) - 完成
  - [x] 编写测试 (router/environment_router_test.go) - 测试通过
  - [ ] (注意) 检查并实现 `alphanumdash` 校验器 (如果需要)
- [x] 实现资产管理 API (服务器) (`/assets`) - 后端基本 CRUD 和路由测试完成
- [ ] 实现服务管理 API (`/services`)
- [ ] 实现服务实例管理基础 API (`/service-instances`)
- [ ] 实现业务管理 API (`/businesses`)
- [ ] 实现 Bug 管理 API (`/bugs`)
- [ ] 实现基础操作审计日志记录

### 前端 UI

- [x] **项目初始化与基础框架 (部分完成)**
  - [x] 选择前端技术栈: Vite + React + TypeScript
  - [x] 集成 UI 组件库: Ant Design
  - [x] 配置 ESLint 和 Prettier (验证) - ✅ Linter 警告已处理, Prettier 格式已修复 (MainLayout.tsx)
- [ ] **UI 组件实现 (基于 Ant Design)**
  - [x] 搭建前端整体布局和导航 (此项可与路由部分合并)
  - [x] 实现登录/登出页面
  - [ ] **仪表盘 (Dashboard) 页面核心组件**
    - [ ] 概览统计卡片
    - [ ] 任务统计表
    - [ ] 用户统计表
    - [ ] 环境统计表
    - [ ] 资产统计表
    - [ ] 服务统计表
    - [ ] 右侧边栏组件 (用户信息、日历、日程、最近活动)
  - [ ] 实现用户管理界面
  - [ ] 实现环境管理界面
  - [ ] 实现资产管理界面 (服务器)
  - [ ] 实现服务管理界面
  - [ ] 实现 Bug 跟踪界面 (列表、创建、详情、状态更新)
- [ ] **数据交互与状态管理**
  - [ ] API 服务层封装 (选择 HTTP 客户端, 封装基础请求模块)
  - [ ] 服务端数据状态管理 (推荐 `TanStack Query` / SWR: 安装配置, 实现数据获取/缓存逻辑)
  - [ ] 全局客户端状态管理 (按需评估并选择方案)
- [ ] **测试**
  - [ ] 单元/组件测试环境 (Vitest + React Testing Library: 安装配置, 编写测试)
  - [ ] 端到端测试 (E2E - 可选评估)

## 阶段三：测试与部署

- [ ] 编写后端单元测试 (路由层测试部分完成)
  - [x] 路由层测试: 用户认证 (auth)
  - [x] 路由层测试: 用户管理 (users CRUD)
  - [x] 路由层测试: 角色管理 (roles CRUD)
  - [x] 路由层测试: 权限管理 (permissions CRUD)
  - [x] 路由层测试: 角色权限分配 (permissions/roles/:id, roles/:id/permissions)
  - [x] 路由层测试: 用户角色分配 (users/:id/roles)
  - [x] 路由层测试: 职责管理 (responsibilities CRUD)
  - [x] 路由层测试: 职责组管理 (responsibility-groups CRUD, 职责组与职责关联)
  - [x] 路由层测试: 环境管理 (environments CRUD) - 测试通过
  - [ ] 服务层单元测试 (待进行)
  - [ ] 仓库层单元测试 (待进行)
- [ ] 编写前端单元/集成测试
- [ ] 完成 V1.0 功能测试
- [ ] 准备 V1.0 部署文档 (`docs/design/deployment_strategy.md`)
- [ ] 执行 V1.0 首次部署

## 阶段四：后续迭代 (V1.x)

- [ ] 实现需求管理模块
- [ ] 实现配置管理模块
- [ ] 实现部署方式管理模块
- [ ] 实现统一接口 (日志获取、部署等)
- [ ] 完善数据采集器功能 (如果采用)
- [ ] 细化权限控制 (RBAC)
- [ ] 增强非功能性需求 (性能、安全、可用性)
- [ ] 添加更多数据可视化报表
- [ ] 集成外部系统 (Git, CI/CD, 监控等)

---

_注意：这是一个初步的任务列表，请根据 `docs/requirements/execution_plan.md`