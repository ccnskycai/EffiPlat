# 项目任务检查表

## 阶段一：项目初始化与核心架构

-   [x] 初始化后端 Go 项目 (`/backend`)
-   [x] 初始化前端 Next.js 项目 (`/frontend`)
-   [x] 确定并配置数据库 (sqlite3)
-   [x] 完成数据库核心表结构设计 (`docs/design/components/database_design.md`)
-   [x] 设计 V1.0 API 接口 (`docs/api/api_design.md`) - 完成
    - [x] 定义基础原则、认证、响应结构
    - [x] 定义用户认证、用户管理、环境管理、资产管理(服务器)、服务管理、服务实例、Bug 管理、审计日志核心端点及结构
    - [x] 细化角色/权限、职责/职责组、业务管理 API 结构
    - [x] 添加更多字段和关联关系到响应体

## 阶段二：核心功能开发 (V1.0 - MVP)

### 后端 API

-   [x] 配置日志库 (`zap`) 并提供基础配置 (dev/prod)
-   [x] **集成数据库迁移工具 (`golang-migrate/migrate`)**
-   [x] **编写初始数据库迁移脚本 (创建核心表)**
-   [x] **(数据填充) 创建 Seeder 包和 Factory 包结构** (`internal/seed`, `internal/factories`)
-   [x] **(数据填充) 为核心模型实现 Factory 模式** (e.g., User)
-   [x] **(数据填充) 编写基础 Seeding 函数** (e.g., SeedUsers, SeedAll)
-   [x] **(数据填充) 提供 Seeding 执行入口** (e.g., cmd/seeder)
-   [ ] 设置日志实例的依赖注入机制
-   [ ] 实现用户认证 API (`/auth`)
-   [ ] 实现用户管理 API (`/users`)
-   [ ] 实现角色与权限管理基础 API (`/roles`)
-   [ ] 实现职责与职责组管理 API (`/responsibilities`, `/responsibility-groups`)
-   [ ] 实现环境管理 API (`/environments`)
-   [ ] 实现资产管理 API (服务器) (`/assets`)
-   [ ] 实现服务管理 API (`/services`)
-   [ ] 实现服务实例管理基础 API (`/service-instances`)
-   [ ] 实现业务管理 API (`/businesses`)
-   [ ] 实现 Bug 管理 API (`/bugs`)
-   [ ] 实现基础操作审计日志记录

### 前端 UI

-   [ ] 搭建前端整体布局和导航
-   [ ] 实现登录/登出页面
-   [ ] 实现用户管理界面
-   [ ] 实现环境管理界面
-   [ ] 实现资产管理界面 (服务器)
-   [ ] 实现服务管理界面
-   [ ] 实现 Bug 跟踪界面 (列表、创建、详情、状态更新)
-   [ ] 实现基础仪表盘 (Dashboard)

## 阶段三：测试与部署

-   [ ] 编写后端单元测试
-   [ ] **(推荐) 编写自动化数据库迁移测试**
-   [ ] 编写前端单元/集成测试
-   [ ] 完成 V1.0 功能测试
-   [ ] 准备 V1.0 部署文档 (`docs/design/deployment_strategy.md`)
-   [ ] 执行 V1.0 首次部署

## 阶段四：后续迭代 (V1.x)

-   [ ] 实现需求管理模块
-   [ ] 实现配置管理模块
-   [ ] 实现部署方式管理模块
-   [ ] 实现统一接口 (日志获取、部署等)
-   [ ] 完善数据采集器功能 (如果采用)
-   [ ] 细化权限控制 (RBAC)
-   [ ] 增强非功能性需求 (性能、安全、可用性)
-   [ ] 添加更多数据可视化报表
-   [ ] 集成外部系统 (Git, CI/CD, 监控等)

---

*注意：这是一个初步的任务列表，请根据 `docs/requirements/execution_plan.md` 和实际项目进展进行调整和细化。*