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

// 引入Role接口
import { Role } from '../services/userService';

// 辅助函数：处理不同格式的角色数据
const renderRoles = (roles?: Array<Role | string | number> | null) => {
  if (!roles || roles.length === 0) {
    return '普通用户';
  }

  if (!Array.isArray(roles)) {
    return '未知角色';
  }

  // 处理不同类型的角色数据
  if (typeof roles[0] === 'string') {
    // 字符串数组
    return roles.join(', ');
  } else if (typeof roles[0] === 'object' && roles[0] !== null) {
    // 对象数组 - 首先将对象转换为字符串数组，然后再连接
    const roleNames: string[] = roles.map((role) => {
      const roleObj = role as Role;
      return roleObj.name || roleObj.id?.toString() || 'unknown';
    });
    
    return roleNames.join(', ');
  } else {
    // 数字数组或其他
    const stringArray = roles.map((item) => {
      // 确保正确处理可能的对象或基本类型
      if (typeof item === 'object' && item !== null) {
        return 'unknown';
      }
      return String(item || '');
    });
    return stringArray.join(', ');
  }
};

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
          <strong>角色:</strong> {renderRoles(user?.roles)}
        </p>
      </Card>
    </div>
  );
};

export default DashboardPage;
