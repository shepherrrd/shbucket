import React, { useState, useEffect } from 'react';
import { api } from '../services/api';

interface SetupStatus {
  is_setup: boolean;
  setup_type?: string;
  node_name?: string;
  message: string;
}

interface MasterSetupForm {
  adminUsername: string;
  adminEmail: string;
  adminPassword: string;
  confirmPassword: string;
  storagePath: string;
  maxStorage: number;
  systemName: string;
  jwtSecret: string;
  defaultAuthRule: {
    type: string;
    enabled: boolean;
    config: any;
  };
  defaultSettings: {
    maxFileSize: number;
    publicRead: boolean;
    versioning: boolean;
    encryption: boolean;
  };
}

interface NodeSetupForm {
  masterUrl: string;
  nodeName: string;
  nodeApiKey: string;
  storagePath: string;
  maxStorage: number;
  masterApiKey: string;
}

const Setup: React.FC = () => {
  const [setupStatus, setSetupStatus] = useState<SetupStatus | null>(null);
  const [setupType, setSetupType] = useState<'master' | 'node'>('master');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const [masterForm, setMasterForm] = useState<MasterSetupForm>({
    adminUsername: 'admin',
    adminEmail: 'admin@shbucket.local',
    adminPassword: '',
    confirmPassword: '',
    storagePath: './storage',
    maxStorage: 10737418240, // 10GB
    systemName: 'SHBucket',
    jwtSecret: '',
    defaultAuthRule: {
      type: 'none',
      enabled: false,
      config: {}
    },
    defaultSettings: {
      maxFileSize: 104857600, // 100MB
      publicRead: false,
      versioning: false,
      encryption: false
    }
  });

  const [nodeForm, setNodeForm] = useState<NodeSetupForm>({
    masterUrl: '',
    nodeName: '',
    nodeApiKey: '',
    storagePath: './storage',
    maxStorage: 10737418240, // 10GB
    masterApiKey: ''
  });

  useEffect(() => {
    checkSetupStatus();
  }, []);

  const checkSetupStatus = async () => {
    try {
      const response = await api.getSetupStatus();
      setSetupStatus(response);
      
      // Don't auto-redirect, let the component render the "already setup" message
      // if (response.is_setup) {
      //   window.location.href = '/login';
      // }
    } catch (err) {
      console.error('Failed to check setup status:', err);
      // If API fails, assume not setup
      setSetupStatus({
        is_setup: false,
        message: 'Setup required'
      });
    }
  };

  const handleMasterSetup = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (masterForm.adminPassword !== masterForm.confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (masterForm.adminPassword.length < 6) {
      setError('Password must be at least 6 characters');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const payload = {
        admin_username: masterForm.adminUsername,
        admin_email: masterForm.adminEmail,
        admin_password: masterForm.adminPassword,
        storage_path: masterForm.storagePath,
        max_storage: masterForm.maxStorage,
        system_name: masterForm.systemName,
        jwt_secret: masterForm.jwtSecret || undefined,
        default_auth_rule: masterForm.defaultAuthRule,
        default_settings: {
          max_file_size: masterForm.defaultSettings.maxFileSize,
          public_read: masterForm.defaultSettings.publicRead,
          versioning: masterForm.defaultSettings.versioning,
          encryption: masterForm.defaultSettings.encryption,
          max_total_size: 0,
          allowed_mime_types: [],
          blocked_mime_types: [],
          allowed_extensions: [],
          blocked_extensions: [],
          max_files_per_bucket: 10000,
          allow_overwrite: true,
          require_content_type: false
        }
      };

      const response = await api.setupMaster(payload);
      
      if (response.success) {
        setSuccess('Master setup completed successfully! Redirecting to login...');
        setTimeout(() => {
          window.location.href = '/login';
        }, 2000);
      }
    } catch (err: any) {
      setError(err.message || 'Setup failed');
    } finally {
      setLoading(false);
    }
  };

  const handleNodeSetup = async (e: React.FormEvent) => {
    e.preventDefault();
    
    setLoading(true);
    setError(null);

    try {
      const payload = {
        master_url: nodeForm.masterUrl,
        node_name: nodeForm.nodeName,
        node_api_key: nodeForm.nodeApiKey,
        storage_path: nodeForm.storagePath,
        max_storage: nodeForm.maxStorage,
        master_api_key: nodeForm.masterApiKey
      };

      const response = await api.setupNode(payload);
      
      if (response.success) {
        setSuccess('Node setup completed successfully! Node is now connected to master.');
        setTimeout(() => {
          window.location.href = '/login';
        }, 2000);
      }
    } catch (err: any) {
      setError(err.message || 'Node setup failed');
    } finally {
      setLoading(false);
    }
  };

  const generateApiKey = () => {
    const key = 'node_' + Array.from(crypto.getRandomValues(new Uint8Array(16)), 
      b => b.toString(16).padStart(2, '0')).join('');
    
    if (setupType === 'node') {
      setNodeForm(prev => ({ ...prev, nodeApiKey: key }));
    }
  };

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  if (setupStatus === null) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Checking system status...</p>
        </div>
      </div>
    );
  }

  if (setupStatus.is_setup) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-dark-950">
        <div className="text-center">
          <div className="text-green-400 text-6xl mb-4">✓</div>
          <h1 className="text-2xl font-bold text-dark-50 mb-2">System Already Configured</h1>
          <p className="text-dark-300 mb-4">{setupStatus.message}</p>
          <button 
            onClick={() => window.location.href = '/login'}
            className="bg-primary-600 text-white px-6 py-2 rounded-lg hover:bg-primary-700"
          >
            Go to Login
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-dark-950 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md mx-auto">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-dark-50">SHBucket Setup</h1>
          <p className="mt-2 text-dark-300">Welcome! Let's configure your SHBucket instance.</p>
        </div>

        {error && (
          <div className="bg-red-900/20 border border-red-500/50 text-red-400 px-4 py-3 rounded mb-4">
            {error}
          </div>
        )}

        {success && (
          <div className="bg-green-900/20 border border-green-500/50 text-green-400 px-4 py-3 rounded mb-4">
            {success}
          </div>
        )}

        <div className="bg-dark-800 shadow rounded-lg p-6 border border-dark-700">
          {/* Setup Type Selection */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-dark-200 mb-3">
              Setup Type
            </label>
            <div className="grid grid-cols-2 gap-3">
              <button
                type="button"
                onClick={() => setSetupType('master')}
                className={`p-4 border rounded-lg text-center ${
                  setupType === 'master'
                    ? 'border-primary-500 bg-primary-500/10 text-primary-400'
                    : 'border-dark-600 hover:border-dark-500 text-dark-300'
                }`}
              >
                <div className="font-medium">Master Server</div>
                <div className="text-sm text-dark-400 mt-1">
                  Primary server with web interface
                </div>
              </button>
              <button
                type="button"
                onClick={() => setSetupType('node')}
                className={`p-4 border rounded-lg text-center ${
                  setupType === 'node'
                    ? 'border-primary-500 bg-primary-500/10 text-primary-400'
                    : 'border-dark-600 hover:border-dark-500 text-dark-300'
                }`}
              >
                <div className="font-medium">Storage Node</div>
                <div className="text-sm text-dark-400 mt-1">
                  Additional storage for master
                </div>
              </button>
            </div>
          </div>

          {setupType === 'master' ? (
            <form onSubmit={handleMasterSetup} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-dark-200">System Name</label>
                <input
                  type="text"
                  value={masterForm.systemName}
                  onChange={(e) => setMasterForm(prev => ({ ...prev, systemName: e.target.value }))}
                  className="mt-1 block w-full bg-dark-700 border border-dark-600 text-dark-100 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-dark-200">Admin Username</label>
                <input
                  type="text"
                  value={masterForm.adminUsername}
                  onChange={(e) => setMasterForm(prev => ({ ...prev, adminUsername: e.target.value }))}
                  className="mt-1 block w-full bg-dark-700 border border-dark-600 text-dark-100 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-dark-200">Admin Email</label>
                <input
                  type="email"
                  value={masterForm.adminEmail}
                  onChange={(e) => setMasterForm(prev => ({ ...prev, adminEmail: e.target.value }))}
                  className="mt-1 block w-full bg-dark-700 border border-dark-600 text-dark-100 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-dark-200">Admin Password</label>
                <input
                  type="password"
                  value={masterForm.adminPassword}
                  onChange={(e) => setMasterForm(prev => ({ ...prev, adminPassword: e.target.value }))}
                  className="mt-1 block w-full bg-dark-700 border border-dark-600 text-dark-100 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  required
                  minLength={6}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-dark-200">Confirm Password</label>
                <input
                  type="password"
                  value={masterForm.confirmPassword}
                  onChange={(e) => setMasterForm(prev => ({ ...prev, confirmPassword: e.target.value }))}
                  className="mt-1 block w-full bg-dark-700 border border-dark-600 text-dark-100 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  required
                  minLength={6}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-dark-200">Storage Path</label>
                <input
                  type="text"
                  value={masterForm.storagePath}
                  onChange={(e) => setMasterForm(prev => ({ ...prev, storagePath: e.target.value }))}
                  className="mt-1 block w-full bg-dark-700 border border-dark-600 text-dark-100 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-dark-200">
                  Max Storage ({formatBytes(masterForm.maxStorage)})
                </label>
                <input
                  type="range"
                  min={1073741824} // 1GB
                  max={1099511627776} // 1TB
                  step={1073741824} // 1GB steps
                  value={masterForm.maxStorage}
                  onChange={(e) => setMasterForm(prev => ({ ...prev, maxStorage: parseInt(e.target.value) }))}
                  className="mt-1 block w-full"
                />
              </div>

              <div className="space-y-2">
                <label className="block text-sm font-medium text-dark-200">Default Settings</label>
                <div className="space-y-2">
                  <label className="flex items-center">
                    <input
                      type="checkbox"
                      checked={masterForm.defaultSettings.publicRead}
                      onChange={(e) => setMasterForm(prev => ({
                        ...prev,
                        defaultSettings: { ...prev.defaultSettings, publicRead: e.target.checked }
                      }))}
                      className="rounded"
                    />
                    <span className="ml-2 text-sm text-dark-300">Public Read by Default</span>
                  </label>
                  <label className="flex items-center">
                    <input
                      type="checkbox"
                      checked={masterForm.defaultSettings.versioning}
                      onChange={(e) => setMasterForm(prev => ({
                        ...prev,
                        defaultSettings: { ...prev.defaultSettings, versioning: e.target.checked }
                      }))}
                      className="rounded"
                    />
                    <span className="ml-2 text-sm text-dark-300">Enable Versioning</span>
                  </label>
                  <label className="flex items-center">
                    <input
                      type="checkbox"
                      checked={masterForm.defaultSettings.encryption}
                      onChange={(e) => setMasterForm(prev => ({
                        ...prev,
                        defaultSettings: { ...prev.defaultSettings, encryption: e.target.checked }
                      }))}
                      className="rounded"
                    />
                    <span className="ml-2 text-sm text-dark-300">Enable Encryption</span>
                  </label>
                </div>
              </div>

              <button
                type="submit"
                disabled={loading}
                className="w-full bg-primary-600 text-white py-2 px-4 rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading ? 'Setting up Master...' : 'Setup Master Server'}
              </button>
            </form>
          ) : (
            <form onSubmit={handleNodeSetup} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-dark-200">Master Server URL</label>
                <input
                  type="url"
                  value={nodeForm.masterUrl}
                  onChange={(e) => setNodeForm(prev => ({ ...prev, masterUrl: e.target.value }))}
                  placeholder="http://master-server:8080"
                  className="mt-1 block w-full bg-dark-700 border border-dark-600 text-dark-100 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-dark-200">Node Name</label>
                <input
                  type="text"
                  value={nodeForm.nodeName}
                  onChange={(e) => setNodeForm(prev => ({ ...prev, nodeName: e.target.value }))}
                  placeholder="storage-node-1"
                  className="mt-1 block w-full bg-dark-700 border border-dark-600 text-dark-100 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-dark-200">Node API Key</label>
                <div className="flex">
                  <input
                    type="text"
                    value={nodeForm.nodeApiKey}
                    onChange={(e) => setNodeForm(prev => ({ ...prev, nodeApiKey: e.target.value }))}
                    className="flex-1 bg-dark-700 border border-dark-600 text-dark-100 rounded-l-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                    required
                  />
                  <button
                    type="button"
                    onClick={generateApiKey}
                    className="bg-dark-600 text-white px-3 py-2 rounded-r-md hover:bg-dark-500 border border-l-0 border-dark-600"
                  >
                    Generate
                  </button>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-dark-200">Master API Key</label>
                <input
                  type="text"
                  value={nodeForm.masterApiKey}
                  onChange={(e) => setNodeForm(prev => ({ ...prev, masterApiKey: e.target.value }))}
                  placeholder="JWT token from master server"
                  className="mt-1 block w-full bg-dark-700 border border-dark-600 text-dark-100 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-dark-200">Storage Path</label>
                <input
                  type="text"
                  value={nodeForm.storagePath}
                  onChange={(e) => setNodeForm(prev => ({ ...prev, storagePath: e.target.value }))}
                  className="mt-1 block w-full bg-dark-700 border border-dark-600 text-dark-100 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-dark-200">
                  Max Storage ({formatBytes(nodeForm.maxStorage)})
                </label>
                <input
                  type="range"
                  min={1073741824} // 1GB
                  max={1099511627776} // 1TB
                  step={1073741824} // 1GB steps
                  value={nodeForm.maxStorage}
                  onChange={(e) => setNodeForm(prev => ({ ...prev, maxStorage: parseInt(e.target.value) }))}
                  className="mt-1 block w-full"
                />
              </div>

              <button
                type="submit"
                disabled={loading}
                className="w-full bg-green-600 text-white py-2 px-4 rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading ? 'Setting up Node...' : 'Setup Storage Node'}
              </button>
            </form>
          )}
        </div>
      </div>
    </div>
  );
};

export default Setup;