import React, { useState, useEffect } from 'react';
import { Transfer, Button, Space, message, Spin, Typography } from 'antd';
import axios from 'axios';

interface Role {
  id: number;
  name: string;
  description?: string;
}

interface UserRoleAssignmentProps {
  userId: number;
  onSave: (userId: number, roleIds: number[]) => Promise<void>;
  onCancel: () => void;
}

const UserRoleAssignment: React.FC<UserRoleAssignmentProps> = ({
  userId,
  onSave,
  onCancel
}) => {
  const [allRoles, setAllRoles] = useState<Role[]>([]);
  const [assignedRoleIds, setAssignedRoleIds] = useState<number[]>([]);
  const [loading, setLoading] = useState(false);
  const [targetKeys, setTargetKeys] = useState<string[]>([]);

  // 获取所有角色和用户已分配的角色
  useEffect(() => {
    const fetchRolesData = async () => {
      setLoading(true);
      try {
        // 获取所有角色
        const rolesResponse = await axios.get('/api/v1/roles');
        console.log('角色列表响应:', rolesResponse);
        const roles: Role[] = rolesResponse.data.data.items;
        setAllRoles(roles);

        // 获取用户详情，包含已分配角色
        const userResponse = await axios.get(`/api/v1/users/${userId}`);
        console.log('用户详情响应:', userResponse);
        const userData = userResponse.data.data;
        
        // 检查用户是否有roles字段
        if (userData.roles) {
          // 根据后端返回的roles格式处理
          // 如果roles是对象数组
          if (Array.isArray(userData.roles) && userData.roles.length > 0 && typeof userData.roles[0] === 'object') {
            const assignedIds = userData.roles.map((role: Role) => role.id);
            setAssignedRoleIds(assignedIds);
            setTargetKeys(assignedIds.map(id => id.toString()));
          } 
          // 如果roles是ID数组
          else if (Array.isArray(userData.roles) && userData.roles.length > 0 && typeof userData.roles[0] === 'number') {
            setAssignedRoleIds(userData.roles as number[]);
            setTargetKeys(userData.roles.map(id => id.toString()));
          }
          // 如果roles是字符串数组(可能是角色名)
          else if (Array.isArray(userData.roles) && userData.roles.length > 0 && typeof userData.roles[0] === 'string') {
            // 暂时只设置空的已分配角色，待角色名与ID匹配后再处理
            setAssignedRoleIds([]);
            setTargetKeys([]);
          }
        }
      } catch (error) {
        message.error('获取角色数据失败');
        console.error('Error fetching roles data:', error);
      } finally {
        setLoading(false);
      }
    };

    if (userId) {
      fetchRolesData();
    }
  }, [userId]);

  // 处理角色选择变化
  const handleChange = (nextTargetKeys: string[]) => {
    setTargetKeys(nextTargetKeys);
  };

  // 处理保存
  const handleSave = async () => {
    const selectedRoleIds = targetKeys.map(key => parseInt(key, 10));
    try {
      await onSave(userId, selectedRoleIds);
      message.success('角色分配成功');
    } catch (error) {
      message.error('角色分配失败');
      console.error('Error saving roles:', error);
    }
  };

  return (
    <div>
      <Typography.Title level={5}>为用户分配角色</Typography.Title>
      <Typography.Paragraph>
        从左侧选择角色添加到右侧，或从右侧移除角色。
      </Typography.Paragraph>

      <Spin spinning={loading}>
        <Transfer
          dataSource={allRoles.map(role => ({
            key: role.id.toString(),
            title: role.name,
            description: role.description || '',
          }))}
          targetKeys={targetKeys}
          onChange={handleChange}
          render={item => item.title}
          listStyle={{
            width: 250,
            height: 300,
          }}
          titles={['可用角色', '已分配角色']}
          showSearch
        />
      </Spin>

      <div style={{ marginTop: 16 }}>
        <Space>
          <Button type="primary" onClick={handleSave} disabled={loading}>
            保存
          </Button>
          <Button onClick={onCancel} disabled={loading}>
            取消
          </Button>
        </Space>
      </div>
    </div>
  );
};

export default UserRoleAssignment;
