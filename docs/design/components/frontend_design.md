# 前端设计文档

## 1. 技术栈

- **构建工具**: Vite
- **核心框架**: React
- **编程语言**: TypeScript
- **UI 框架**: Ant Design
- **路由**: React Router DOM
- **状态管理**: React Context，可根据后续复杂度调整。
- **数据请求**: React Query (TanStack Query)
- **代码规范**: ESLint, Prettier

## 2. 项目初始化与结构

项目将使用 `npm create vite@latest frontend -- --template react-ts` (或 yarn) 初始化。

预期的核心目录结构：

```
frontend/
├── public/               # 静态资源，会被直接拷贝到构建输出目录
├── src/
│   ├── App.tsx           # 根应用组件，包含路由配置
│   ├── main.tsx          # 应用入口，渲染 App 组件，全局样式导入
│   ├── assets/           # 项目特定静态资源 (图片、字体等)
│   ├── components/       # 可复用的 UI 组件 (按功能或原子设计组织)
│   │   ├── common/         # 通用基础组件 (Button, Input, Modal等封装或自定义)
│   │   └── layout/         # 布局相关组件 (Header, Sidebar, Footer)
│   ├── config/           # 应用配置 (如 API 地址、常量)
│   ├── contexts/         # React Context (如果使用)
│   ├── hooks/            # 自定义 React Hooks
│   ├── layouts/          # 页面级布局组件 (如 AuthLayout, DashboardLayout)
│   ├── pages/            # 页面级组件 (每个路由对应一个或多个组件)
│   │   ├── Auth/
│   │   │   └── LoginPage.tsx
│   │   ├── DashboardPage.tsx
│   │   ├── UsersPage.tsx
│   │   └── ...
│   ├── services/         # API 请求服务 (封装 fetch/axios, SWR/React Query 配置)
│   ├── store/            # 状态管理 (如 Zustand 的 store 定义)
│   ├── styles/           # 全局样式、主题配置、工具类
│   │   └── global.css
│   ├── types/            # TypeScript 类型定义 (共享的接口、类型)
│   └── utils/            # 通用工具函数
├── index.html            # Vite 入口 HTML
├── vite.config.ts        # Vite 配置文件
├── tsconfig.json         # TypeScript 配置文件
├── package.json
└── ...                 # 其他配置文件 (.eslintrc, .prettierrc)
```

## 3. 核心功能模块规划 (V1.0 MVP)

根据后端 API 设计和任务清单，前端 V1.0 核心模块包括：

- **认证模块**:
  - 登录页面 (`/login`)
  - 处理登录逻辑，存储/清除认证 Token (JWT)
  - 路由守卫/重定向未认证用户
- **主布局**:
  - 包含顶部导航栏/Header (显示用户信息、登出按钮)
  - 侧边导航栏 (根据用户角色和权限动态生成菜单项)
  - 内容区域
- **用户管理模块 (CRUD)**:
  - 用户列表展示 (`/users`)
  - 创建用户、编辑用户、查看用户详情
- **环境管理模块 (CRUD)**
- **资产管理模块 (服务器 - CRUD)**
- **服务管理模块 (CRUD)**
- **Bug 管理模块 (CRUD)**:
  - Bug 列表展示、筛选、排序
  - 创建 Bug、编辑 Bug、查看 Bug 详情
  - Bug 状态流转
- **仪表盘 (Dashboard)**:
  - 展示关键摘要信息和统计图表 (待定)

## 4. 路由设计 (React Router DOM)

- 使用 `BrowserRouter`。
- 主要路由将在 `App.tsx` 中定义。
- 示例：
  - `/login` -> `LoginPage.tsx`
  - `/` (或 `/dashboard`) -> `DashboardPage.tsx` (需要认证)
  - `/users` -> `UserListPage.tsx` (需要认证)
  - `/users/:id` -> `UserDetailPage.tsx` (需要认证)
  - 其他模块类似。
- 考虑使用嵌套路由和 Outlet 来实现共享布局下的子页面切换。
- 需要实现私有路由组件/Hook (`PrivateRoute`)，用于检查用户认证状态，未认证则重定向到登录页。

## 5. UI 框架集成

- **选择**: 待团队最终确认选择 Ant Design 或 Material UI。
- **集成**:
  - 在 `main.tsx` 中引入全局样式 (或配置按需加载)。
  - 封装常用的 UI 组件为项目内的 `common` 组件，方便统一风格和未来可能的替换。
  - 考虑主题定制能力。

## 6. API 通信

- 在 `src/services` 或 `src/lib/api.ts` 中封装 API 请求逻辑。
- 使用 `axios` 或 `fetch` API。
- 统一处理请求头 (如 `Authorization: Bearer <token>`)。
- 统一处理 API 响应和错误。
- 使用 SWR 或 React Query 进行数据获取、缓存和状态同步。
- 开发时，在 `vite.config.ts` 中配置 API 代理到后端服务 (如 `http://localhost:8080`) 以解决跨域问题。

## 7. 状态管理

- **用户认证状态**: 全局管理用户登录状态和用户信息 (可使用 React Context 或 Zustand)。
- **表单状态**: 使用组件内部状态或 React Hook Form / Formik。
- **服务端缓存/UI 状态**: SWR/React Query 处理大部分与服务端数据相关的状态。
- **其他全局 UI 状态**: 如通知、模态框状态等，可根据需要选择合适的方案。

## 8. 注意事项

- **代码分割**: Vite 天然支持基于路由的代码分割。
- **TypeScript 类型**: 为 props, API 响应/请求, store state 等定义明确的类型。
- **组件组织**: 遵循单一职责原则，保持组件的小巧和可复用性。
- **错误处理**: 在 UI 和 API 调用层面进行适当的错误处理和用户反馈。
- **可访问性 (a11y)**: 开发时关注基本的 Web 可访问性标准。
