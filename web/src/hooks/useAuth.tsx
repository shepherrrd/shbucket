import { useState, useEffect, createContext, useContext } from 'react';
import type { ReactNode } from 'react';
import type { User, LoginRequest } from '../types';
import { apiClient } from '../services/api';
import toast from 'react-hot-toast';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (credentials: LoginRequest) => Promise<void>;
  logout: () => Promise<void>;
  setAPIKey: (apiKey: string) => void;
  isAuthenticated: boolean;
  hasRole: (role: string) => boolean;
  hasPermission: (permission: string) => boolean;
  authMethod: 'token' | 'api_key' | null;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [authMethod, setAuthMethod] = useState<'token' | 'api_key' | null>(null);

  useEffect(() => {
    checkAuthStatus();
  }, []);

  const checkAuthStatus = async () => {
    try {
      const token = localStorage.getItem('shbucket_token');
      const apiKey = localStorage.getItem('shbucket_api_key');
      
      if (token || apiKey) {
        const currentUser = await apiClient.getCurrentUser();
        setUser(currentUser);
        setAuthMethod(apiKey ? 'api_key' : 'token');
      }
    } catch (error) {
      // Auth is invalid, clear it
      localStorage.removeItem('shbucket_token');
      localStorage.removeItem('shbucket_api_key');
      setAuthMethod(null);
    } finally {
      setLoading(false);
    }
  };

  const login = async (credentials: LoginRequest) => {
    try {
      const response = await apiClient.login(credentials);
      setUser(response.user);
      setAuthMethod('token');
      toast.success('Logged in successfully');
    } catch (error) {
      throw error;
    }
  };

  const setAPIKey = (apiKey: string) => {
    apiClient.setAPIKey(apiKey);
    setAuthMethod('api_key');
    checkAuthStatus(); // Re-check to get user info
    toast.success('API key set successfully');
  };

  const logout = async () => {
    try {
      await apiClient.logout();
    } catch (error) {
      // Continue with logout even if API call fails
    } finally {
      setUser(null);
      setAuthMethod(null);
      toast.success('Logged out successfully');
    }
  };

  const isAuthenticated = !!user;

  const hasRole = (role: string): boolean => {
    if (!user) return false;
    
    const roleHierarchy: Record<string, string[]> = {
      admin: ['admin', 'manager', 'editor', 'viewer'],
      manager: ['manager', 'editor', 'viewer'],
      editor: ['editor', 'viewer'],
      viewer: ['viewer'],
    };

    return roleHierarchy[user.role]?.includes(role) || false;
  };

  const hasPermission = (permission: string): boolean => {
    if (!user) return false;

    const permissions: Record<string, string[]> = {
      admin: ['read', 'write', 'delete', 'manage', 'admin'],
      manager: ['read', 'write', 'delete', 'manage'],
      editor: ['read', 'write', 'delete'],
      viewer: ['read'],
    };

    return permissions[user.role]?.includes(permission) || false;
  };

  const value = {
    user,
    loading,
    login,
    logout,
    setAPIKey,
    isAuthenticated,
    hasRole,
    hasPermission,
    authMethod,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}