# 登录/登出功能详细设计文档

## 1. 引言

本文档详细描述了 EffiPlat 项目中用户登录和登出功能的前端实现方案。设计目标是提供一个安全、用户友好且易于维护的认证流程。

## 2. 相关文档

本设计基于以下 API 文档：

- [API 设计规范](../design/components/api_design.md)
- [用户认证 API 设计方案](../design/components/auth_api.md)
- [效果图](../ui_references/login)

**核心原则**：所有 API 交互将严格遵循 `docs/design/components/api_design.md` 中定义的统一响应格式。

## 3. API 接口回顾

以下是认证功能依赖的核心 API 端点，均位于 `/api/v1/auth` 路径下。

### 3.1 用户登录

- **Endpoint**: `POST /login`
- **请求体**:
  ```json
  {
    "email": "user@example.com",
    "password": "your_password"
  }
  ```
- **成功响应 (200 OK)**:
  ```json
  {
    "code": 0,
    "message": "Login successful",
    "data": {
      "token": "jwt_token_string",
      "userId": 1,
      "name": "John Doe",
      "roles": ["admin", "developer"]
    }
  }
  ```
- **错误响应 (例如 401 Unauthorized)**:
  ```json
  {
    "code": 40101, // 具体业务错误码
    "message": "Invalid credentials",
    "data": null
  }
  ```

### 3.2 获取当前用户信息

- **Endpoint**: `GET /me`
- **认证**: 需要 `Authorization: Bearer <token>` 请求头。
- **成功响应 (200 OK)**:
  ```json
  {
    "code": 0,
    "message": "User info retrieved",
    "data": {
      "id": 1,
      "name": "John Doe",
      "email": "user@example.com"
      // 建议也包含 roles，便于前端权限管理
    }
  }
  ```

### 3.3 用户登出

- **Endpoint**: `POST /logout`
- **认证**: 需要 `Authorization: Bearer <token>` 请求头。
- **成功响应 (200 OK)**:
  ```json
  {
    "code": 0,
    "message": "Logout successful",
    "data": null
  }
  ```

## 4. Token 管理：存储与生命周期

### 4.1 Token 存储方式

#### 4.1.1 选择的存储位置：`localStorage`

- **原因与工作方式**：用户成功登录后，后端返回 JWT。前端需将其持久化存储，以便后续请求时使用。`localStorage` 提供了这种浏览器端的持久化存储能力（关闭标签或窗口后数据仍在），且实现简单。具体操作为：
  - 登录成功：`localStorage.setItem('authToken', receivedToken);`
  - 读取 Token：`const token = localStorage.getItem('authToken');` 然后将其放入请求头。
- **与 API 设计的契合**：当前 API 设计（`docs/design/components/api_design.md`）是在登录接口的响应体中直接返回 Token，这通常意味着期望由前端负责存储和管理 Token。

#### 4.1.2 其他存储方式考量

- **`sessionStorage`**：生命周期与浏览器会话绑定，关闭标签页即清除。不适用于需要持久登录的场景。
- **HTTP-only Cookies**: 更安全的选择，Token 由后端通过`Set-Cookie`头设置并标记为`HttpOnly`，前端 JavaScript 无法直接访问，可有效防止 XSS 窃取 Token。浏览器会自动处理后续请求的 Cookie 携带。
  - **当前未选用的原因**：需要后端 API 设计配合（直接设置 Cookie 而非在响应体返回 Token）。若未来 API 调整，可重新评估此方案。

### 4.2 Token 生命周期详解

JWT 的生命周期通常包括以下关键阶段：

#### 4.2.1 签发 (Issuance)

- 用户提交凭据后，后端验证成功，生成 JWT。
- JWT 内部包含用户信息、权限及一个**过期时间 (Expiration Time)**。
- JWT 通过 API 响应体返回给前端。

#### 4.2.2 前端存储 (Storage by Frontend)

