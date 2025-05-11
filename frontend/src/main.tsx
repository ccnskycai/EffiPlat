import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import 'antd/dist/reset.css'; // Ant Design v5+ CSS reset
// import './index.css'; // 移除这一行

// Temporarily remove StrictMode to solve the infinite loop issue
// This is a known issue with zustand and React.StrictMode in development
ReactDOM.createRoot(document.getElementById('root')!).render(
  <App />
);
