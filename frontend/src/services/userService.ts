import apiClient from './apiClient';
import { isAxiosError } from 'axios';

// 定义用户相关类型
// 角色接口定义
export interface Role {
  id: number;
  name: string;
  description?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface User {
  id: number;
  name: string;
  email: string;
  status: string;
  createdAt: string;
  updatedAt: string;
  roles?: Role[] | number[] | string[] | null;
  department?: string | null;
}

export interface UserCreateRequest {
  name: string;
  email: string;
  password: string;
  status?: string;
}

export interface UserUpdateRequest {
  name?: string;
  email?: string;
  password?: string;
  status?: string;
}

export interface UserListParams {
  page?: number;
  pageSize?: number;
  name?: string;
  email?: string;
  status?: string;
  sortBy?: string;
  order?: 'asc' | 'desc';
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

// 用户角色相关接口
export interface RoleAssignRequest {
  role_ids: number[];
}

// 用户服务类
class UserService {
  private baseUrl = '/api/v1';

  // 获取用户列表
  async getUsers(params: UserListParams = {}): Promise<ApiResponse<PaginatedResponse<User>>> {
    console.log('[userService] 获取用户列表，请求URL:', `${this.baseUrl}/users`, '参数:', params);
    try {
      const response = await apiClient.get(`${this.baseUrl}/users`, { params });
      console.log('[userService] 获取用户列表成功，响应:', response);
      return response.data;
    } catch (error) {
      console.error('[userService] 获取用户列表失败:', error);
      if (isAxiosError(error)) {
        console.error('[userService] 请求详情:', {
          status: error.response?.status,
          statusText: error.response?.statusText,
          data: error.response?.data,
          headers: error.response?.headers,
          url: error.config?.url,
          method: error.config?.method,
        });
      }
      throw error;
    }
  }

  // 获取单个用户详情
  async getUserById(userId: number): Promise<ApiResponse<User>> {
    const response = await apiClient.get(`${this.baseUrl}/users/${userId}`);
    return response.data;
  }

  // 创建用户
  async createUser(userData: UserCreateRequest): Promise<ApiResponse<User>> {
    const response = await apiClient.post(`${this.baseUrl}/users`, userData);
    return response.data;
  }

  // 更新用户
  async updateUser(userId: number, userData: UserUpdateRequest): Promise<ApiResponse<User>> {
    const response = await apiClient.put(`${this.baseUrl}/users/${userId}`, userData);
    return response.data;
  }

  // 删除用户
  async deleteUser(userId: number): Promise<void> {
    await apiClient.delete(`${this.baseUrl}/users/${userId}`);
  }

  // 批量删除用户
  async batchDeleteUsers(userIds: number[]): Promise<void> {
    await apiClient.delete(`${this.baseUrl}/users/batch`, {
      data: { ids: userIds }
    });
  }

  // 为用户分配角色
  async assignRolesToUser(userId: number, roleIds: number[]): Promise<ApiResponse<User>> {
    const response = await apiClient.post(`${this.baseUrl}/users/${userId}/roles`, {
      role_ids: roleIds
    });
    return response.data;
  }

  // 从用户移除角色
  async removeRolesFromUser(userId: number, roleIds: number[]): Promise<ApiResponse<User>> {
    const response = await apiClient.delete(`${this.baseUrl}/users/${userId}/roles`, {
      data: { role_ids: roleIds }
    });
    return response.data;
  }
}

export const userService = new UserService();
export default userService;
