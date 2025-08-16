import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { apiClient } from '../services/api';
import toast from 'react-hot-toast';
import {
  ServerIcon,
  PlusIcon,
  PencilIcon,
  TrashIcon,
  ArrowLeftIcon,
  CheckCircleIcon,
  XCircleIcon,
} from '@heroicons/react/24/outline';
import type { StorageNode, CreateNodeRequest } from '../types';

export default function StorageNodes() {
  const [showCreateModal, setShowCreateModal] = useState(false);
  // const [selectedNode, setSelectedNode] = useState<StorageNode | null>(null);
  const queryClient = useQueryClient();

  const { data: nodesData, isLoading, error } = useQuery({
    queryKey: ['nodes'],
    queryFn: () => apiClient.getNodes(),
  });

  const createNodeMutation = useMutation({
    mutationFn: (node: CreateNodeRequest) => apiClient.createNode(node),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nodes'] });
      setShowCreateModal(false);
      toast.success('Storage node created successfully!');
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to create storage node');
    },
  });

  const deleteNodeMutation = useMutation({
    mutationFn: (id: string) => apiClient.deleteNode(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nodes'] });
      toast.success('Storage node deleted successfully!');
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to delete storage node');
    },
  });

  const checkHealthMutation = useMutation({
    mutationFn: (id: string) => apiClient.checkNodeHealth(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['nodes'] });
      toast.success('Health check completed!');
    },
    onError: (error: any) => {
      toast.error(error.message || 'Health check failed');
    },
  });

  const nodes = nodesData?.nodes || [];

  const handleCreateNode = (e: React.FormEvent) => {
    e.preventDefault();
    const formData = new FormData(e.target as HTMLFormElement);
    const node: CreateNodeRequest = {
      name: formData.get('name') as string,
      url: formData.get('url') as string,
      auth_key: formData.get('auth_key') as string,
      max_storage: parseInt(formData.get('max_storage') as string) || 1000000000, // 1GB default
      priority: parseInt(formData.get('priority') as string) || 1,
      is_active: true,
    };
    createNodeMutation.mutate(node);
  };

  const handleDeleteNode = (node: StorageNode) => {
    if (confirm(`Are you sure you want to delete storage node "${node.name}"?`)) {
      deleteNodeMutation.mutate(node.id);
    }
  };

  const getHealthStatusIcon = (isHealthy: boolean) => {
    if (isHealthy) {
      return <CheckCircleIcon className="h-5 w-5 text-green-400" />;
    } else {
      return <XCircleIcon className="h-5 w-5 text-red-400" />;
    }
  };

  const getHealthStatusColor = (isHealthy: boolean) => {
    return isHealthy ? 'text-green-400' : 'text-red-400';
  };

  const getHealthStatusText = (isHealthy: boolean) => {
    return isHealthy ? 'Healthy' : 'Unhealthy';
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-dark-950 flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-dark-950 flex items-center justify-center">
        <div className="text-red-400">Error loading storage nodes: {(error as Error).message}</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-dark-950">
      {/* Header */}
      <div className="bg-dark-900 border-b border-dark-700">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <div className="flex items-center space-x-4">
              <Link
                to="/"
                className="p-2 text-dark-400 hover:text-white rounded-md hover:bg-dark-800 transition-colors"
              >
                <ArrowLeftIcon className="h-6 w-6" />
              </Link>
              <div>
                <h1 className="text-3xl font-bold text-white">Storage Nodes</h1>
                <p className="text-dark-400">Manage your storage infrastructure</p>
              </div>
            </div>
            <div className="flex space-x-3">
              <button
                onClick={() => apiClient.checkAllNodesHealth()}
                className="inline-flex items-center px-4 py-2 bg-green-600 hover:bg-green-700 text-white font-medium rounded-lg transition-colors"
              >
                Check All Health
              </button>
              <button
                onClick={() => setShowCreateModal(true)}
                className="inline-flex items-center px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors"
              >
                <PlusIcon className="h-5 w-5 mr-2" />
                Add Node
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Main content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {nodes.length === 0 ? (
          <div className="text-center py-12">
            <ServerIcon className="mx-auto h-12 w-12 text-dark-400" />
            <h3 className="mt-2 text-sm font-medium text-dark-300">No storage nodes</h3>
            <p className="mt-1 text-sm text-dark-500">Get started by adding your first storage node.</p>
            <div className="mt-6">
              <button
                onClick={() => setShowCreateModal(true)}
                className="inline-flex items-center px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors"
              >
                <PlusIcon className="h-5 w-5 mr-2" />
                Add Node
              </button>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {nodes.map((node) => (
              <div key={node.id} className="bg-dark-900 rounded-lg border border-dark-700 p-6">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center space-x-3">
                    <ServerIcon className="h-8 w-8 text-green-400" />
                    <div>
                      <h3 className="text-lg font-semibold text-white">{node.name}</h3>
                      <p className="text-sm text-dark-400">{node.url}</p>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    {getHealthStatusIcon(node.is_healthy)}
                    <span className={`text-sm ${getHealthStatusColor(node.is_healthy)}`}>
                      {getHealthStatusText(node.is_healthy)}
                    </span>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-4 mb-4">
                  <div>
                    <p className="text-sm text-dark-500">Priority</p>
                    <p className="text-white">{node.priority}</p>
                  </div>
                  <div>
                    <p className="text-sm text-dark-500">Status</p>
                    <p className="text-white">{node.is_active ? 'Active' : 'Inactive'}</p>
                  </div>
                  <div>
                    <p className="text-sm text-dark-500">Max Storage</p>
                    <p className="text-white">{formatBytes(node.max_storage)}</p>
                  </div>
                  <div>
                    <p className="text-sm text-dark-500">Used Storage</p>
                    <p className="text-white">{formatBytes(node.used_storage)}</p>
                  </div>
                </div>

                <div className="mb-4">
                  <div className="flex justify-between text-sm text-dark-400 mb-1">
                    <span>Storage Usage</span>
                    <span>{node.max_storage > 0 ? Math.round((node.used_storage / node.max_storage) * 100) : 0}%</span>
                  </div>
                  <div className="w-full bg-dark-700 rounded-full h-2">
                    <div
                      className="bg-primary-500 h-2 rounded-full"
                      style={{ width: `${node.max_storage > 0 ? Math.min((node.used_storage / node.max_storage) * 100, 100) : 0}%` }}
                    ></div>
                  </div>
                </div>

                <div className="flex justify-between items-center">
                  <div className="flex items-center space-x-2">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      node.is_active 
                        ? 'bg-green-100 text-green-800' 
                        : 'bg-red-100 text-red-800'
                    }`}>
                      {node.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </div>
                  <div className="flex space-x-2">
                    <button
                      onClick={() => checkHealthMutation.mutate(node.id)}
                      disabled={checkHealthMutation.isPending}
                      className="p-1 text-dark-400 hover:text-green-400"
                      title="Check Health"
                    >
                      <CheckCircleIcon className="h-4 w-4" />
                    </button>
                    <button
                      onClick={() => {/* TODO: Implement edit functionality */}}
                      className="p-1 text-dark-400 hover:text-white"
                      title="Edit"
                    >
                      <PencilIcon className="h-4 w-4" />
                    </button>
                    <button
                      onClick={() => handleDeleteNode(node)}
                      className="p-1 text-dark-400 hover:text-red-400"
                      title="Delete"
                    >
                      <TrashIcon className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Create Node Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-dark-900 rounded-lg border border-dark-700 p-6 w-full max-w-lg max-h-[90vh] overflow-y-auto">
            <h3 className="text-lg font-semibold text-white mb-4">Add Storage Node</h3>
            <form onSubmit={handleCreateNode}>
              <div className="space-y-4">
                <div>
                  <label htmlFor="name" className="block text-sm font-medium text-dark-300">
                    Node Name
                  </label>
                  <input
                    type="text"
                    id="name"
                    name="name"
                    required
                    className="mt-1 block w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    placeholder="My Storage Node"
                  />
                </div>
                <div>
                  <label htmlFor="url" className="block text-sm font-medium text-dark-300">
                    Node URL
                  </label>
                  <input
                    type="url"
                    id="url"
                    name="url"
                    required
                    className="mt-1 block w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    placeholder="http://localhost:8081"
                  />
                </div>
                <div>
                  <label htmlFor="auth_key" className="block text-sm font-medium text-dark-300">
                    Authentication Key
                  </label>
                  <input
                    type="text"
                    id="auth_key"
                    name="auth_key"
                    required
                    minLength={32}
                    className="mt-1 block w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    placeholder="Enter a secure 32+ character authentication key"
                  />
                  <p className="mt-1 text-xs text-dark-500">
                    This key will be used to authenticate requests from the master to this node (minimum 32 characters)
                  </p>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label htmlFor="max_storage" className="block text-sm font-medium text-dark-300">
                      Max Storage (Bytes)
                    </label>
                    <input
                      type="number"
                      id="max_storage"
                      name="max_storage"
                      min="1"
                      defaultValue="1073741824"
                      className="mt-1 block w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    />
                    <p className="mt-1 text-xs text-dark-500">1073741824 = 1GB</p>
                  </div>
                  <div>
                    <label htmlFor="priority" className="block text-sm font-medium text-dark-300">
                      Priority
                    </label>
                    <input
                      type="number"
                      id="priority"
                      name="priority"
                      min="0"
                      max="100"
                      defaultValue="1"
                      className="mt-1 block w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    />
                    <p className="mt-1 text-xs text-dark-500">Higher number = higher priority</p>
                  </div>
                </div>
              </div>
              <div className="flex justify-end space-x-3 mt-6">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 text-dark-300 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={createNodeMutation.isPending}
                  className="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors disabled:opacity-50"
                >
                  {createNodeMutation.isPending ? 'Adding...' : 'Add Node'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}