- 前端从登录响应中提取 JWT。
- 将 JWT 存储在`localStorage`中。
- 更新应用内认证状态（用户已登录、用户信息等）。

#### 4.2.3 使用 (Usage in API Calls)

- 用户执行需认证操作时，前端从`localStorage`读取 JWT。
- 将 JWT 添加到 HTTP 请求的`Authorization`头部（通常为 `Bearer <JWT_TOKEN_STRING>`）。
- 此逻辑通常由 HTTP 客户端库的请求拦截器统一处理。

#### 4.2.4 服务端验证 (Verification by Backend)

- 后端接收到请求后，提取 JWT。
- 验证 JWT 的签名（确保未被篡改且由后端签发）和是否已过期。
- 验证通过则处理请求，否则返回错误（通常是`401 Unauthorized`）。

#### 4.2.5 过期处理 (Expiration Handling)

- **后端行为**：收到过期的 JWT 时，拒绝请求并返回`401 Unauthorized`。
- **前端行为**：
  1. HTTP 响应拦截器捕获到`401`错误。
  2. 执行登出逻辑：清除`localStorage`中的 Token 和应用内的认证状态。
  3. 重定向用户到登录页面。
- **注意**：当前设计未包含客户端主动检查 Token 过期时间或静默刷新 Token 的机制。依赖后端返回的`401`进行处理。

#### 4.2.6 主动登出与失效 (Invalidation/Logout)

- **用户主动登出**：
  1. 用户点击"登出"按钮。
  2. 前端调用`POST /api/v1/auth/logout`接口。
  3. **核心步骤**：前端必须从`localStorage`中删除 JWT (`localStorage.removeItem('authToken');`)。
  4. 清除应用内的认证状态，重置用户信息。
  5. 重定向用户到登录页面。
- **服务端 Token 状态**：根据`docs/design/components/auth_api.md`，当前后端登出接口可能不实现 Token 黑名单。这意味着登出主要依赖前端清除 Token。理论上，在 Token 自然过期前，若被泄露，仍可能被使用。

#### 4.2.7 未来考虑：Token 刷新与服务端黑名单

- **Token 刷新机制**：为提升用户体验，未来可考虑引入 Refresh Token 机制，允许在 Access Token 过期后静默获取新的 Access Token，避免用户频繁重登。这需要后端 API 支持。
- **服务端 Token 黑名单**：为增强安全性，尤其是在主动登出或检测到 Token 泄露时，后端可实现 Token 黑名单机制，使特定 Token 立即失效，即使它们尚未达到自然过期时间。

## 5. 前端实现方案

### 5.1 建议文件结构

```
frontend/src/
├── pages/
│   └── LoginPage.tsx
├── components/
│   └── auth/
│       └── LoginForm.tsx
├── services/
│   └── authService.ts
├── store/              // 或 context/
│   └── authStore.ts    // 或 authContext.tsx / useAuth.ts (自定义 Hook)
├── layouts/
│   └── MainLayout.tsx  // 可能包含登出按钮
├── App.tsx             // 路由配置
└── main.tsx            // 应用入口，可能包含全局初始化逻辑
```

### 5.2 API 服务层 (`services/authService.ts`)

此文件将封装所有与认证相关的 API 调用。

- **`login(credentials: LoginCredentials): Promise<LoginResponseData>`**:
  - 发送 `POST /api/v1/auth/login` 请求。
  - 处理响应，成功时返回 `response.data.data` (包含 `token` 和 `user` 信息)。
  - 处理错误，检查 `response.data.code` 并抛出或返回结构化错误。
- **`logout(): Promise<void>`**:
  - 发送 `POST /api/v1/auth/logout` 请求。
  - 处理响应和错误。
- **`getCurrentUser(): Promise<User>`**:
  - 发送 `GET /api/v1/auth/me` 请求。
  - 成功时返回 `response.data.data` (用户信息)。
  - 处理错误。

