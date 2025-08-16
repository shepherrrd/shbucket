import axios from 'axios';
import type { AxiosResponse } from 'axios';
import toast from 'react-hot-toast';
import type {
  User,
  Bucket,
  File,
  StorageNode,
  LoginRequest,
  LoginResponse,
  CreateBucketRequest,
  UpdateBucketRequest,
  CreateNodeRequest,
  SignedURLResponse,
  APIKey,
  CreateAPIKeyRequest,
  CreateAPIKeyResponse,
} from '../types';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

class APIClient {
  private token: string | null = null;
  private apiKey: string | null = null;

  constructor() {
    // Get token and API key from localStorage on initialization
    this.token = localStorage.getItem('shbucket_token');
    this.apiKey = localStorage.getItem('shbucket_api_key');
    
    // Set up axios interceptors
    axios.defaults.baseURL = API_BASE_URL;
    
    // Request interceptor to add auth token or API key
    axios.interceptors.request.use((config) => {
      if (this.apiKey) {
        config.headers['X-API-Key'] = this.apiKey;
      } else if (this.token) {
        config.headers.Authorization = `Bearer ${this.token}`;
      }
      return config;
    });

    // Response interceptor for error handling
    axios.interceptors.response.use(
      (response) => response,
      (error) => {
        // Only auto-logout on 401 for auth endpoints, not login attempts, and only if we have a token or API key
        if (error.response?.status === 401 && !error.config?.url?.includes('/auth/login') && (this.token || this.apiKey)) {
          this.logout();
          toast.error('Authentication failed. Please login again.');
        }
        return Promise.reject(error);
      }
    );
  }

  private setToken(token: string) {
    this.token = token;
    localStorage.setItem('shbucket_token', token);
  }

  private clearToken() {
    this.token = null;
    localStorage.removeItem('shbucket_token');
  }

  setAPIKey(apiKey: string) {
    this.apiKey = apiKey;
    localStorage.setItem('shbucket_api_key', apiKey);
  }

  private clearAPIKey() {
    this.apiKey = null;
    localStorage.removeItem('shbucket_api_key');
  }

  clearAuth() {
    this.clearToken();
    this.clearAPIKey();
  }

  isAuthenticated(): boolean {
    return !!(this.token || this.apiKey);
  }

  async request<T = any>(method: 'GET' | 'POST' | 'PUT' | 'DELETE', url: string, data?: any): Promise<T> {
    try {
      const response: AxiosResponse<T> = await axios.request({
        method,
        url,
        data,
      });
      return response.data;
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || error.message || 'An error occurred';
      throw new Error(errorMessage);
    }
  }

  // Auth endpoints
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const response = await this.request<LoginResponse>('POST', '/auth/login', credentials);
    this.setToken(response.token);
    return response;
  }

  async logout(): Promise<void> {
    try {
      // Only make API call if we actually have a token (not for API keys)
      if (this.token) {
        await this.request('POST', '/auth/logout');
      }
    } catch (error) {
      // Continue with logout even if API call fails
    } finally {
      this.clearAuth();
    }
  }

  async getCurrentUser(): Promise<User> {
    return this.request<User>('GET', '/auth/me');
  }

  // Bucket endpoints
  async getBuckets(): Promise<{ buckets: Bucket[]; total: number }> {
    return this.request('GET', '/buckets');
  }

  async getBucket(id: string): Promise<Bucket> {
    return this.request('GET', `/buckets/${id}`);
  }

  async createBucket(bucket: CreateBucketRequest): Promise<Bucket> {
    return this.request('POST', '/buckets', bucket);
  }

  async updateBucket(id: string, updates: UpdateBucketRequest): Promise<Bucket> {
    return this.request('PUT', `/buckets/${id}`, updates);
  }

  async deleteBucket(id: string): Promise<void> {
    return this.request('DELETE', `/buckets/${id}`);
  }

  // File endpoints
  async getFiles(bucketId: string): Promise<{ files: File[]; total: number }> {
    return this.request('GET', `/buckets/${bucketId}/files`);
  }

  async uploadFile(bucketId: string, file: globalThis.File): Promise<File> {
    const formData = new FormData();
    formData.append('file', file);

    try {
      const response: AxiosResponse<File> = await axios.post(
        `/buckets/${bucketId}/files`,
        formData,
        {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        }
      );
      return response.data;
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || error.message || 'Upload failed';
      throw new Error(errorMessage);
    }
  }

  async downloadFile(bucketId: string, fileId: string): Promise<Blob> {
    try {
      const response: AxiosResponse<Blob> = await axios.get(
        `/file/${bucketId}/${fileId}`,
        { responseType: 'blob' }
      );
      return response.data;
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || error.message || 'Download failed';
      throw new Error(errorMessage);
    }
  }

  async deleteFile(bucketId: string, fileId: string): Promise<void> {
    return this.request('DELETE', `/buckets/${bucketId}/files/${fileId}`);
  }

  async getFileMetadata(bucketId: string, fileName: string): Promise<{ file: File }> {
    return this.request('GET', `/buckets/${bucketId}/files/${fileName}/metadata`);
  }

  async updateFileAuth(bucketId: string, fileName: string, authRule: any): Promise<void> {
    return this.request('PUT', `/buckets/${bucketId}/files/${fileName}/auth`, { auth_rule: authRule });
  }

  // Signed URLs
  async generateSignedURL(bucketId: string, fileId: string, expiresIn: number, singleUse: boolean = false): Promise<SignedURLResponse> {
    return this.request('POST', `/buckets/${bucketId}/files/${fileId}/signed-url`, { expires_in: expiresIn, single_use: singleUse });
  }

  // API Key endpoints
  async getAPIKeys(): Promise<{ api_keys: APIKey[]; total: number }> {
    return this.request('GET', '/api-keys');
  }

  async createAPIKey(request: CreateAPIKeyRequest): Promise<CreateAPIKeyResponse> {
    return this.request('POST', '/api-keys', request);
  }

  async deleteAPIKey(id: string): Promise<void> {
    return this.request('DELETE', `/api-keys/${id}`);
  }

  // Storage Nodes
  async getNodes(): Promise<{ nodes: StorageNode[]; total: number }> {
    return this.request('GET', '/nodes');
  }

  async createNode(node: CreateNodeRequest): Promise<StorageNode> {
    return this.request('POST', '/nodes', node);
  }

  async updateNode(id: string, updates: Partial<CreateNodeRequest>): Promise<void> {
    return this.request('PUT', `/nodes/${id}`, updates);
  }

  async deleteNode(id: string): Promise<void> {
    return this.request('DELETE', `/nodes/${id}`);
  }

  async checkNodeHealth(id: string): Promise<any> {
    return this.request('GET', `/nodes/${id}/health`);
  }

  async checkAllNodesHealth(): Promise<any> {
    return this.request('GET', '/nodes/health');
  }

  // Users (admin only)
  async getUsers(): Promise<{ users: User[]; total: number }> {
    return this.request('GET', '/users');
  }

  // Health
  async getHealth(): Promise<{ status: string; service: string }> {
    return this.request('GET', '/health');
  }

  // Setup endpoints (no auth required)
  async getSetupStatus(): Promise<{ is_setup: boolean; setup_type?: string; message: string }> {
    return this.request('GET', '/setup/status');
  }

  async setupMaster(data: any): Promise<any> {
    return this.request('POST', '/setup/master', data);
  }

  async setupNode(data: any): Promise<any> {
    return this.request('POST', '/setup/node', data);
  }
}

export const apiClient = new APIClient();
export const api = apiClient;