import React from 'react';
import { Form, Input, Button, Alert } from 'antd';
import { MailOutlined, LockOutlined } from '@ant-design/icons';
import { useAuthStore } from '../../store/authStore';
import { shallow } from 'zustand/shallow';
import { LoginCredentials } from '../../services/authService';

const LoginForm: React.FC = () => {
  // Use individual selectors with proper typing
  const login = useAuthStore((state) => state.login);
  const isLoading = useAuthStore((state) => state.isLoading);
  const authError = useAuthStore((state) => state.error);

  const [form] = Form.useForm<LoginCredentials>();

  // Actual async logic for form submission
  const handleFormSubmit = async (values: LoginCredentials) => {
    try {
      await login(values);
      // Redirect is handled by LoginPage or globally based on isAuthenticated
      // form.resetFields(); // Optionally reset fields after successful login attempt
    } catch (error) {
      // Error is already set in the authStore, Alert component will display it.
      // No need to set form-specific errors here unless the API returns field-specific validation messages
      console.error('Login attempt failed in component:', error);
    }
  };

  // Wrapper for onFinish prop to satisfy linter rule
  const onFinishWrapper = (values: LoginCredentials) => {
    void handleFormSubmit(values);
  };

  return (
    <Form
      form={form}
      name="login_form"
      initialValues={{
        email: 'testuser1@example.com',
        password: 'password',
        remember: true,
      }}
      onFinish={onFinishWrapper}
      layout="vertical"
      requiredMark="optional"
    >
      {authError && (
        <Form.Item>
          <Alert message={authError} type="error" showIcon />
        </Form.Item>
      )}

      <Form.Item
        name="email"
        label="Email"
        rules={[
          {
            required: true,
            message: 'Please input your Email!',
          },
          {
            type: 'email',
            message: 'The input is not valid E-mail!',
          },
        ]}
      >
        <Input prefix={<MailOutlined className="site-form-item-icon" />} placeholder="Email" />
      </Form.Item>

      <Form.Item
        name="password"
        label="Password"
        rules={[{ required: true, message: 'Please input your Password!' }]}
      >
        <Input.Password
          prefix={<LockOutlined className="site-form-item-icon" />}
          placeholder="Password"
        />
      </Form.Item>

      {/* Add remember me or forgot password later if needed as per design evolution */}
      {/* 
      <Form.Item>
        <Form.Item name="remember" valuePropName="checked" noStyle>
          <Checkbox>Remember me</Checkbox>
        </Form.Item>
        <a className="login-form-forgot" href="">
          Forgot password
        </a>
      </Form.Item>
      */}

      <Form.Item>
        <Button type="primary" htmlType="submit" loading={isLoading} style={{ width: '100%' }}>
          Log in
        </Button>
      </Form.Item>
    </Form>
  );
};

export default LoginForm;
