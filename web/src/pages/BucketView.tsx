import { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../services/api';
import toast from 'react-hot-toast';
import ImageView from '../components/ImageView';
import {
  ArrowLeftIcon,
  ArrowUpTrayIcon,
  TrashIcon,
  DocumentIcon,
  PhotoIcon,
  FolderIcon,
  InformationCircleIcon,
  LinkIcon,
} from '@heroicons/react/24/outline';
import SignedURLDialog from '../components/SignedURLDialog';

interface FileWithPreview {
  id: string;
  bucket_id: string;
  name: string;
  path: string;
  size: number;
  mime_type: string;
  extension: string;
  checksum: string;
  metadata: Record<string, any>;
  secured_url?: string;
  created_at: string;
  updated_at: string;
  preview?: string;
}

export default function BucketView() {
  const { bucketId } = useParams<{ bucketId: string }>();
  const queryClient = useQueryClient();
  const [selectedFile, setSelectedFile] = useState<FileWithPreview | null>(null);
  const [showUploadModal, setShowUploadModal] = useState(false);
  const [showSignedURLDialog, setShowSignedURLDialog] = useState(false);

  const { data: bucket, isLoading: bucketLoading } = useQuery({
    queryKey: ['bucket', bucketId],
    queryFn: () => bucketId ? apiClient.getBucket(bucketId) : null,
    enabled: !!bucketId,
  });

  const { data: filesData, isLoading: filesLoading } = useQuery({
    queryKey: ['files', bucketId],
    queryFn: () => bucketId ? apiClient.getFiles(bucketId) : null,
    enabled: !!bucketId,
  });

  const uploadFileMutation = useMutation({
    mutationFn: ({ bucketId, file }: { bucketId: string; file: globalThis.File }) => 
      apiClient.uploadFile(bucketId, file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['files', bucketId] });
      setShowUploadModal(false);
      toast.success('File uploaded successfully!');
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to upload file');
    },
  });

  const deleteFileMutation = useMutation({
    mutationFn: ({ bucketId, fileId }: { bucketId: string; fileId: string }) => 
      apiClient.deleteFile(bucketId, fileId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['files', bucketId] });
      setSelectedFile(null);
      toast.success('File deleted successfully!');
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to delete file');
    },
  });

  const files: FileWithPreview[] = filesData?.files || [];

  const handleUploadFile = (e: React.FormEvent) => {
    e.preventDefault();
    if (!bucketId) return;
    
    const formData = new FormData(e.target as HTMLFormElement);
    const file = formData.get('file') as globalThis.File;
    if (file) {
      uploadFileMutation.mutate({ bucketId, file });
    }
  };

  const handleDeleteFile = (file: FileWithPreview) => {
    if (!bucketId) return;
    
    if (confirm(`Are you sure you want to delete "${file.name}"?`)) {
      deleteFileMutation.mutate({ bucketId, fileId: file.id });
    }
  };

  const isImageFile = (mimeType: string) => {
    return mimeType.startsWith('image/');
  };

  const getFileIcon = (mimeType: string) => {
    if (isImageFile(mimeType)) {
      return <PhotoIcon className="h-6 w-6 text-blue-400" />;
    }
    return <DocumentIcon className="h-6 w-6 text-gray-400" />;
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (bucketLoading || filesLoading) {
    return (
      <div className="min-h-screen bg-dark-950 flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500"></div>
      </div>
    );
  }

  if (!bucket) {
    return (
      <div className="min-h-screen bg-dark-950 flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-xl font-semibold text-white mb-2">Bucket not found</h2>
          <Link
            to="/buckets"
            className="inline-flex items-center px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition-colors"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-2" />
            Back to Buckets
          </Link>
        </div>
      </div>
    );
  }

  return (
      <div className="min-h-screen bg-dark-950">
      {/* Header */}
      <div className="bg-dark-900 border-b border-dark-700">
        <div className="max-w-full mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            <div className="flex items-center space-x-4">
              <Link
                to="/buckets"
                className="p-2 text-dark-400 hover:text-white rounded-md hover:bg-dark-800 transition-colors"
              >
                <ArrowLeftIcon className="h-6 w-6" />
              </Link>
              <div className="flex items-center space-x-3">
                <FolderIcon className="h-8 w-8 text-blue-400" />
                <div>
                  <h1 className="text-2xl font-bold text-white">{bucket.name}</h1>
                  <p className="text-dark-400 text-sm">{bucket.description || 'No description'}</p>
                </div>
              </div>
            </div>
            <button
              onClick={() => setShowUploadModal(true)}
              className="inline-flex items-center px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors"
            >
              <ArrowUpTrayIcon className="h-5 w-5 mr-2" />
              Upload File
            </button>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex h-[calc(100vh-80px)]">
        {/* Files Grid */}
        <div className="flex-1 p-6 overflow-y-auto">
          {files.length === 0 ? (
            <div className="text-center py-12">
              <DocumentIcon className="mx-auto h-16 w-16 text-dark-400" />
              <h3 className="mt-4 text-lg font-medium text-dark-300">No files</h3>
              <p className="mt-2 text-dark-500">Upload your first file to get started.</p>
              <button
                onClick={() => setShowUploadModal(true)}
                className="mt-4 inline-flex items-center px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors"
              >
                <ArrowUpTrayIcon className="h-5 w-5 mr-2" />
                Upload File
              </button>
            </div>
          ) : (
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8 gap-4">
              {files.map((file) => (
                <div
                  key={file.id}
                  onClick={() => setSelectedFile(file)}
                  className={`group cursor-pointer rounded-lg border-2 transition-all ${
                    selectedFile?.id === file.id
                      ? 'border-primary-500 bg-primary-500/10'
                      : 'border-dark-700 hover:border-dark-600 bg-dark-800 hover:bg-dark-700'
                  }`}
                >
                  <div className="aspect-square p-3 flex flex-col items-center justify-center">
                    {isImageFile(file.mime_type) ? (
                      <div className="w-full h-full rounded-md overflow-hidden bg-dark-700">
                        <ImageView
                          bucketId={bucketId!}
                          fileId={file.id}
                          alt={file.name}
                          className="w-full h-full object-cover"
                        />
                      </div>
                    ) : (
                      <div className="w-full h-full rounded-md bg-dark-700 flex items-center justify-center">
                        {getFileIcon(file.mime_type)}
                      </div>
                    )}
                  </div>
                  <div className="px-3 pb-3">
                    <p className="text-sm font-medium text-white truncate" title={file.name}>
                      {file.name}
                    </p>
                    <p className="text-xs text-dark-400">
                      {formatFileSize(file.size)}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* File Details Panel */}
        {selectedFile && (
          <div className="w-96 bg-dark-900 border-l border-dark-700 overflow-y-auto">
            <div className="p-6">
              {/* File Preview */}
              <div className="mb-6">
                {isImageFile(selectedFile.mime_type) ? (
                  <div className="aspect-square rounded-lg overflow-hidden bg-dark-800">
                    <ImageView
                      bucketId={bucketId!}
                      fileId={selectedFile.id}
                      alt={selectedFile.name}
                      className="w-full h-full object-contain"
                    />
                  </div>
                ) : (
                  <div className="aspect-square rounded-lg bg-dark-800 flex items-center justify-center">
                    {getFileIcon(selectedFile.mime_type)}
                  </div>
                )}
              </div>

              {/* File Information */}
              <div className="space-y-4">
                <div>
                  <h3 className="text-lg font-semibold text-white mb-3 flex items-center">
                    <InformationCircleIcon className="h-5 w-5 mr-2" />
                    File Details
                  </h3>
                </div>

                <div>
                  <label className="block text-sm font-medium text-dark-300 mb-1">Name</label>
                  <p className="text-white break-words">{selectedFile.name}</p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-dark-300 mb-1">Size</label>
                  <p className="text-white">{formatFileSize(selectedFile.size)}</p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-dark-300 mb-1">Type</label>
                  <p className="text-white">{selectedFile.mime_type}</p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-dark-300 mb-1">Extension</label>
                  <p className="text-white">{selectedFile.extension}</p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-dark-300 mb-1">Upload Date</label>
                  <p className="text-white">{formatDate(selectedFile.created_at)}</p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-dark-300 mb-1">Checksum</label>
                  <p className="text-xs text-dark-400 font-mono break-all">{selectedFile.checksum}</p>
                </div>

                {/* Actions */}
                <div className="pt-4 border-t border-dark-700 space-y-3">
                  <button
                    onClick={async () => {
                      try {
                        const blob = await apiClient.downloadFile(bucketId!, selectedFile.id);
                        const url = window.URL.createObjectURL(blob);
                        const a = document.createElement('a');
                        a.href = url;
                        a.download = selectedFile.name;
                        document.body.appendChild(a);
                        a.click();
                        window.URL.revokeObjectURL(url);
                        document.body.removeChild(a);
                      } catch (error) {
                        toast.error('Failed to download file');
                      }
                    }}
                    className="block w-full px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white text-center font-medium rounded-lg transition-colors"
                  >
                    Download
                  </button>
                  <button
                    onClick={() => setShowSignedURLDialog(true)}
                    className="block w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-center font-medium rounded-lg transition-colors flex items-center justify-center"
                  >
                    <LinkIcon className="h-4 w-4 mr-2" />
                    Generate Signed URL
                  </button>
                  <button
                    onClick={() => handleDeleteFile(selectedFile)}
                    disabled={deleteFileMutation.isPending}
                    className="w-full flex items-center justify-center px-4 py-2 bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg transition-colors disabled:opacity-50"
                  >
                    <TrashIcon className="h-4 w-4 mr-2" />
                    {deleteFileMutation.isPending ? 'Deleting...' : 'Delete'}
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Upload Modal */}
      {showUploadModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-dark-900 rounded-lg border border-dark-700 p-6 w-full max-w-md">
            <h3 className="text-lg font-semibold text-white mb-4">Upload File to {bucket.name}</h3>
            <form onSubmit={handleUploadFile}>
              <div className="space-y-4">
                <div>
                  <label htmlFor="file" className="block text-sm font-medium text-dark-300">
                    Select File
                  </label>
                  <input
                    type="file"
                    id="file"
                    name="file"
                    required
                    className="mt-1 block w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white file:mr-4 file:py-1 file:px-2 file:rounded file:border-0 file:text-sm file:bg-primary-600 file:text-white hover:file:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                  />
                </div>
              </div>
              <div className="flex justify-end space-x-3 mt-6">
                <button
                  type="button"
                  onClick={() => setShowUploadModal(false)}
                  className="px-4 py-2 text-dark-300 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={uploadFileMutation.isPending}
                  className="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors disabled:opacity-50"
                >
                  {uploadFileMutation.isPending ? 'Uploading...' : 'Upload'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Signed URL Dialog */}
      {selectedFile && (
        <SignedURLDialog
          file={selectedFile}
          isOpen={showSignedURLDialog}
          onClose={() => setShowSignedURLDialog(false)}
        />
      )}
      </div>
  );
}