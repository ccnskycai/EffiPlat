import React, { useState, useEffect } from 'react';
import {
  Table,
  Card,
  Button,
  Space,
  Modal,
  Popconfirm,
  message,
  Input,
  Row,
  Col,
  Select,
  Typography
} from 'antd';
import { 
  PlusOutlined,
  SearchOutlined, 
  EditOutlined,
  DeleteOutlined,
  UserSwitchOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';
import { useUserStore } from '../store/userStore';
import UserForm from '../components/users/UserForm';
import UserRoleAssignment from '../components/users/UserRoleAssignment';
import type { User, UserCreateRequest, UserUpdateRequest } from '../services/userService';

const { Title } = Typography;
const { Option } = Select;

const UsersPage: React.FC = () => {
  // 状态
  const [isFormVisible, setIsFormVisible] = useState(false);
  const [isRoleModalVisible, setIsRoleModalVisible] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [selectedUserIds, setSelectedUserIds] = useState<number[]>([]);
  const [searchName, setSearchName] = useState('');
  const [searchEmail, setSearchEmail] = useState('');
  const [searchStatus, setSearchStatus] = useState<string | undefined>(undefined);

  // 从store获取状态和方法
  const { 
    users, 
    total, 
    page, 
    pageSize, 
    loading, 
    fetchUsers,
    createUser,
    updateUser,
    deleteUser,
    assignRoles,
    batchDeleteUsers,
    searchParams,
    setSearchParams
  } = useUserStore();

  // 首次加载和参数变化时获取用户数据
  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  // 处理搜索
  const handleSearch = () => {
    setSearchParams({
      page: 1, // 重置到第一页
      name: searchName || undefined,
      email: searchEmail || undefined,
      status: searchStatus,
    });
    fetchUsers();
  };

  // 重置搜索
  const handleResetSearch = () => {
    setSearchName('');
    setSearchEmail('');
    setSearchStatus(undefined);
    setSearchParams({
      page: 1,
      name: undefined,
      email: undefined,
      status: undefined,
    });
    fetchUsers();
  };

  // 处理表格分页、排序变化
  const handleTableChange = (pagination: any, filters: any, sorter: any) => {
    setSearchParams({
      page: pagination.current,
      pageSize: pagination.pageSize,
      sortBy: sorter.field,
      order: sorter.order === 'ascend' ? 'asc' : sorter.order === 'descend' ? 'desc' : undefined
    });
    fetchUsers();
  };

  // 处理创建用户
  const handleCreateUser = async (userData: UserCreateRequest) => {
    await createUser(userData);
    setIsFormVisible(false);
    message.success('用户创建成功');
  };

  // 处理更新用户
  const handleUpdateUser = async (userData: UserUpdateRequest) => {
    if (editingUser) {
      await updateUser(editingUser.id, userData);
      setIsFormVisible(false);
      setEditingUser(null);
      message.success('用户更新成功');
    }
  };

  // 打开编辑表单
  const openEditForm = (user: User) => {
    setEditingUser(user);
    setIsFormVisible(true);
  };

  // 关闭表单
  const closeForm = () => {
    setIsFormVisible(false);
    setEditingUser(null);
  };

  // 处理删除用户
  const handleDeleteUser = async (userId: number) => {
    try {
      await deleteUser(userId);
      message.success('用户删除成功');
    } catch (error) {
      message.error('删除用户失败');
    }
  };

  // 打开角色分配模态框
  const openRoleAssignment = (user: User) => {
    setEditingUser(user);
    setIsRoleModalVisible(true);
  };

  // 处理角色分配
  const handleAssignRoles = async (userId: number, roleIds: number[]) => {
    await assignRoles(userId, roleIds);
    setIsRoleModalVisible(false);
    setEditingUser(null);
    message.success('角色分配成功');
  };

  // 批量删除用户
  const handleBatchDelete = async () => {
    Modal.confirm({
      title: '确认删除所选用户?',
      icon: <ExclamationCircleOutlined />,
      content: '此操作不可撤销，确定要删除选择的用户吗？',
      okText: '确认',
      cancelText: '取消',
      onOk: async () => {
        try {
          // 调用我们实现的批量删除方法
          await batchDeleteUsers(selectedUserIds);
          message.success('批量删除成功');
          setSelectedUserIds([]);
        } catch (error) {
          message.error('批量删除失败');
          console.error('批量删除错误:', error);
        }
      },
    });
  };

  // 表格列定义
  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      sorter: true,
    },
    {
      title: '用户名',
      dataIndex: 'name',
      key: 'name',
      sorter: true,
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
      sorter: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      sorter: true,
      render: (status: string) => {
        let color;
        let text;
        switch (status) {
          case 'active':
            color = 'green';
            text = '激活';
            break;
          case 'inactive':
            color = 'orange';
            text = '未激活';
            break;
          case 'locked':
            color = 'red';
            text = '锁定';
            break;
          default:
            color = 'default';
            text = status;
        }
        return <span style={{ color }}>{text}</span>;
      }
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      sorter: true,
      render: (date: string) => new Date(date).toLocaleString()
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: User) => (
        <Space size="small">
          <Button 
            type="link" 
            icon={<EditOutlined />} 
            onClick={() => openEditForm(record)}
          >
            编辑
          </Button>
          <Button
            type="link"
            icon={<UserSwitchOutlined />}
            onClick={() => openRoleAssignment(record)}
          >
            角色
          </Button>
          <Popconfirm
            title="确认删除此用户?"
            onConfirm={() => handleDeleteUser(record.id)}
            okText="确认"
            cancelText="取消"
          >
            <Button type="link" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 表格行选择配置
  const rowSelection = {
    selectedRowKeys: selectedUserIds,
    onChange: (selectedRowKeys: React.Key[]) => {
      setSelectedUserIds(selectedRowKeys.map(key => Number(key)));
    },
  };

  return (
    <div style={{ padding: '20px' }}>
      <Title level={2}>用户管理</Title>
      
      {/* 搜索区域 */}
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col span={6}>
            <Input
              placeholder="按用户名搜索"
              value={searchName}
              onChange={e => setSearchName(e.target.value)}
              prefix={<SearchOutlined />}
            />
          </Col>
          <Col span={6}>
            <Input
              placeholder="按邮箱搜索"
              value={searchEmail}
              onChange={e => setSearchEmail(e.target.value)}
              prefix={<SearchOutlined />}
            />
          </Col>
          <Col span={6}>
            <Select
              placeholder="按状态筛选"
              style={{ width: '100%' }}
              value={searchStatus}
              onChange={setSearchStatus}
              allowClear
            >
              <Option value="active">激活</Option>
              <Option value="inactive">未激活</Option>
              <Option value="locked">锁定</Option>
            </Select>
          </Col>
          <Col span={6}>
            <Space>
              <Button type="primary" onClick={handleSearch}>
                搜索
              </Button>
              <Button onClick={handleResetSearch}>
                重置
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>
      
      {/* 操作按钮区域 */}
      <Card style={{ marginBottom: 16 }}>
        <Space>
          <Button 
            type="primary" 
            icon={<PlusOutlined />}
            onClick={() => setIsFormVisible(true)}
          >
            创建用户
          </Button>
          <Button 
            danger 
            disabled={selectedUserIds.length === 0}
            onClick={handleBatchDelete}
          >
            批量删除
          </Button>
        </Space>
      </Card>
      
      {/* 用户表格 */}
      <Card>
        <Table
          rowKey="id"
          rowSelection={rowSelection}
          columns={columns}
          dataSource={users}
          loading={loading}
          pagination={{
            current: page,
            pageSize: pageSize,
            total: total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
          onChange={handleTableChange}
        />
      </Card>
      
      {/* 用户表单模态框 */}
      <Modal
        title={editingUser ? '编辑用户' : '创建用户'}
        open={isFormVisible}
        onCancel={closeForm}
        footer={null}
        destroyOnClose
      >
        <UserForm
          user={editingUser || undefined}
          isEdit={!!editingUser}
          loading={loading}
          onSubmit={editingUser ? handleUpdateUser : handleCreateUser}
          onCancel={closeForm}
        />
      </Modal>
      
      {/* 角色分配模态框 */}
      <Modal
        title="用户角色分配"
        open={isRoleModalVisible}
        onCancel={() => {
          setIsRoleModalVisible(false);
          setEditingUser(null);
        }}
        footer={null}
        width={700}
        destroyOnClose
      >
        {editingUser && (
          <UserRoleAssignment
            userId={editingUser.id}
            onSave={handleAssignRoles}
            onCancel={() => {
              setIsRoleModalVisible(false);
              setEditingUser(null);
            }}
          />
        )}
      </Modal>
    </div>
  );
};

export default UsersPage;
