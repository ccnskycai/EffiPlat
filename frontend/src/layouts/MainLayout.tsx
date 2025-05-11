import React, { useState } from 'react';
import { Layout, Menu, Avatar, Input, Typography, Space } from 'antd';
import { Link, useLocation, Location, Outlet, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import {
  DashboardOutlined,
  TeamOutlined,
  LogoutOutlined,
  SearchOutlined,
  BellOutlined,
  SettingOutlined,
  BugOutlined,
  DatabaseOutlined,
  ClusterOutlined,
} from '@ant-design/icons';
import effiplatLogo from '../assets/effiplat-logo.png';

const { Header, Sider, Content, Footer } = Layout;
const { Title, Text } = Typography;

const menuItems = [
  {
    key: 'dashboard',
    icon: <DashboardOutlined />,
    label: '仪表盘',
    path: '/dashboard',
  },
  {
    key: 'users',
    icon: <TeamOutlined />,
    label: '用户管理',
    path: '/users',
  },
  {
    key: 'environments',
    icon: <SettingOutlined />,
    label: '环境管理',
    path: '/environments',
  },
  {
    key: 'assets',
    icon: <DatabaseOutlined />,
    label: '资产管理',
    path: '/assets',
  },
  {
    key: 'services',
    icon: <ClusterOutlined />,
    label: '服务管理',
    path: '/services',
  },
  {
    key: 'bugs',
    icon: <BugOutlined />,
    label: 'Bug管理',
    path: '/bugs',
  },
];

const MainLayout: React.FC = () => {
  const location: Location = useLocation();
  const navigate = useNavigate();
  const logout = useAuthStore((state) => state.logout);
  const [collapsed, setCollapsed] = useState(false);
  
  const handleLogout = async () => {
    try {
      await logout();
      navigate('/login');
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  const getCurrentPageTitle = () => {
    const currentTopLevelPath = location.pathname.split('/')[1] || 'dashboard';
    const currentItem = menuItems.find((item) => item.key === currentTopLevelPath);
    return currentItem ? currentItem.label : '仪表盘';
  };

  const defaultSelectedKey = location.pathname.split('/')[1] || 'dashboard';

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        breakpoint="lg"
        collapsedWidth="80"
        width={260}
        collapsible
        collapsed={collapsed}
        onCollapse={(value) => setCollapsed(value)}
        theme="light"
        style={{
          overflow: 'auto',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
          borderRight: '1px solid #f0f0f0',
        }}
      >
        <div
          style={{
            height: 64,
            margin: '16px 0',
            display: 'flex',
            alignItems: 'center',
            justifyContent: collapsed ? 'center' : 'flex-start',
            paddingLeft: collapsed ? 0 : '24px',
          }}
        >
          <img
            src={effiplatLogo}
            alt="EffiPlat Logo"
            style={{
              height: 32,
              marginRight: collapsed ? 0 : 8,
            }}
          />
          <Title
            level={3}
            style={{
              color: '#1890ff',
              margin: 0,
              display: collapsed ? 'none' : 'block',
            }}
          >
            EffiPlat
          </Title>
        </div>
        <Menu
          theme="light"
          mode="inline"
          defaultSelectedKeys={[defaultSelectedKey]}
          style={{ borderRight: 0 }}
          items={menuItems.map((item) => ({
            key: item.key,
            icon: item.icon,
            label: <Link to={item.path}>{item.label}</Link>,
          }))}
        />
        <div
          style={{
            position: 'absolute',
            bottom: 60,
            width: '100%',
            textAlign: collapsed ? 'center' : 'left',
            padding: collapsed ? '0' : '0 24px',
          }}
        >
          <Menu
            theme="light"
            mode={collapsed ? 'vertical' : 'inline'}
            selectable={false}
            style={{ borderRight: 0 }}
            items={[
              {
                key: 'signout',
                icon: <LogoutOutlined />,
                label: !collapsed ? <span onClick={handleLogout} style={{ cursor: 'pointer' }}>Sign Out</span> : '',
                onClick: handleLogout,
              },
            ]}
          />
        </div>
      </Sider>
      <Layout
        style={{
          marginLeft: collapsed ? 80 : 260,
          transition: 'margin-left 0.2s',
        }}
      >
        <Header
          style={{
            background: '#fff',
            padding: '0 24px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: '1px solid #f0f0f0',
          }}
        >
          <Title level={4} style={{ margin: 0 }}>
            {getCurrentPageTitle()}
          </Title>
          <Space size="middle">
            <Input
              prefix={<SearchOutlined />}
              placeholder="Search anything"
              style={{ width: 250 }}
            />
            <BellOutlined style={{ fontSize: 20 }} />
            <div style={{ display: 'flex', alignItems: 'center' }}>
              <Avatar src="https://joeschmoe.io/api/v1/random" />
              <div style={{ marginLeft: 8, display: 'flex', flexDirection: 'column' }}>
                <Text strong ellipsis={{ tooltip: 'Phillip Stanton' }}>
                  Phillip Stanton
                </Text>
                <Text
                  type="secondary"
                  style={{ fontSize: 12, lineHeight: '1' }}
                  ellipsis={{ tooltip: 'Admin' }}
                >
                  Admin
                </Text>
              </div>
            </div>
            <SettingOutlined style={{ fontSize: 20 }} />
          </Space>
        </Header>
        <Content
          style={{
            margin: '24px 16px',
            padding: 24,
            background: '#f0f2f5',
            minHeight: 'calc(100vh - 64px - 48px - 69px)',
          }}
        >
          <Outlet />
        </Content>
        <Footer
          style={{
            textAlign: 'center',
            background: '#fff',
            borderTop: '1px solid #f0f0f0',
            padding: '12px 50px',
          }}
        >
          {/* Removed Copyright Text component and its links */}
          {/* 
          <Text type="secondary">
            Copyright © 2025 Peterdraw.
            <Link to="/privacy" style={{ marginLeft: 8 }}>
              Privacy policy
            </Link>
            <Link to="/terms" style={{ marginLeft: 8 }}>
              Terms and conditions
            </Link>
            <Link to="/contact" style={{ marginLeft: 8 }}>
              Contact
            </Link>
          </Text>
          */}
        </Footer>
      </Layout>
    </Layout>
  );
};

export default MainLayout;
