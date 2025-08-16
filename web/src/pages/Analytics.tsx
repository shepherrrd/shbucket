import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { apiClient } from '../services/api';
import {
  ArrowLeftIcon,
  FolderIcon,
  ServerIcon,
  CloudArrowUpIcon,
  CloudArrowDownIcon,
  UsersIcon,
  DocumentIcon,
} from '@heroicons/react/24/outline';

export default function Analytics() {
  const { data: bucketsData, isLoading: bucketsLoading } = useQuery({
    queryKey: ['buckets'],
    queryFn: () => apiClient.getBuckets(),
  });

  const { data: usersData, isLoading: usersLoading } = useQuery({
    queryKey: ['users'],
    queryFn: () => apiClient.getUsers(),
  });

  const { data: nodesData, isLoading: nodesLoading } = useQuery({
    queryKey: ['nodes'],
    queryFn: () => apiClient.getNodes(),
  });

  const buckets = bucketsData?.buckets || [];
  const users = usersData?.users || [];
  const nodes = nodesData?.nodes || [];

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getTotalFiles = () => {
    return buckets.reduce((total, bucket) => total + (bucket.stats?.total_files || 0), 0);
  };

  const getTotalSize = () => {
    return buckets.reduce((total, bucket) => total + (bucket.stats?.total_size || 0), 0);
  };

  const getNodeCapacityUsed = () => {
    const totalCapacity = nodes.reduce((total, node) => total + node.max_storage, 0);
    const totalUsed = nodes.reduce((total, node) => total + node.used_storage, 0);
    return { totalCapacity, totalUsed, percentage: totalCapacity > 0 ? (totalUsed / totalCapacity) * 100 : 0 };
  };

  const getActiveUsers = () => {
    return users.filter(user => user.is_active).length;
  };

  const getHealthyNodes = () => {
    return nodes.filter(node => node.is_healthy).length;
  };

  const capacityStats = getNodeCapacityUsed();

  const stats = [
    {
      name: 'Total Buckets',
      value: buckets.length,
      icon: FolderIcon,
      color: 'text-blue-400',
      bgColor: 'bg-blue-500/10',
      loading: bucketsLoading,
    },
    {
      name: 'Total Files',
      value: getTotalFiles(),
      icon: DocumentIcon,
      color: 'text-green-400',
      bgColor: 'bg-green-500/10',
      loading: bucketsLoading,
    },
    {
      name: 'Total Size',
      value: formatBytes(getTotalSize()),
      icon: CloudArrowUpIcon,
      color: 'text-purple-400',
      bgColor: 'bg-purple-500/10',
      loading: bucketsLoading,
      isSize: true,
    },
    {
      name: 'Active Users',
      value: getActiveUsers(),
      icon: UsersIcon,
      color: 'text-yellow-400',
      bgColor: 'bg-yellow-500/10',
      loading: usersLoading,
    },
    {
      name: 'Storage Nodes',
      value: `${getHealthyNodes()}/${nodes.length}`,
      icon: ServerIcon,
      color: 'text-indigo-400',
      bgColor: 'bg-indigo-500/10',
      loading: nodesLoading,
      subtitle: 'Healthy/Total',
    },
    {
      name: 'Capacity Used',
      value: `${capacityStats.percentage.toFixed(1)}%`,
      icon: CloudArrowDownIcon,
      color: 'text-red-400',
      bgColor: 'bg-red-500/10',
      loading: nodesLoading,
      subtitle: `${formatBytes(capacityStats.totalUsed)} / ${formatBytes(capacityStats.totalCapacity)}`,
    },
  ];

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
                <h1 className="text-3xl font-bold text-white">Analytics</h1>
                <p className="text-dark-400">View usage statistics and system metrics</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Main content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Stats grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
          {stats.map((stat) => (
            <div
              key={stat.name}
              className={`${stat.bgColor} rounded-lg p-6 border border-dark-700`}
            >
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <p className="text-sm font-medium text-dark-300">{stat.name}</p>
                  <p className="text-2xl font-bold text-white mt-1">
                    {stat.loading ? (
                      <div className="animate-pulse bg-dark-600 h-8 w-16 rounded"></div>
                    ) : (
                      stat.value
                    )}
                  </p>
                  {stat.subtitle && (
                    <p className="text-xs text-dark-400 mt-1">{stat.subtitle}</p>
                  )}
                </div>
                <div className={`${stat.color} ml-4`}>
                  <stat.icon className="h-8 w-8" />
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Detailed sections */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Bucket Overview */}
          <div className="bg-dark-900 rounded-lg border border-dark-700 p-6">
            <h3 className="text-lg font-semibold text-white mb-4">Bucket Overview</h3>
            {bucketsLoading ? (
              <div className="animate-pulse">
                {[...Array(3)].map((_, i) => (
                  <div key={i} className="flex justify-between items-center mb-3">
                    <div className="bg-dark-600 h-4 w-24 rounded"></div>
                    <div className="bg-dark-600 h-4 w-16 rounded"></div>
                  </div>
                ))}
              </div>
            ) : buckets.length === 0 ? (
              <p className="text-dark-400 text-sm">No buckets available</p>
            ) : (
              <div className="space-y-3">
                {buckets.slice(0, 5).map((bucket) => (
                  <div key={bucket.id} className="flex justify-between items-center">
                    <div>
                      <p className="text-white font-medium">{bucket.name}</p>
                      <p className="text-dark-400 text-xs">{bucket.description || 'No description'}</p>
                    </div>
                    <div className="text-right">
                      <p className="text-white text-sm">{bucket.stats?.total_files || 0} files</p>
                      <p className="text-dark-400 text-xs">{formatBytes(bucket.stats?.total_size || 0)}</p>
                    </div>
                  </div>
                ))}
                {buckets.length > 5 && (
                  <p className="text-dark-400 text-sm text-center pt-2">
                    And {buckets.length - 5} more buckets...
                  </p>
                )}
              </div>
            )}
          </div>

          {/* Storage Nodes Status */}
          <div className="bg-dark-900 rounded-lg border border-dark-700 p-6">
            <h3 className="text-lg font-semibold text-white mb-4">Storage Nodes Status</h3>
            {nodesLoading ? (
              <div className="animate-pulse">
                {[...Array(3)].map((_, i) => (
                  <div key={i} className="flex justify-between items-center mb-3">
                    <div className="bg-dark-600 h-4 w-24 rounded"></div>
                    <div className="bg-dark-600 h-4 w-16 rounded"></div>
                  </div>
                ))}
              </div>
            ) : nodes.length === 0 ? (
              <p className="text-dark-400 text-sm">No storage nodes configured</p>
            ) : (
              <div className="space-y-3">
                {nodes.map((node) => (
                  <div key={node.id} className="flex justify-between items-center">
                    <div className="flex items-center space-x-2">
                      <div className={`w-2 h-2 rounded-full ${
                        node.is_healthy ? 'bg-green-400' : 'bg-red-400'
                      }`}></div>
                      <div>
                        <p className="text-white font-medium">{node.name}</p>
                        <p className="text-dark-400 text-xs">{node.url}</p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="text-white text-sm">{node.is_healthy ? 'Healthy' : 'Unhealthy'}</p>
                      <p className="text-dark-400 text-xs">
                        {node.max_storage > 0 ? ((node.used_storage / node.max_storage) * 100).toFixed(1) : 0}% used
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* User Activity */}
          <div className="bg-dark-900 rounded-lg border border-dark-700 p-6">
            <h3 className="text-lg font-semibold text-white mb-4">User Activity</h3>
            {usersLoading ? (
              <div className="animate-pulse">
                {[...Array(3)].map((_, i) => (
                  <div key={i} className="flex justify-between items-center mb-3">
                    <div className="bg-dark-600 h-4 w-24 rounded"></div>
                    <div className="bg-dark-600 h-4 w-16 rounded"></div>
                  </div>
                ))}
              </div>
            ) : users.length === 0 ? (
              <p className="text-dark-400 text-sm">No users available</p>
            ) : (
              <div className="space-y-3">
                {users.slice(0, 5).map((user) => (
                  <div key={user.id} className="flex justify-between items-center">
                    <div>
                      <p className="text-white font-medium">{user.username}</p>
                      <p className="text-dark-400 text-xs">{user.email}</p>
                    </div>
                    <div className="text-right">
                      <p className={`text-sm capitalize ${user.is_active ? 'text-green-400' : 'text-red-400'}`}>
                        {user.is_active ? 'Active' : 'Inactive'}
                      </p>
                      <p className="text-dark-400 text-xs">
                        {user.last_login ? new Date(user.last_login).toLocaleDateString() : 'Never'}
                      </p>
                    </div>
                  </div>
                ))}
                {users.length > 5 && (
                  <p className="text-dark-400 text-sm text-center pt-2">
                    And {users.length - 5} more users...
                  </p>
                )}
              </div>
            )}
          </div>

          {/* System Health */}
          <div className="bg-dark-900 rounded-lg border border-dark-700 p-6">
            <h3 className="text-lg font-semibold text-white mb-4">System Health</h3>
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <span className="text-dark-300">Overall Status</span>
                <span className="text-green-400 font-medium">Operational</span>
              </div>
              
              <div className="flex justify-between items-center">
                <span className="text-dark-300">Storage Capacity</span>
                <div className="flex items-center space-x-2">
                  <div className="w-24 bg-dark-700 rounded-full h-2">
                    <div
                      className={`h-2 rounded-full ${
                        capacityStats.percentage < 70 ? 'bg-green-400' :
                        capacityStats.percentage < 90 ? 'bg-yellow-400' : 'bg-red-400'
                      }`}
                      style={{ width: `${Math.min(capacityStats.percentage, 100)}%` }}
                    ></div>
                  </div>
                  <span className="text-white text-sm">{capacityStats.percentage.toFixed(1)}%</span>
                </div>
              </div>

              <div className="flex justify-between items-center">
                <span className="text-dark-300">Healthy Nodes</span>
                <span className="text-white">{getHealthyNodes()}/{nodes.length}</span>
              </div>

              <div className="flex justify-between items-center">
                <span className="text-dark-300">Active Users</span>
                <span className="text-white">{getActiveUsers()}/{users.length}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}