```typescript
// 示例接口定义
interface LoginCredentials {
  email: string;
  password: string;
}

interface User {
  id: number;
  name: string;
  email: string;
  roles?: string[]; // 根据实际API响应调整
}

interface LoginResponseData {
  token: string;
  userId?: number; // 或 user: User;
  name?: string;
  email?: string;
  roles?: string[];
  // 如果API返回嵌套的user对象
  user?: User;
}
```

### 5.3 状态管理 (例如 `store/authStore.ts` 使用 Zustand 或 React Context)

负责全局管理用户认证状态。

- **State**:
  - `isAuthenticated: boolean`
  - `user: User | null`
  - `token: string | null`
  - `isLoading: boolean` (用于登录等异步操作)
  - `error: string | null` (用于显示认证相关的错误信息)
- **Actions**:
  - **`login(credentials: LoginCredentials)`**:
    1.  设置 `isLoading` 为 `true`，清除 `error`。
    2.  调用 `authService.login(credentials)`。
    3.  成功时：
        - 将返回的 `token` 存储到 `localStorage`。
        - 更新状态：`isAuthenticated = true`, `user = extracted_user_info`, `token = new_token`。
        - 配置 HTTP 客户端默认携带此 `token`。
        - 设置 `isLoading` 为 `false`。
    4.  失败时：
        - 更新状态：`error = error_message`。
        - 设置 `isLoading` 为 `false`。
  - **`logout()`**:
    1.  调用 `authService.logout()` (即使失败也应继续前端清理)。
    2.  清除 `localStorage` 中的 `token`。
    3.  重置状态：`isAuthenticated = false`, `user = null`, `token = null`。
    4.  从 HTTP 客户端默认配置中移除 `token`。
  - **`loadUserFromToken()` 或 `initializeAuth()`**:
    1.  从 `localStorage` 读取 `token`。
    2.  如果 `token` 存在：
        - 配置 HTTP 客户端默认携带此 `token`。
        - 调用 `authService.getCurrentUser()`。
        - 成功时：更新状态 `isAuthenticated = true`, `user = user_info`, `token = stored_token`。
        - 失败时（如 token 失效）：调用 `logout()` 清理。
- **Token Storage**: 使用 `localStorage` 存储 JWT。

### 5.4 UI 组件

本节描述的 UI 组件在实现时，应参考 `../ui_references/login` 目录下提供的设计效果图。

#### 5.4.1 `components/auth/LoginForm.tsx`

- 使用 Ant Design 组件 (`Form`, `Input.Password`, `Input`, `Button`, `Alert`)。
- 表单字段：邮箱、密码。
- 客户端校验：必填项、邮箱格式。
- 提交逻辑：调用状态管理中的 `login` action。
- 根据状态管理中的 `isLoading` 和 `error` 显示加载指示和错误提示。

#### 5.4.2 `pages/LoginPage.tsx`

- 页面布局，居中显示 `LoginForm` 组件。
- 监听状态管理中的 `isAuthenticated` 状态，如果已认证，则重定向到仪表盘或用户之前尝试访问的页面。
- 在 `login` action 成功后，处理页面重定向逻辑。

### 5.5 登出逻辑

- 通常在主布局（如 `layouts/MainLayout.tsx`）的用户菜单中提供一个"登出"按钮。
- 点击按钮时，调用状态管理中的 `logout` action。
- `logout` action 成功后（或无论如何都执行），重定向到 `/login` 页面。

## 6. 路由设计 (`App.tsx` 或路由配置文件)

- **Public Routes**:
  - `/login`: 指向 `LoginPage.tsx`。
- **Protected Routes**:
  - 所有需要认证才能访问的页面（如 `/dashboard`, `/users`, `/environments` 等）。
  - 实现方式：创建一个 `ProtectedRoute` 高阶组件或使用 `react-router-dom` v6 的 `<Outlet />` 和布局路由。
  - 在 `ProtectedRoute` 中检查 `isAuthenticated` 状态。如果未认证，重定向到 `/login`，可以考虑传递原始目标路径（`from` location state）以便登录后重定向回去。
