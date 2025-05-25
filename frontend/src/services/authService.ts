import { isAxiosError } from 'axios';
import apiClient, { setAuthToken } from './apiClient';

// Interfaces based on auth_feature_design.md
export interface LoginCredentials {
  email: string;
  password: string;
}

export interface User {
  id: number;
  name: string;
  email: string;
  roles?: string[];
  department: string | null;
  status: string;
  createdAt?: string;
  updatedAt?: string;
}

// This is the structure authStore expects the login function to resolve to.
export interface AuthResponseData {
  token: string;
  user: Omit<User, 'email'>; // Contains id, name, optional roles
}

// Expected structure for the /me endpoint response data
export type MeApiResponseData = User;

// Define an interface for the actual successful login response from the backend
interface ActualBackendLoginSuccessResponse {
  token: string;
  user: {
    id: number;
    name: string;
    email: string;
    department?: string | null;
    status: string;
    createdAt: string;
    updatedAt: string;
    roles?: string[];
  };
}

// Define an interface for the wrapped API response format
interface WrappedApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

// Interface for an unexpected success payload (e.g. HTTP 200 but not the expected structure)
interface UnexpectedSuccessPayload {
  message?: string;
  [key: string]: unknown; // Use unknown for better type safety
}

// Interface for a structured API error response from backend (e.g. on 401, 400)
interface ApiErrorResponse {
  code?: number;
  message?: string;
  data?: unknown; // Use unknown for better type safety
}

// Specific error message/type for when /me endpoint is not found
export const ME_ENDPOINT_NOT_AVAILABLE_ERROR = 'ME_ENDPOINT_NOT_AVAILABLE_ERROR';

/**
 * Handles user login.
 * @param credentials - User's email and password.
 * @returns - Promise resolving to token and basic user info.
 */
export const login = async (credentials: LoginCredentials): Promise<AuthResponseData> => {
  try {
    // 详细记录登录请求信息
    console.log('[authService.ts] login: Attempting to login with credentials:', {
      email: credentials.email,
      passwordLength: credentials.password.length
    });

    const response = await apiClient.post<
      ActualBackendLoginSuccessResponse | WrappedApiResponse<ActualBackendLoginSuccessResponse> | UnexpectedSuccessPayload
    >('/api/v1/auth/login', credentials);

    console.log('[authService.ts] login: Backend response status:', response.status);
    console.log('[authService.ts] login: Backend response data:', response.data);

    // 检查是否是包装的API响应格式
    if (
      response.data &&
      typeof response.data === 'object' &&
      'code' in response.data &&
      'message' in response.data &&
      'data' in response.data &&
      typeof response.data.data === 'object' &&
      response.data.data !== null &&
      'token' in response.data.data &&
      'user' in response.data.data
    ) {
      const wrappedResponse = response.data as WrappedApiResponse<ActualBackendLoginSuccessResponse>;
      const loginData = wrappedResponse.data;
      
      // 检查code是否表示成功
      if (wrappedResponse.code !== 0) {
        throw new Error(wrappedResponse.message || 'Login failed with non-zero code');
      }
      
      const backendUser = loginData.user;
      
      const userData: Omit<User, 'email'> = {
        id: backendUser.id,
        name: backendUser.name,
        department: backendUser.department ?? null,
        status: backendUser.status,
        createdAt: backendUser.createdAt,
        updatedAt: backendUser.updatedAt,
        roles: backendUser.roles
      };

      // 设置认证令牌
      setAuthToken(loginData.token);
      
      return { token: loginData.token, user: userData };
    }
    // 检查是否是直接返回的格式
    else if (
      response.data &&
      'token' in response.data &&
      'user' in response.data &&
      typeof response.data.token === 'string' &&
      typeof response.data.user === 'object' &&
      response.data.user !== null
    ) {
      const loginData = response.data as ActualBackendLoginSuccessResponse;
      const backendUser = loginData.user;
      
      const userData: Omit<User, 'email'> = {
        id: backendUser.id,
        name: backendUser.name,
        department: backendUser.department ?? null,
        status: backendUser.status,
        createdAt: backendUser.createdAt,
        updatedAt: backendUser.updatedAt,
        roles: backendUser.roles
      };

      // 设置认证令牌
      setAuthToken(loginData.token);

      return { token: loginData.token, user: userData };
    } else {
      const errorPayload = response.data as UnexpectedSuccessPayload;
      const errorMessage =
        typeof errorPayload?.message === 'string'
          ? errorPayload.message
          : 'Login failed: Unexpected response structure from backend.';
      console.error(
        '[authService.ts] login: Unexpected success response structure:',
        response.data
      );
      throw new Error(errorMessage);
    }
  } catch (error: unknown) {
    console.error('[authService.ts] login: Caught error:', error);

    if (isAxiosError(error)) {
      // 更详细的错误信息输出
      console.error('[authService.ts] login: Axios error details:', {
        message: error.message,
        code: error.code,
        status: error.response?.status,
        statusText: error.response?.statusText,
        data: error.response?.data,
        headers: error.response?.headers,
        url: error.config?.url,
        method: error.config?.method,
        requestData: error.config?.data
      });
      
      // 分析403错误的可能原因
      if (error.response?.status === 403) {
        console.error(
          '[authService.ts] 403 Forbidden could indicate: (1) 账户凭证不正确, (2) 账户被锁定, (3) 认证端点不正确'
        );
      }
    }
    throw error;
  }
};

