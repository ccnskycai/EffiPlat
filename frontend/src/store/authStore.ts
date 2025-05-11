// Complete rewrite of auth store without any Zustand persist functionality
// Using vanilla React state management patterns to avoid getSnapshot errors

import { create } from 'zustand';
import {
  getCurrentUser as apiGetCurrentUser,
  login as apiLogin,
  logout as apiLogout,
  setAuthToken,
} from '../services/authService';

// Define User interface
interface User {
  id: number;
  name: string;
  email?: string;
  roles?: string[];
}

// Interface for data stored in localStorage
interface StoredAuthData {
  token: string | null;
  user: User | null;
}

// Store state type definition
type AuthState = {
  isAuthenticated: boolean;
  user: User | null;
  token: string | null;
  isLoading: boolean;
  error: string | null;
  login: (credentials: Parameters<typeof apiLogin>[0]) => Promise<void>;
  logout: () => Promise<void>;
  initializeAuth: () => Promise<void>;
  clearError: () => void;
  setAuth: (data: { token: string; user: User }) => void;
  resetAuth: () => void;
};

// Try to load initial data from localStorage
let initialToken: string | null = null;
let initialUser: User | null = null;
let initialIsAuthenticated = false;

try {
  const savedAuthData = localStorage.getItem('auth-data');
  if (savedAuthData) {
    const parsed: StoredAuthData = JSON.parse(savedAuthData) as StoredAuthData;
    initialToken = parsed.token || null;
    initialUser = parsed.user || null;
    initialIsAuthenticated = !!parsed.token;

    // Set token for API calls if available
    if (initialToken) {
      setAuthToken(initialToken);
    }
  }
} catch (e) {
  console.error('Failed to parse saved auth data:', e);
}

// Create a basic store without any persistence middleware
export const useAuthStore = create<AuthState>((set, get) => ({
  // Initial state from localStorage if available
  isAuthenticated: initialIsAuthenticated,
  user: initialUser,
  token: initialToken,
  isLoading: false,
  error: null,

  // Set authentication data
  setAuth: ({ token, user }: { token: string; user: User }) => {
    setAuthToken(token);

    // Update local storage
    localStorage.setItem('auth-data', JSON.stringify({ token, user }));

    // Update state
    set({
      isAuthenticated: true,
      user,
      token,
      error: null,
    });
  },

  // Reset authentication data
  resetAuth: () => {
    setAuthToken(null);
    localStorage.removeItem('auth-data');
    set({
      isAuthenticated: false,
      user: null,
      token: null,
      error: null,
    });
  },

  // Clear error state
  clearError: () => set({ error: null }),

  // Login function
  login: async (credentials: Parameters<typeof apiLogin>[0]) => {
    set({ isLoading: true, error: null });

    try {
      // Perform login
      const { token, user: basicUserFromApi } = await apiLogin(credentials);

      // Construct the user object for the store.
      // basicUserFromApi is of type Omit<User, 'email'>, so it won't have an email property.
      // The User interface in the store has email as optional (email?: string).
      const userForStore: User = {
        id: basicUserFromApi.id,
        name: basicUserFromApi.name,
        roles: basicUserFromApi.roles || undefined, // roles is optional in User interface, ensure it's handled if basicUserFromApi.roles can be null/undefined
        // email is not provided by basicUserFromApi, and it's optional in User, so it will be undefined here.
      };
      get().setAuth({ token, user: userForStore });
    } catch (error: unknown) {
      // Explicitly type error as unknown to satisfy linter for general catch blocks
      const errorMessage = error instanceof Error ? error.message : 'Login failed';
      set({
        isLoading: false,
        error: errorMessage,
      });
      throw error; // Re-throw the error if further handling is needed 소비자측에서
    } finally {
      set({ isLoading: false });
    }
  },

  // Logout function
  logout: async () => {
    set({ isLoading: true });

    try {
      const token = get().token;
      if (token) {
        try {
          await apiLogout();
        } catch (error) {
          // Log the error but don't prevent the local logout
          // Only log if it's a real error, not a misleading success message
          if (!(error instanceof Error && error.message.includes('successful'))) {
            console.error('Backend logout failed:', error);
          }
        }
      }
      // Always reset auth state locally even if the API call fails
      get().resetAuth();
    } finally {
      set({ isLoading: false });
    }
  },

  // Initialize auth state
  initializeAuth: async () => {
    const token = get().token;

    if (!token) {
      set({ isLoading: false });
      return;
    }

    set({ isLoading: true });
    setAuthToken(token);

    try {
      const user = await apiGetCurrentUser();
      set({
        isAuthenticated: true,
        user,
        isLoading: false,
      });
    } catch (error) {
      console.error('Failed to initialize auth:', error);
      get().resetAuth();
      set({ isLoading: false });
    }
  },
}));
