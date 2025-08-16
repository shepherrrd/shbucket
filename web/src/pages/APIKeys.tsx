import { useState, useEffect } from 'react';
import { Plus, Copy, Trash2, Eye, EyeOff, Key, Clock, CheckCircle2, AlertCircle } from 'lucide-react';
import { toast } from 'react-hot-toast';
import { api } from '../services/api';
import type { APIKey, CreateAPIKeyRequest, APIKeyPermission } from '../types';

export default function APIKeys() {
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newApiKey, setNewApiKey] = useState<string>('');

  useEffect(() => {
    loadApiKeys();
  }, []);

  const loadApiKeys = async () => {
    try {
      const response = await api.getAPIKeys();
      setApiKeys(response.api_keys || []);
    } catch (error: any) {
      toast.error('Failed to load API keys: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this API key? This action cannot be undone.')) {
      return;
    }

    try {
      await api.deleteAPIKey(id);
      setApiKeys(apiKeys.filter(key => key.id !== id));
      toast.success('API key deleted successfully');
    } catch (error: any) {
      toast.error('Failed to delete API key: ' + error.message);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const getStatusColor = (key: APIKey) => {
    if (!key.is_active) return 'text-red-500';
    if (key.expires_at && new Date(key.expires_at) < new Date()) return 'text-orange-500';
    return 'text-green-500';
  };

  const getStatusText = (key: APIKey) => {
    if (!key.is_active) return 'Inactive';
    if (key.expires_at && new Date(key.expires_at) < new Date()) return 'Expired';
    return 'Active';
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">API Keys</h1>
          <p className="text-gray-600 dark:text-gray-400">
            Manage API keys for programmatic access to your buckets
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          <Plus className="w-4 h-4 mr-2" />
          Create API Key
        </button>
      </div>

      {apiKeys.length === 0 ? (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-12 text-center">
          <Key className="w-12 h-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">
            No API keys found
          </h3>
          <p className="text-gray-600 dark:text-gray-400 mb-4">
            Create your first API key to start accessing SHBucket programmatically.
          </p>
          <button
            onClick={() => setShowCreateModal(true)}
            className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700"
          >
            <Plus className="w-4 h-4 mr-2" />
            Create API Key
          </button>
        </div>
      ) : (
        <div className="bg-white dark:bg-gray-800 shadow overflow-hidden rounded-md">
          <ul className="divide-y divide-gray-200 dark:divide-gray-700">
            {apiKeys.map((apiKey) => (
              <li key={apiKey.id} className="px-6 py-4">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <div className="flex items-center">
                      <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                        {apiKey.name}
                      </h3>
                      <span className={`ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(apiKey)}`}>
                        {getStatusText(apiKey) === 'Active' && <CheckCircle2 className="w-3 h-3 mr-1" />}
                        {getStatusText(apiKey) === 'Expired' && <AlertCircle className="w-3 h-3 mr-1" />}
                        {getStatusText(apiKey) === 'Inactive' && <AlertCircle className="w-3 h-3 mr-1" />}
                        {getStatusText(apiKey)}
                      </span>
                    </div>
                    <div className="mt-1">
                      <p className="text-sm text-gray-600 dark:text-gray-400">
                        Key prefix: <span className="font-mono">{apiKey.key_prefix}***</span>
                      </p>
                      <div className="flex items-center space-x-4 mt-2 text-xs text-gray-500 dark:text-gray-400">
                        <span>Created: {formatDate(apiKey.created_at)}</span>
                        {apiKey.expires_at && (
                          <span className="flex items-center">
                            <Clock className="w-3 h-3 mr-1" />
                            Expires: {formatDate(apiKey.expires_at)}
                          </span>
                        )}
                        {apiKey.last_used && (
                          <span>Last used: {formatDate(apiKey.last_used)}</span>
                        )}
                      </div>
                    </div>
                    <div className="mt-2">
                      <div className="flex items-center space-x-4 text-xs">
                        <span className={`px-2 py-1 rounded ${apiKey.permissions.read ? 'bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-100' : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-100'}`}>
                          Read: {apiKey.permissions.read ? 'Yes' : 'No'}
                        </span>
                        <span className={`px-2 py-1 rounded ${apiKey.permissions.write ? 'bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-100' : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-100'}`}>
                          Write: {apiKey.permissions.write ? 'Yes' : 'No'}
                        </span>
                        <span className={`px-2 py-1 rounded ${apiKey.permissions.sign_urls ? 'bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-100' : 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-100'}`}>
                          Sign URLs: {apiKey.permissions.sign_urls ? 'Yes' : 'No'}
                        </span>
                        {(apiKey.permissions.buckets && apiKey.permissions.buckets.length > 0) && (
                          <span className="px-2 py-1 bg-blue-100 text-blue-800 dark:bg-blue-800 dark:text-blue-100 rounded">
                            Buckets: {apiKey.permissions.buckets.length}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    <button
                      onClick={() => handleDelete(apiKey.id)}
                      className="inline-flex items-center p-2 border border-transparent rounded-md text-red-700 bg-red-100 hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 dark:bg-red-800 dark:text-red-100 dark:hover:bg-red-700"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}

      {showCreateModal && (
        <CreateAPIKeyModal
          onClose={() => setShowCreateModal(false)}
          onCreated={(key, apiKeyData) => {
            setApiKeys([...apiKeys, apiKeyData]);
            setNewApiKey(key);
            setShowCreateModal(false);
          }}
        />
      )}

      {newApiKey && (
        <NewAPIKeyModal
          apiKey={newApiKey}
          onClose={() => setNewApiKey('')}
        />
      )}
    </div>
  );
}

interface CreateAPIKeyModalProps {
  onClose: () => void;
  onCreated: (key: string, apiKey: APIKey) => void;
}

function CreateAPIKeyModal({ onClose, onCreated }: CreateAPIKeyModalProps) {
  const [name, setName] = useState('');
  const [permissions, setPermissions] = useState<APIKeyPermission>({
    read: true,
    write: false,
    sign_urls: false,
    buckets: []
  });
  const [expiresIn, setExpiresIn] = useState<string>('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      toast.error('Please enter a name for the API key');
      return;
    }

    setLoading(true);
    try {
      const request: CreateAPIKeyRequest = {
        name: name.trim(),
        permissions,
      };

      if (expiresIn) {
        request.expires_in = parseInt(expiresIn) * 24 * 60 * 60; // Convert days to seconds
      }

      const response = await api.createAPIKey(request);
      toast.success('API key created successfully');
      onCreated(response.key, response.api_key);
    } catch (error: any) {
      toast.error('Failed to create API key: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white dark:bg-gray-800">
        <div className="mt-3">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-medium text-gray-900 dark:text-white">
              Create API Key
            </h3>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
            >
              ×
            </button>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Name
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2 text-gray-900 placeholder-gray-500 focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white dark:placeholder-gray-400"
                placeholder="My API Key"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Permissions
              </label>
              <div className="space-y-2">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={permissions.read}
                    onChange={(e) => setPermissions({...permissions, read: e.target.checked})}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                  <span className="ml-2 text-sm text-gray-900 dark:text-white">Read files</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={permissions.write}
                    onChange={(e) => setPermissions({...permissions, write: e.target.checked})}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                  <span className="ml-2 text-sm text-gray-900 dark:text-white">Upload/delete files</span>
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={permissions.sign_urls}
                    onChange={(e) => setPermissions({...permissions, sign_urls: e.target.checked})}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                  />
                  <span className="ml-2 text-sm text-gray-900 dark:text-white">Generate signed URLs</span>
                </label>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Expires in (days, optional)
              </label>
              <input
                type="number"
                value={expiresIn}
                onChange={(e) => setExpiresIn(e.target.value)}
                className="mt-1 block w-full border border-gray-300 rounded-md px-3 py-2 text-gray-900 placeholder-gray-500 focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white dark:placeholder-gray-400"
                placeholder="30"
                min="1"
                max="365"
              />
              <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                Leave empty for no expiration
              </p>
            </div>

            <div className="flex justify-end space-x-3 pt-4">
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white dark:hover:bg-gray-600"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={loading}
                className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
              >
                {loading ? 'Creating...' : 'Create'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

interface NewAPIKeyModalProps {
  apiKey: string;
  onClose: () => void;
}

function NewAPIKeyModal({ apiKey, onClose }: NewAPIKeyModalProps) {
  const [showKey, setShowKey] = useState(false);
  const [copied, setCopied] = useState(false);

  const copyToClipboard = () => {
    navigator.clipboard.writeText(apiKey);
    setCopied(true);
    toast.success('API key copied to clipboard');
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white dark:bg-gray-800">
        <div className="mt-3">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-medium text-gray-900 dark:text-white">
              API Key Created
            </h3>
          </div>

          <div className="space-y-4">
            <div className="p-4 bg-yellow-50 dark:bg-yellow-900 border border-yellow-200 dark:border-yellow-700 rounded-md">
              <div className="flex">
                <AlertCircle className="h-5 w-5 text-yellow-400" />
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                    Important!
                  </h3>
                  <p className="mt-1 text-sm text-yellow-700 dark:text-yellow-300">
                    This is the only time you'll see this API key. Make sure to copy it and store it securely.
                  </p>
                </div>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                API Key
              </label>
              <div className="flex items-center space-x-2">
                <div className="flex-1 relative">
                  <input
                    type={showKey ? 'text' : 'password'}
                    value={apiKey}
                    readOnly
                    className="block w-full border border-gray-300 rounded-md px-3 py-2 text-gray-900 bg-gray-50 focus:outline-none dark:bg-gray-700 dark:border-gray-600 dark:text-white font-mono text-sm"
                  />
                  <button
                    type="button"
                    onClick={() => setShowKey(!showKey)}
                    className="absolute inset-y-0 right-0 pr-3 flex items-center"
                  >
                    {showKey ? (
                      <EyeOff className="h-4 w-4 text-gray-400" />
                    ) : (
                      <Eye className="h-4 w-4 text-gray-400" />
                    )}
                  </button>
                </div>
                <button
                  onClick={copyToClipboard}
                  className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white dark:hover:bg-gray-600"
                >
                  <Copy className="w-4 h-4" />
                  {copied && <span className="ml-1">✓</span>}
                </button>
              </div>
            </div>

            <div className="flex justify-end pt-4">
              <button
                onClick={onClose}
                className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              >
                Done
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}