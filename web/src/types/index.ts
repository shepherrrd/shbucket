// User types
export interface User {
  id: string;
  username: string;
  email: string;
  role: 'admin' | 'manager' | 'editor' | 'viewer';
  is_active: boolean;
  created_at: string;
  updated_at: string;
  last_login?: string;
}

// Auth types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  token: string;
}

// Bucket types
export interface Bucket {
  id: string;
  name: string;
  description: string;
  auth_rule: AuthRule;
  settings: BucketSettings;
  stats: BucketStats;
  created_at: string;
  updated_at: string;
}

export interface AuthRule {
  id?: string;
  type: 'none' | 'api_key' | 'signed_url' | 'jwt' | 'session';
  enabled: boolean;
  config: AuthConfig;
  created_at?: string;
  updated_at?: string;
}

export interface AuthConfig {
  signing_secret?: string;
  jwt_secret?: string;
  session_timeout?: number;
  max_expiry?: number;
  default_expiry?: number;
  allowed_api_keys?: string[];
}

export interface BucketSettings {
  max_file_size: number;
  max_total_size: number;
  allowed_mime_types: string[];
  blocked_mime_types: string[];
  allowed_extensions: string[];
  blocked_extensions: string[];
  max_files_per_bucket: number;
  public_read: boolean;
  versioning: boolean;
  encryption: boolean;
  allow_overwrite: boolean;
  require_content_type: boolean;
}

export interface BucketStats {
  total_files: number;
  total_size: number;
  last_access?: string;
}

export interface CreateBucketRequest {
  name: string;
  description: string;
  auth_rule: AuthRule;
  settings: BucketSettings;
}

export interface UpdateBucketRequest {
  description?: string;
  auth_rule?: AuthRule;
  settings?: BucketSettings;
}

// File types
export interface File {
  id: string;
  bucket_id: string;
  name: string;
  path: string;
  size: number;
  mime_type: string;
  extension: string;
  checksum: string;
  auth_rule?: AuthRule;
  metadata: Record<string, any>;
  secured_url?: string;
  created_at: string;
  updated_at: string;
}

// Storage Node types
export interface StorageNode {
  id: string;
  name: string;
  url: string;
  max_storage: number;
  used_storage: number;
  priority: number;
  is_active: boolean;
  is_healthy: boolean;
  created_at: string;
  updated_at: string;
  last_ping?: string;
}

export interface CreateNodeRequest {
  name: string;
  url: string;
  auth_key: string;
  max_storage: number;
  priority: number;
  is_active: boolean;
}

// Signed URL types
export interface SignedURLRequest {
  bucket_name: string;
  file_name: string;
  bucket_id: string;
  file_id?: string;
  method: string;
  expires_in: number;
}

export interface SignedURLResponse {
  url: string;
  expires_at: string;
  success: boolean;
  message: string;
}

// API Key types
export interface APIKeyPermission {
  read: boolean;
  write: boolean;
  sign_urls: boolean;
  buckets: string[];
}

export interface APIKey {
  id: string;
  name: string;
  key_prefix: string;
  user_id: string;
  username: string;
  is_active: boolean;
  permissions: APIKeyPermission;
  expires_at?: string;
  last_used?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateAPIKeyRequest {
  name: string;
  permissions: APIKeyPermission;
  expires_in?: number; // seconds from now, optional
}

export interface CreateAPIKeyResponse {
  api_key: APIKey;
  key: string; // The actual key - only returned once
  success: boolean;
  message: string;
}

// API Response types
export interface APIResponse<T = any> {
  data?: T;
  error?: string;
  message?: string;
}

// Pagination types
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
  has_next: boolean;
  has_prev: boolean;
}