import React, { useEffect } from 'react';
import { Form, Input, Select, Button, Space, message } from 'antd';
import { User, UserCreateRequest, UserUpdateRequest } from '../../services/userService';

const { Option } = Select;

interface UserFormProps {
  user?: User;
  isEdit?: boolean;
  loading?: boolean;
  onSubmit: (values: UserCreateRequest | UserUpdateRequest) => Promise<void>;
  onCancel: () => void;
}

const UserForm: React.FC<UserFormProps> = ({
  user,
  isEdit = false,
  loading = false,
  onSubmit,
  onCancel
}) => {
  const [form] = Form.useForm();

  // 初始化表单值
  useEffect(() => {
    if (user && isEdit) {
      form.setFieldsValue({
        name: user.name,
        email: user.email,
        status: user.status,
      });
    } else {
      form.resetFields();
    }
  }, [form, user, isEdit]);

  // 提交表单
  const handleSubmit = async (values: UserCreateRequest | UserUpdateRequest) => {
    try {
      await onSubmit(values);
      message.success(`${isEdit ? '更新' : '创建'}用户成功`);
      form.resetFields();
    } catch (error) {
      message.error(`${isEdit ? '更新' : '创建'}用户失败: ${error instanceof Error ? error.message : '未知错误'}`);
    }
  };

  return (
    <Form
      form={form}
      layout="vertical"
      onFinish={handleSubmit}
      initialValues={{ status: 'active' }}
      requiredMark
    >
      <Form.Item
        name="name"
        label="用户名"
        rules={[{ required: true, message: '请输入用户名' }]}
      >
        <Input placeholder="请输入用户名" />
      </Form.Item>

      <Form.Item
        name="email"
        label="邮箱"
        rules={[
          { required: true, message: '请输入邮箱' },
          { type: 'email', message: '请输入有效的邮箱地址' }
        ]}
      >
        <Input placeholder="请输入邮箱" />
      </Form.Item>

      {!isEdit && (
        <Form.Item
          name="password"
          label="密码"
          rules={[
            { required: true, message: '请输入密码' },
            { min: 8, message: '密码长度至少为8个字符' }
          ]}
        >
          <Input.Password placeholder="请输入密码" />
        </Form.Item>
      )}

      {isEdit && (
        <Form.Item
          name="password"
          label="密码 (留空保持不变)"
          rules={[
            { min: 8, message: '密码长度至少为8个字符' }
          ]}
        >
          <Input.Password placeholder="留空则不修改密码" />
        </Form.Item>
      )}

      <Form.Item
        name="status"
        label="状态"
        rules={[{ required: true, message: '请选择用户状态' }]}
      >
        <Select placeholder="请选择用户状态">
          <Option value="active">激活</Option>
          <Option value="inactive">未激活</Option>
          <Option value="locked">锁定</Option>
        </Select>
      </Form.Item>

      <Form.Item>
        <Space>
          <Button type="primary" htmlType="submit" loading={loading}>
            {isEdit ? '更新' : '创建'}
          </Button>
          <Button onClick={onCancel}>取消</Button>
        </Space>
      </Form.Item>
    </Form>
  );
};

export default UserForm;
