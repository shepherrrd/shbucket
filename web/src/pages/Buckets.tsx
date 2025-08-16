import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate } from 'react-router-dom';
import { apiClient } from '../services/api';
import toast from 'react-hot-toast';
import {
  FolderIcon,
  PlusIcon,
  PencilIcon,
  TrashIcon,
  ArrowLeftIcon,
} from '@heroicons/react/24/outline';
import type { Bucket, CreateBucketRequest, UpdateBucketRequest } from '../types';

export default function Buckets() {
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [selectedBucket, setSelectedBucket] = useState<Bucket | null>(null);
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: bucketsData, isLoading, error } = useQuery({
    queryKey: ['buckets'],
    queryFn: () => apiClient.getBuckets(),
  });

  const createBucketMutation = useMutation({
    mutationFn: (bucket: CreateBucketRequest) => apiClient.createBucket(bucket),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['buckets'] });
      setShowCreateModal(false);
      toast.success('Bucket created successfully!');
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to create bucket');
    },
  });

  const updateBucketMutation = useMutation({
    mutationFn: ({ bucketId, updates }: { bucketId: string; updates: UpdateBucketRequest }) => 
      apiClient.updateBucket(bucketId, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['buckets'] });
      setShowEditModal(false);
      setSelectedBucket(null);
      toast.success('Bucket updated successfully!');
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to update bucket');
    },
  });

  const deleteBucketMutation = useMutation({
    mutationFn: (bucketId: string) => apiClient.deleteBucket(bucketId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['buckets'] });
      toast.success('Bucket deleted successfully!');
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to delete bucket');
    },
  });


  const buckets = bucketsData?.buckets || [];

  const handleCreateBucket = (e: React.FormEvent) => {
    e.preventDefault();
    const formData = new FormData(e.target as HTMLFormElement);
    const bucket: CreateBucketRequest = {
      name: formData.get('name') as string,
      description: formData.get('description') as string,
      auth_rule: {
        type: 'none',
        enabled: false,
        config: {},
      },
      settings: {
        max_file_size: 100 * 1024 * 1024, // 100MB
        max_total_size: 1024 * 1024 * 1024, // 1GB
        allowed_mime_types: [],
        blocked_mime_types: [],
        allowed_extensions: [],
        blocked_extensions: [],
        max_files_per_bucket: 1000,
        public_read: formData.get('public_read') === 'on',
        versioning: false,
        encryption: false,
        allow_overwrite: true,
        require_content_type: false,
      },
    };
    createBucketMutation.mutate(bucket);
  };

  const handleDeleteBucket = (bucket: Bucket) => {
    if (confirm(`Are you sure you want to delete bucket "${bucket.name}"?`)) {
      deleteBucketMutation.mutate(bucket.id);
    }
  };

  const handleEditBucket = (bucket: Bucket) => {
    setSelectedBucket(bucket);
    setShowEditModal(true);
  };

  const handleUpdateBucket = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedBucket) return;
    
    const formData = new FormData(e.target as HTMLFormElement);
    const updates: UpdateBucketRequest = {
      description: formData.get('description') as string,
      settings: {
        ...selectedBucket.settings,
        public_read: formData.get('public_read') === 'on',
        max_file_size: parseInt(formData.get('max_file_size') as string) * 1024 * 1024, // Convert MB to bytes
        versioning: formData.get('versioning') === 'on',
        encryption: formData.get('encryption') === 'on',
      },
    };
    updateBucketMutation.mutate({ bucketId: selectedBucket.id, updates });
  };

  const handleViewBucket = (bucket: Bucket) => {
    navigate(`/buckets/${bucket.id}`);
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
        <div className="text-red-400">Error loading buckets: {(error as Error).message}</div>
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
                <h1 className="text-3xl font-bold text-white">Buckets</h1>
                <p className="text-dark-400">Manage your storage buckets</p>
              </div>
            </div>
            <button
              onClick={() => setShowCreateModal(true)}
              className="inline-flex items-center px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors"
            >
              <PlusIcon className="h-5 w-5 mr-2" />
              Create Bucket
            </button>
          </div>
        </div>
      </div>

      {/* Main content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {buckets.length === 0 ? (
          <div className="text-center py-12">
            <FolderIcon className="mx-auto h-12 w-12 text-dark-400" />
            <h3 className="mt-2 text-sm font-medium text-dark-300">No buckets</h3>
            <p className="mt-1 text-sm text-dark-500">Get started by creating a new bucket.</p>
            <div className="mt-6">
              <button
                onClick={() => setShowCreateModal(true)}
                className="inline-flex items-center px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors"
              >
                <PlusIcon className="h-5 w-5 mr-2" />
                Create Bucket
              </button>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {buckets.map((bucket) => (
              <div key={bucket.id} className="bg-dark-900 rounded-lg border border-dark-700 hover:border-primary-500 transition-colors">
                <div 
                  onClick={() => handleViewBucket(bucket)}
                  className="p-6 cursor-pointer"
                >
                  <div className="flex items-center justify-between mb-4">
                    <FolderIcon className="h-8 w-8 text-blue-400" />
                    <div className="flex space-x-2" onClick={(e) => e.stopPropagation()}>
                      <button
                        onClick={() => handleEditBucket(bucket)}
                        className="p-1 text-dark-400 hover:text-white"
                        title="Edit"
                      >
                        <PencilIcon className="h-4 w-4" />
                      </button>
                      <button
                        onClick={() => handleDeleteBucket(bucket)}
                        className="p-1 text-dark-400 hover:text-red-400"
                        title="Delete"
                      >
                        <TrashIcon className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                  <h3 className="text-lg font-semibold text-white mb-2">{bucket.name}</h3>
                  <p className="text-dark-400 text-sm mb-4">{bucket.description || 'No description'}</p>
                  <div className="flex items-center justify-between text-sm text-dark-500">
                    <span>{bucket.stats?.total_files || 0} files</span>
                    <span>{bucket.settings.public_read ? 'Public' : 'Private'}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Create Bucket Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-dark-900 rounded-lg border border-dark-700 p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold text-white mb-4">Create New Bucket</h3>
            <form onSubmit={handleCreateBucket}>
              <div className="space-y-4">
                <div>
                  <label htmlFor="name" className="block text-sm font-medium text-dark-300">
                    Bucket Name
                  </label>
                  <input
                    type="text"
                    id="name"
                    name="name"
                    required
                    className="mt-1 block w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    placeholder="my-bucket"
                  />
                </div>
                <div>
                  <label htmlFor="description" className="block text-sm font-medium text-dark-300">
                    Description
                  </label>
                  <textarea
                    id="description"
                    name="description"
                    rows={3}
                    className="mt-1 block w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    placeholder="Bucket description..."
                  />
                </div>
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    id="public_read"
                    name="public_read"
                    className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-dark-600 rounded bg-dark-800"
                  />
                  <label htmlFor="public_read" className="ml-2 text-sm text-dark-300">
                    Allow public read access
                  </label>
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
                  disabled={createBucketMutation.isPending}
                  className="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors disabled:opacity-50"
                >
                  {createBucketMutation.isPending ? 'Creating...' : 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit Bucket Modal */}
      {showEditModal && selectedBucket && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-dark-900 rounded-lg border border-dark-700 p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold text-white mb-4">Edit Bucket: {selectedBucket.name}</h3>
            <form onSubmit={handleUpdateBucket}>
              <div className="space-y-4">
                <div>
                  <label htmlFor="edit-description" className="block text-sm font-medium text-dark-300">
                    Description
                  </label>
                  <textarea
                    id="edit-description"
                    name="description"
                    rows={3}
                    defaultValue={selectedBucket.description}
                    className="mt-1 block w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                    placeholder="Bucket description..."
                  />
                </div>
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    id="edit-public_read"
                    name="public_read"
                    defaultChecked={selectedBucket.settings.public_read}
                    className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-dark-600 rounded bg-dark-800"
                  />
                  <label htmlFor="edit-public_read" className="ml-2 text-sm text-dark-300">
                    Allow public read access
                  </label>
                </div>
                <div>
                  <label htmlFor="edit-max_file_size" className="block text-sm font-medium text-dark-300">
                    Max File Size (MB)
                  </label>
                  <input
                    type="number"
                    id="edit-max_file_size"
                    name="max_file_size"
                    defaultValue={Math.round(selectedBucket.settings.max_file_size / (1024 * 1024))}
                    min="1"
                    className="mt-1 block w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  />
                </div>
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    id="edit-versioning"
                    name="versioning"
                    defaultChecked={selectedBucket.settings.versioning}
                    className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-dark-600 rounded bg-dark-800"
                  />
                  <label htmlFor="edit-versioning" className="ml-2 text-sm text-dark-300">
                    Enable versioning
                  </label>
                </div>
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    id="edit-encryption"
                    name="encryption"
                    defaultChecked={selectedBucket.settings.encryption}
                    className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-dark-600 rounded bg-dark-800"
                  />
                  <label htmlFor="edit-encryption" className="ml-2 text-sm text-dark-300">
                    Enable encryption
                  </label>
                </div>
              </div>
              <div className="flex justify-end space-x-3 mt-6">
                <button
                  type="button"
                  onClick={() => { setShowEditModal(false); setSelectedBucket(null); }}
                  className="px-4 py-2 text-dark-300 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={updateBucketMutation.isPending}
                  className="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors disabled:opacity-50"
                >
                  {updateBucketMutation.isPending ? 'Updating...' : 'Update'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

    </div>
  );
}