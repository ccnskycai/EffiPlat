import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import MainLayout from './layouts/MainLayout';
import HomePage from './pages/HomePage';
import AboutPage from './pages/AboutPage';

function App() {
  return (
    <Router>
      <MainLayout>
        <nav className="p-4 bg-gray-200">
          <ul className="flex space-x-4">
            <li>
              <Link to="/" className="text-blue-500 hover:text-blue-700">
                首页
              </Link>
            </li>
            <li>
              <Link to="/about" className="text-blue-500 hover:text-blue-700">
                关于
              </Link>
            </li>
          </ul>
        </nav>
        <div className="p-4">
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/about" element={<AboutPage />} />
          </Routes>
        </div>
      </MainLayout>
    </Router>
  );
}

export default App;
