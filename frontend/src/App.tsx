import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useLocation, Outlet } from 'react-router-dom';
import { useAuthStore } from './store/authStore';
import LoginPage from './pages/LoginPage';
import DashboardPage from './pages/DashboardPage';
import UsersPage from './pages/UsersPage';
import MainLayout from './layouts/MainLayout';
import { Spin } from 'antd'; // For loading indicator

// Component to handle root and catch-all redirects
const RootRedirect: React.FC = () => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  return <Navigate to={isAuthenticated ? 'dashboard' : '/login'} replace />;
};

// A wrapper for protected routes
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  // Use individual selectors with proper typing
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const isLoadingAuth = useAuthStore((state) => state.isLoading);
  const location = useLocation();

  if (isLoadingAuth) {
    // Show a global loading spinner while checking auth status
    return (
      <div
        style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}
      >
        <Spin size="large" />
      </div>
    );
  }

  if (!isAuthenticated) {
    // Redirect them to the /login page, but save the current location they were
    // trying to go to when they were redirected. This allows us to send them
    // along to that page after they login, which is a nicer user experience
    // than dropping them off on the home page.
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return children;
};

function App() {
  // Use selectors with the updated auth store pattern
  const initializeAuth = useAuthStore((state) => state.initializeAuth);
  const appIsLoading = useAuthStore((state) => state.isLoading);
  const token = useAuthStore((state) => state.token);
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  useEffect(() => {
    void initializeAuth();
  }, [initializeAuth]);

  // Potentially show a full-screen loader while initializeAuth is running for the very first time.
  // The isLoading in ProtectedRoute handles subsequent navigations or if initializeAuth is slow.
  // This specific isLoading might be true during the very first check.
  // This logic can be refined based on how appIsLoading behaves during initializeAuth.
  if (appIsLoading && !token && !isAuthenticated) {
    // A more specific check for initial load before token is even checked
    return (
      <div
        style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}
      >
        <Spin size="large" tip="Initializing..." fullscreen />
      </div>
    );
  }

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />

        {/* Protected routes using MainLayout */}
        <Route
          element={
            <ProtectedRoute>
              <MainLayout />
            </ProtectedRoute>
          }
        >
          {/* Default route for authenticated users landing at root of protected section */}
          {/* This effectively makes '/' (within protected context) go to '/dashboard' */}
          <Route index element={<Navigate to="dashboard" replace />} />
          <Route path="dashboard" element={<DashboardPage />} />
          
          {/* Added placeholder components for menu items to prevent navigation loops */}
          <Route path="users" element={<UsersPage />} />
          <Route path="environments" element={<div>Environment Management Page (Coming Soon)</div>} />
          <Route path="assets" element={<div>Asset Management Page (Coming Soon)</div>} />
          <Route path="services" element={<div>Service Management Page (Coming Soon)</div>} />
          <Route path="bugs" element={<div>Bug Management Page (Coming Soon)</div>} />
        </Route>

        {/* Fallback for unauthenticated users trying to access root or any other non-login path before being caught by ProtectedRoute */}
        {/* If ProtectedRoute redirects unauthenticated users to /login, this might only catch paths not covered by any route. */}
        {/* A more specific catch-all for truly undefined paths might be <Route path="*" element={<NotFoundPage />} /> or redirect to login/dashboard */}
        <Route
          path="/"
          element={<RootRedirect />}
        />
        <Route
          path="*"
          element={<RootRedirect />}
        />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