/**
 * Fetches the current authenticated user's information.
 * Returns null if the user cannot be fetched (e.g., /me not available or other errors).
 * @returns - Promise resolving to the user's information or null.
 */
export const getCurrentUser = async (): Promise<User | null> => {
  try {
    console.log('[authService.ts] getCurrentUser: Making request to /api/v1/auth/me endpoint...');
    const response = await apiClient.get('/api/v1/auth/me');
    console.log('[authService.ts] getCurrentUser: Backend response received!');
    
    // First check if we're dealing with a wrapped API response (code/message/data structure)
    if (
      response.data &&
      typeof response.data === 'object' &&
      'code' in response.data &&
      'message' in response.data &&
      'data' in response.data
    ) {
      console.log('[authService.ts] getCurrentUser: Detected wrapped API response format.');
      const apiResponse = response.data as { code: number; message: string; data: MeApiResponseData };
      
      if (apiResponse.code === 0 && apiResponse.data) {
        console.log(
          '[authService.ts] getCurrentUser: Successfully fetched user data:',
          JSON.stringify(apiResponse.data, null, 2)
        );
        return apiResponse.data;
      } else {
        const message = typeof apiResponse.message === 'string'
          ? apiResponse.message
          : 'Failed to fetch user information (non-zero code or missing data)';
        console.error('[authService.ts] getCurrentUser: ' + message);
        return null;
      }
    } 
    // Handle direct user object response (no wrapper with code/message/data)
    else if (response.data && typeof response.data === 'object' && 'id' in response.data && 'email' in response.data) {
      const userData = response.data as MeApiResponseData;
      console.log(
        '[authService.ts] getCurrentUser: Successfully fetched user data (direct format):',
        JSON.stringify(userData, null, 2)
      );
      return userData;
    } 
    // Unexpected response format
    else {
      console.error(
        '[authService.ts] getCurrentUser: Failed to fetch user. Unexpected response structure:',
        JSON.stringify(response.data, null, 2)
      );
      return null;
    }
  } catch (error: unknown) {
    console.error('[authService.ts] getCurrentUser: Catching error during /me call.', error);

    if (isAxiosError(error)) {
      console.error('[authService.ts] getCurrentUser: Axios error details:', {
        message: error.message,
        code: error.code,
        status: error.response?.status,
        data: error.response?.data
          ? JSON.stringify(error.response.data, null, 2)
          : 'No response data',
        config_url: error.config?.url,
      });

      if (error.response?.status === 404) {
        console.warn(
          '[authService.ts] getCurrentUser: /me endpoint returned 404. Endpoint not available?'
        );
        return null;
      }
      return null;
    }
    if (error instanceof Error) {
      console.error(
        '[authService.ts] getCurrentUser: Generic error instance:',
        error.message,
        error.stack
      );
      return null;
    }
    console.error('[authService.ts] getCurrentUser: Unknown error type:', error);
    return null;
  }
};

/**
 * Logs out the current user.
 * Assumes token is set in apiClient default headers.
 * @returns - Promise resolving when logout is successful.
 */
export const logout = async (): Promise<void> => {
  try {
    const response = await apiClient.post<{
      code: number;
      message: string;
      data: null;
    }>('/api/v1/auth/logout');

    // For development/debugging purposes, log the response
    console.log('[authService.ts] logout: Response received:', response.data);
    
    // Check if the response contains a non-zero code, which indicates an error from the backend
    if (response.data.code !== 0 && response.data.code !== undefined) {
      // Only throw an error if the code is explicitly non-zero
      throw new Error(response.data.message || 'Backend logout failed with unknown error');
    }
    
    // If we reach here, logout was successful
    // 清除认证令牌
    setAuthToken(null);
    
  } catch (error: unknown) {
    // Only log the error if it's not a successful logout with a misleading message
    if (!(error instanceof Error && error.message.includes('successful'))) {
      console.error('[authService.ts] Logout API call failed:', error);
    }
    
    if (isAxiosError(error) && error.response) {
      const apiErrorData = error.response.data as ApiErrorResponse | undefined;
      const message =
        typeof apiErrorData?.message === 'string' ? apiErrorData.message : 'Logout API call failed';
      throw new Error(message);
    }
    if (error instanceof Error) {
      throw new Error(error.message || 'Logout API call failed due to an unknown error');
    }
    throw new Error('Logout API call failed due to an unknown error');
  } finally {
    // 无论如何都清除令牌，确保用户登出
    setAuthToken(null);
  }
};
