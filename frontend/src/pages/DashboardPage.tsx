import React from 'react';
import { Card, Row, Col, Typography, Statistic } from 'antd';
import { useAuthStore } from '../store/authStore';
import {
  DashboardOutlined,
  TeamOutlined,
  SettingOutlined,
  DatabaseOutlined,
} from '@ant-design/icons';

const { Title } = Typography;

const DashboardPage: React.FC = () => {
  const user = useAuthStore((state) => state.user);

  return (
    <div style={{ padding: '20px' }}>
      <Title level={4} style={{ marginBottom: 24 }}>
        欢迎, {user?.name || '用户'}!
      </Title>

      {/* 统计卡片 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="用户总数"
              value={1234}
              prefix={<TeamOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="环境总数"
              value={56}
              prefix={<SettingOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="资产总数"
              value={789}
              prefix={<DatabaseOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="活跃服务"
              value={42}
              prefix={<DashboardOutlined />}
              valueStyle={{ color: '#eb2f96' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 用户信息卡片 */}
      <Card title="用户信息" style={{ marginTop: 24 }}>
        <p>
          <strong>用户 ID:</strong> {user?.id}
        </p>
        <p>
          <strong>用户名:</strong> {user?.name}
        </p>
        <p>
          <strong>邮箱:</strong> {user?.email || '未设置'}
        </p>
        <p>
          <strong>角色:</strong> {user?.roles?.join(', ') || '普通用户'}
        </p>
      </Card>
    </div>
  );
};

export default DashboardPage;