- **访问 `/login` 时的行为**: 如果用户已认证并尝试访问 `/login`，应自动重定向到仪表盘页面。

## 7. UI/UX 注意事项

- **加载状态**: 在登录请求期间，登录按钮应显示加载状态，并禁用表单。
- **错误提示**:
  - 客户端校验错误应实时显示在表单字段下方。
  - 服务端返回的登录错误（如"无效凭证"）应通过 `Alert` 组件清晰展示。
- **成功反馈**: 登录成功后，通常通过页面重定向来体现，无需额外成功消息。
- **重定向**:
  - 登录成功：默认重定向到仪表盘（`/dashboard`）。如果用户因访问受保护页面而被重定向到登录页，则登录成功后应重定向回原页面。
  - 登出成功：重定向到登录页 (`/login`)。
- **"记住我"**: V1.0 初步不包含此功能，待后续评估。
- **"忘记密码"**: V1.0 初步不包含此功能，待后续评估。

## 8. HTTP 客户端配置 (例如 Axios)

- **请求拦截器**:
  - 在每个请求发送前检查状态管理中是否存在 `token`。
  - 如果存在，则将 `token`添加到 `Authorization` 请求头: `Bearer ${token}`。
- **响应拦截器**:
  - 全局处理 API 错误。
  - 特别是当 API 返回 `401 Unauthorized` (通常表示 `token` 无效或过期) 且不是发生在登录接口本身时，应自动调用状态管理中的 `logout()` action，并强制用户重新登录。

## 9. 安全性考虑

- **Token 存储**: 使用 `localStorage` 存储 JWT。虽然方便，但需注意 XSS 风险（通过其他途径缓解 XSS 是关键）。`HttpOnly` Cookies 是更安全的选择，但需要后端配合设置，且前端 JS 无法直接访问，这会改变 Token 的传递和管理方式。当前 API 设计返回 Token 在响应体中，倾向于前端存储。
- **HTTPS**: 应用必须通过 HTTPS 提供服务，以确保传输过程中的数据安全。前端在 API 请求时也必须使用`https://`协议，并且在生产环境中前端应用本身也应通过 HTTPS 部署。
- **输入校验**: 前后端都必须进行严格的输入校验。

## 10. 待决策事项

以下事项需要在开发前或开发过程中明确：

- **全局状态管理库**: **推荐并决定采用 Zustand**。（原因：轻量简洁，性能良好，与 React Hooks 及 TypeScript 集成优秀，适合管理当前项目的客户端状态如认证信息等）。
- **HTTP 客户端库**: **推荐使用 Axios**。（原因：拦截器功能强大，错误处理更直观，API 易用性高，适合本项目需求）。
- **"记住我" 功能**: **V1.0 不包含**。
- **"忘记密码" 功能**: **V1.0 不包含**。
- **登录成功后的具体重定向逻辑**: **已明确。**
  - **机制**：当用户尝试访问受保护路由但未认证时，`ProtectedRoute` (或类似组件) 应将用户重定向到 `/login`。在重定向时，应将用户原本尝试访问的路径 (e.g., `location.pathname`) 通过 `react-router-dom` 的 `state` 对象传递，例如: `navigate('/login', { state: { from: location.pathname } });`。
  - **登录页面处理**：`LoginPage.tsx` 组件应从 `location.state?.from` 读取此路径。
  - **登录成功后**：
    - 若 `location.state?.from` 存在，则导航至该路径 (e.g., `navigate(location.state.from, { replace: true });`)。
    - 若不存在（用户直接访问登录页），则导航至默认路径 (e.g., `/dashboard`，`navigate('/dashboard', { replace: true });`)。
    - 使用 `{ replace: true }` 可以防止用户通过浏览器后退按钮回到登录页。
- **`/api/v1/auth/me` 响应中是否包含用户角色信息？** **是，应包含用户角色信息。** (这对前端权限控制和初始化非常重要)。

---

此文档后续可根据决策和开发进展进行更新。
