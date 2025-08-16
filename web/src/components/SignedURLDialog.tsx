import { useState } from 'react';
import { Copy, ExternalLink, Clock, AlertCircle } from 'lucide-react';
import { toast } from 'react-hot-toast';
import { api } from '../services/api';
import type { File } from '../types';

interface SignedURLDialogProps {
  file: File;
  isOpen: boolean;
  onClose: () => void;
}

export default function SignedURLDialog({ file, isOpen, onClose }: SignedURLDialogProps) {
  const [expiresIn, setExpiresIn] = useState<string>('3600'); // Default 1 hour
  const [singleUse, setSingleUse] = useState<boolean>(false); // Default multi-use
  const [signedURL, setSignedURL] = useState<string>('');
  const [expiresAt, setExpiresAt] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);

  if (!isOpen) return null;

  const generateURL = async () => {
    if (!expiresIn || parseInt(expiresIn) < 60) {
      toast.error('Expiration must be at least 60 seconds');
      return;
    }

    setLoading(true);
    try {
      const response = await api.generateSignedURL(
        file.bucket_id,
        file.id,
        parseInt(expiresIn),
        singleUse
      );
      
      setSignedURL(response.url);
      setExpiresAt(response.expires_at);
      toast.success('Signed URL generated successfully');
    } catch (error: any) {
      toast.error('Failed to generate signed URL: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = () => {
    navigator.clipboard.writeText(signedURL);
    setCopied(true);
    toast.success('Signed URL copied to clipboard');
    setTimeout(() => setCopied(false), 2000);
  };

  const openURL = () => {
    window.open(signedURL, '_blank');
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  };

  const getExpirationOptions = () => [
    { value: '300', label: '5 minutes' },
    { value: '1800', label: '30 minutes' },
    { value: '3600', label: '1 hour' },
    { value: '21600', label: '6 hours' },
    { value: '86400', label: '1 day' },
    { value: '604800', label: '1 week' },
  ];

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-20 mx-auto p-5 border w-[500px] shadow-lg rounded-md bg-white dark:bg-gray-800">
        <div className="mt-3">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-medium text-gray-900 dark:text-white">
              Generate Signed URL
            </h3>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-xl"
            >
              ×
            </button>
          </div>

          <div className="space-y-4">
            {/* File Info */}
            <div className="p-3 bg-gray-50 dark:bg-gray-700 rounded-md">
              <div className="flex items-center space-x-3">
                <div className="flex-shrink-0">
                  {file.mime_type?.startsWith('image/') ? (
                    <img
                      src={`/api/v1/file/${file.bucket_id}/${file.id}`}
                      alt={file.name}
                      className="h-10 w-10 rounded object-cover"
                    />
                  ) : (
                    <div className="h-10 w-10 bg-blue-100 dark:bg-blue-800 rounded flex items-center justify-center">
                      <span className="text-blue-600 dark:text-blue-300 font-medium text-sm">
                        {file.extension?.toUpperCase() || '?'}
                      </span>
                    </div>
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                    {file.name}
                  </p>
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    {(file.size / 1024 / 1024).toFixed(2)} MB • {file.mime_type}
                  </p>
                </div>
              </div>
            </div>

            {/* Expiration Setting */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                URL Expiration
              </label>
              <div className="grid grid-cols-2 gap-2">
                {getExpirationOptions().map((option) => (
                  <button
                    key={option.value}
                    onClick={() => setExpiresIn(option.value)}
                    className={`p-2 text-sm border rounded-md ${
                      expiresIn === option.value
                        ? 'border-blue-500 bg-blue-50 text-blue-700 dark:bg-blue-900 dark:text-blue-300'
                        : 'border-gray-300 hover:bg-gray-50 dark:border-gray-600 dark:hover:bg-gray-700 dark:text-white'
                    }`}
                  >
                    {option.label}
                  </button>
                ))}
              </div>
              <div className="mt-2">
                <input
                  type="number"
                  value={expiresIn}
                  onChange={(e) => setExpiresIn(e.target.value)}
                  placeholder="Custom seconds"
                  className="block w-full border border-gray-300 rounded-md px-3 py-2 text-gray-900 placeholder-gray-500 focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white dark:placeholder-gray-400"
                  min="60"
                  max="604800"
                />
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  Minimum 60 seconds, maximum 1 week (604800 seconds)
                </p>
              </div>
            </div>

            {/* Single-use setting */}
            <div>
              <label className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={singleUse}
                  onChange={(e) => setSingleUse(e.target.checked)}
                  className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                />
                <div>
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                    Single-use URL
                  </span>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    URL becomes invalid after first access
                  </p>
                </div>
              </label>
            </div>

            {/* Info Alert */}
            <div className="p-3 bg-blue-50 dark:bg-blue-900 border border-blue-200 dark:border-blue-700 rounded-md">
              <div className="flex">
                <AlertCircle className="h-5 w-5 text-blue-400 flex-shrink-0" />
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-blue-800 dark:text-blue-200">
                    About Signed URLs
                  </h3>
                  <p className="mt-1 text-sm text-blue-700 dark:text-blue-300">
                    Signed URLs provide temporary access to files without requiring authentication. 
                    They automatically expire after the specified time.
                  </p>
                </div>
              </div>
            </div>

            {/* Generate Button */}
            {!signedURL && (
              <div className="flex justify-end">
                <button
                  onClick={generateURL}
                  disabled={loading}
                  className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {loading ? 'Generating...' : 'Generate Signed URL'}
                </button>
              </div>
            )}

            {/* Generated URL */}
            {signedURL && (
              <div className="space-y-3">
                <div className="p-3 bg-green-50 dark:bg-green-900 border border-green-200 dark:border-green-700 rounded-md">
                  <div className="flex">
                    <div className="flex-shrink-0">
                      <Clock className="h-5 w-5 text-green-400" />
                    </div>
                    <div className="ml-3">
                      <h3 className="text-sm font-medium text-green-800 dark:text-green-200">
                        URL Generated Successfully
                      </h3>
                      <p className="mt-1 text-sm text-green-700 dark:text-green-300">
                        Expires: {formatDate(expiresAt)}
                      </p>
                    </div>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Signed URL
                  </label>
                  <div className="flex items-center space-x-2">
                    <input
                      type="text"
                      value={signedURL}
                      readOnly
                      className="flex-1 border border-gray-300 rounded-md px-3 py-2 text-gray-900 bg-gray-50 focus:outline-none dark:bg-gray-700 dark:border-gray-600 dark:text-white font-mono text-sm"
                    />
                    <button
                      onClick={copyToClipboard}
                      className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white dark:hover:bg-gray-600"
                    >
                      <Copy className="w-4 h-4" />
                      {copied && <span className="ml-1 text-green-600">✓</span>}
                    </button>
                    <button
                      onClick={openURL}
                      className="inline-flex items-center px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white dark:hover:bg-gray-600"
                    >
                      <ExternalLink className="w-4 h-4" />
                    </button>
                  </div>
                </div>

                <div className="flex justify-between pt-4">
                  <button
                    onClick={() => {
                      setSignedURL('');
                      setExpiresAt('');
                      setCopied(false);
                      setSingleUse(false);
                    }}
                    className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white dark:hover:bg-gray-600"
                  >
                    Generate New URL
                  </button>
                  <button
                    onClick={onClose}
                    className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                  >
                    Done
                  </button>
                </div>
              </div>
            )}

            {/* Close Button (when no URL generated) */}
            {!signedURL && (
              <div className="flex justify-end pt-4">
                <button
                  onClick={onClose}
                  className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white dark:hover:bg-gray-600"
                >
                  Cancel
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}