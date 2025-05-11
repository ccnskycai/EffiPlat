import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom'; // Assuming react-router-dom for navigation
import { useAuthStore } from '../store/authStore';
import LoginForm from '../components/auth/LoginForm';
import { Card, Layout, Row, Col, Typography } from 'antd';
import logoPath from '../assets/effiplat-logo.png';

const { Title } = Typography;
const { Content } = Layout;

const LoginPage: React.FC = () => {
  const navigate = useNavigate();
  // Use individual selector with proper typing
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const authError = useAuthStore((state) => state.error);
  const clearError = useAuthStore((state) => state.clearError);

  useEffect(() => {
    // If user is already authenticated, redirect to dashboard or home page
    if (isAuthenticated) {
      // TODO: Make '/dashboard' a configurable path or import from a constants file
      void navigate('/dashboard', { replace: true }); // Added void to address no-floating-promises
    }
    // Clear any existing auth errors when the login page loads
    // This prevents a stale error message from a previous attempt from showing immediately
    if (authError) {
      clearError();
    }
  }, [isAuthenticated, navigate, authError, clearError]);

  // If already authenticated, render nothing or a loading indicator while redirecting
  if (isAuthenticated) {
    return null; // Or <LoadingIndicator />
  }

  return (
    <Layout style={{ minHeight: '100vh', background: '#f0f2f5' /* Or your theme's bg color */ }}>
      <Content style={{ display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
        <Row justify="center" align="middle" style={{ width: '100%' }}>
          <Col xs={22} sm={16} md={12} lg={8} xl={6}>
            <Card
              variant="outlined"
              style={{
                boxShadow: '0 4px 8px 0 rgba(0,0,0,0.2)',
                borderRadius: '8px',
              }}
            >
              <div style={{ textAlign: 'center', marginBottom: '24px' }}>
                {/* You can add a logo here */}
                <img
                  src={logoPath}
                  alt="EffiPlat Logo"
                  style={{ height: '40px', marginBottom: '10px' }}
                />
                <Title level={2} style={{ color: '#1890ff' /* Ant Design primary color */ }}>
                  EffiPlat Login
                </Title>
              </div>
              <LoginForm />
            </Card>
          </Col>
        </Row>
      </Content>
    </Layout>
  );
};

export default LoginPage;
