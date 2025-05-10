import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000, // 你可以指定开发服务器端口
    open: true, // 自动在浏览器中打开
  },
});
