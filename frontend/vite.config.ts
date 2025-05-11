import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000, // 你可以指定开发服务器端口
    open: true, // 自动在浏览器中打开
    proxy: {
      // 将所有 /api 开头的请求代理到后端服务
      '/api': {
        target: 'http://localhost:8080', // 后端服务地址
        changeOrigin: true,
        // 如果后端不需要 /api 前缀，可以使用 rewrite
        // rewrite: (path) => path.replace(/^\/api/, '')
      },
    },
  },
});
