# 前端客户端技术选型

本文档整合自 `technology_stack.md` 中关于前端客户端的技术选型部分。

*   **框架/库**: [**Next.js** (初步选定)]
    *   *理由*: 基于 React 的流行框架，支持 SSR/SSG/ISR 提高首屏加载速度和 SEO（虽然 SEO 在内网系统不关键，但框架本身能力全面），自带路由和构建优化，开发体验好。
*   **编程语言**: [**TypeScript** (初步选定)]
    *   *理由*: 提供静态类型检查，提高代码健壮性和可维护性，与现代前端生态结合良好。
*   **UI 组件库**: [**Ant Design** 或 **Material UI (MUI)** (待定)]
    *   *理由*: 两者都提供丰富、高质量、遵循设计规范的 UI 组件，能极大加速前端开发。Ant Design 国内使用广泛，文档友好；MUI 遵循 Google Material Design。根据团队偏好和项目风格选择。
*   **状态管理**: [**Zustand** 或 **Redux Toolkit** (待定)]
    *   *理由*: Zustand 简单轻量，API 友好；Redux Toolkit 是官方推荐的 Redux 使用方式，功能强大，生态成熟。根据项目复杂度和团队经验选择。
*   **数据请求**: [**axios** 或 内建 `fetch` 结合 **SWR/React Query** (待定)]
    *   *理由*: SWR/React Query 能很好地处理数据缓存、重新验证、状态同步等。
*   **构建工具**: Next.js 内建 (基于 Webpack/Turbopack)。