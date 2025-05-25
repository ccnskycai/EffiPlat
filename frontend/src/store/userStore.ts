import { create } from 'zustand';
import { 
  userService, 
  User, 
  UserCreateRequest, 
  UserUpdateRequest,
  UserListParams,
  PaginatedResponse
} from '../services/userService';

interface UserState {
  users: User[];
  currentUser: User | null;
  total: number;
  page: number;
  pageSize: number;
  loading: boolean;
  error: string | null;
  
  // 查询参数
  searchParams: UserListParams;
  
  // 操作方法
  fetchUsers: (params?: UserListParams) => Promise<void>;
  fetchUserById: (userId: number) => Promise<void>;
  createUser: (userData: UserCreateRequest) => Promise<User | null>;
  updateUser: (userId: number, userData: UserUpdateRequest) => Promise<User | null>;
  deleteUser: (userId: number) => Promise<void>;
  batchDeleteUsers: (userIds: number[]) => Promise<void>;
  assignRoles: (userId: number, roleIds: number[]) => Promise<void>;
  removeRoles: (userId: number, roleIds: number[]) => Promise<void>;
  setSearchParams: (params: Partial<UserListParams>) => void;
  resetState: () => void;
}

export const useUserStore = create<UserState>((set, get) => ({
  users: [],
  currentUser: null,
  total: 0,
  page: 1,
  pageSize: 10,
  loading: false,
  error: null,
  
  searchParams: {
    page: 1,
    pageSize: 10,
  },
  
  // 获取用户列表
  fetchUsers: async (params?: UserListParams) => {
    try {
      set({ loading: true, error: null });
      
      // 合并搜索参数
      const searchParams = params 
        ? { ...get().searchParams, ...params } 
        : get().searchParams;
      
      console.log('正在获取用户列表，参数:', searchParams);
      const response = await userService.getUsers(searchParams);
      console.log('获取用户列表响应:', response);
      
      // 检查响应格式
      if (!response.data) {
        throw new Error('API响应格式错误: 缺少data字段');
      }
      
      // 从response.data中提取数据，注意可能的字段名差异
      const items = response.data.items || response.data.users || [];
      const total = response.data.total || 0;
      const page = response.data.page || params?.page || get().page;
      const pageSize = response.data.pageSize || params?.pageSize || get().pageSize;
      
      console.log('解析后的数据:', { items, total, page, pageSize });
      
      set({ 
        users: items, 
        total, 
        page, 
        pageSize,
        loading: false,
        searchParams
      });
    } catch (error) {
      console.error('获取用户列表失败:', error);
      set({ 
        loading: false, 
        error: error instanceof Error ? error.message : '获取用户列表失败' 
      });
    }
  },
  
  // 获取单个用户
  fetchUserById: async (userId: number) => {
    try {
      set({ loading: true, error: null });
      const response = await userService.getUserById(userId);
      set({ currentUser: response.data, loading: false });
    } catch (error) {
      set({ 
        loading: false, 
        error: error instanceof Error ? error.message : '获取用户详情失败' 
      });
    }
  },
  
  // 创建用户
  createUser: async (userData: UserCreateRequest) => {
    try {
      set({ loading: true, error: null });
      const response = await userService.createUser(userData);
      
      // 创建成功后刷新用户列表
      await get().fetchUsers();
      
      set({ loading: false });
      return response.data;
    } catch (error) {
      set({ 
        loading: false, 
        error: error instanceof Error ? error.message : '创建用户失败' 
      });
      return null;
    }
  },
  
  // 更新用户
  updateUser: async (userId: number, userData: UserUpdateRequest) => {
    try {
      set({ loading: true, error: null });
      const response = await userService.updateUser(userId, userData);
      
      // 更新本地用户列表中的对应用户
      const updatedUsers = get().users.map(user => 
        user.id === userId ? response.data : user
      );
      
      set({ 
        users: updatedUsers,
        currentUser: userId === get().currentUser?.id ? response.data : get().currentUser,
        loading: false 
      });
      
      return response.data;
    } catch (error) {
      set({ 
        loading: false, 
        error: error instanceof Error ? error.message : '更新用户失败' 
      });
      return null;
    }
  },
  
  // 删除用户
  deleteUser: async (userId: number) => {
    try {
      set({ loading: true, error: null });
      await userService.deleteUser(userId);
      
      // 从列表中移除删除的用户
      const updatedUsers = get().users.filter(user => user.id !== userId);
      
      set({ 
        users: updatedUsers,
        currentUser: userId === get().currentUser?.id ? null : get().currentUser,
        loading: false 
      });
      
      // 如果当前页变空，且不是第一页，则转到上一页
      if (updatedUsers.length === 0 && get().page > 1) {
        await get().fetchUsers({ page: get().page - 1 });
      }
    } catch (error) {
      set({ 
        loading: false, 
        error: error instanceof Error ? error.message : '删除用户失败' 
      });
    }
  },
  
  // 批量删除用户
  batchDeleteUsers: async (userIds: number[]) => {
    try {
      set({ loading: true, error: null });
      await userService.batchDeleteUsers(userIds);
      
      // 从列表中移除所有删除的用户
      const updatedUsers = get().users.filter(user => !userIds.includes(user.id));
      
      // 如果当前用户被删除，则重置当前用户
      const currentUser = get().currentUser;
      const currentUserDeleted = currentUser ? userIds.includes(currentUser.id) : false;
      
      set({ 
        users: updatedUsers,
        currentUser: currentUserDeleted ? null : get().currentUser,
        loading: false 
      });
      
      // 如果当前页变空，且不是第一页，则转到上一页
      if (updatedUsers.length === 0 && get().page > 1) {
        await get().fetchUsers({ page: get().page - 1 });
      } else {
        // 刷新当前页
        await get().fetchUsers();
      }
    } catch (error) {
      set({ 
        loading: false, 
        error: error instanceof Error ? error.message : '批量删除用户失败' 
      });
    }
  },
  
  // 为用户分配角色
  assignRoles: async (userId: number, roleIds: number[]) => {
    try {
      set({ loading: true, error: null });
      await userService.assignRolesToUser(userId, roleIds);
      
      // 如果是当前用户，刷新用户详情
      if (userId === get().currentUser?.id) {
        await get().fetchUserById(userId);
      }
      
      set({ loading: false });
    } catch (error) {
      set({ 
        loading: false, 
        error: error instanceof Error ? error.message : '分配角色失败' 
      });
    }
  },
  
  // 从用户移除角色
  removeRoles: async (userId: number, roleIds: number[]) => {
    try {
      set({ loading: true, error: null });
      await userService.removeRolesFromUser(userId, roleIds);
      
      // 如果是当前用户，刷新用户详情
      if (userId === get().currentUser?.id) {
        await get().fetchUserById(userId);
      }
      
      set({ loading: false });
    } catch (error) {
      set({ 
        loading: false, 
        error: error instanceof Error ? error.message : '移除角色失败' 
      });
    }
  },
  
  // 设置搜索参数
  setSearchParams: (params: Partial<UserListParams>) => {
    set({ searchParams: { ...get().searchParams, ...params } });
  },
  
  // 重置状态
  resetState: () => {
    set({
      users: [],
      currentUser: null,
      total: 0,
      page: 1,
      pageSize: 10,
      loading: false,
      error: null,
      searchParams: {
        page: 1,
        pageSize: 10,
      }
    });
  }
}));

export default useUserStore;
