import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { apiClient } from '../services/api';
import { 
  FolderIcon, 
  ServerIcon, 
  UsersIcon, 
  ChartBarIcon,
  Cog6ToothIcon,
  KeyIcon
} from '@heroicons/react/24/outline';

export default function Dashboard() {
  const { hasRole } = useAuth();
  
  const { data: bucketsData, isLoading: bucketsLoading } = useQuery({
    queryKey: ['buckets'],
    queryFn: () => apiClient.getBuckets(),
  });

  const { data: nodesData, isLoading: nodesLoading } = useQuery({
    queryKey: ['nodes'],
    queryFn: () => apiClient.getNodes(),
    enabled: hasRole('admin'),
  });

  const { data: usersData, isLoading: usersLoading } = useQuery({
    queryKey: ['users'],
    queryFn: () => apiClient.getUsers(),
    enabled: hasRole('admin'),
  });

  const buckets = bucketsData?.buckets || [];
  const nodes = nodesData?.nodes || [];
  const users = usersData?.users || [];

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
      name: 'Storage Nodes',
      value: nodes.length,
      icon: ServerIcon,
      color: 'text-green-400',
      bgColor: 'bg-green-500/10',
      loading: nodesLoading,
      show: hasRole('admin'),
    },
    {
      name: 'Total Users',
      value: users.length,
      icon: UsersIcon,
      color: 'text-purple-400',
      bgColor: 'bg-purple-500/10',
      loading: usersLoading,
      show: hasRole('admin'),
    },
  ];

  return (
    <div>
      {/* Main content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Stats grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          {stats
            .filter(stat => stat.show !== false)
            .map((stat) => (
              <div
                key={stat.name}
                className={`${stat.bgColor} rounded-lg p-6 border border-dark-700`}
              >
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-dark-300">{stat.name}</p>
                    <div className="text-2xl font-bold text-white">
                      {stat.loading ? (
                        <div className="animate-pulse bg-dark-600 h-8 w-16 rounded"></div>
                      ) : (
                        stat.value
                      )}
                    </div>
                  </div>
                  <div className={`${stat.color}`}>
                    <stat.icon className="h-8 w-8" />
                  </div>
                </div>
              </div>
            ))}
        </div>

        {/* Quick actions */}
        <div className="bg-dark-900 rounded-lg border border-dark-700 p-6">
          <h2 className="text-xl font-semibold text-white mb-4">Quick Actions</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <Link to="/buckets" className="flex items-center p-4 bg-dark-800 rounded-lg hover:bg-dark-700 transition-colors text-left">
              <FolderIcon className="h-6 w-6 text-blue-400 mr-3" />
              <div>
                <p className="font-medium text-white">Manage Buckets</p>
                <p className="text-sm text-dark-400">Create and configure storage buckets</p>
              </div>
            </Link>
            
            {(
              <Link to="/nodes" className="flex items-center p-4 bg-dark-800 rounded-lg hover:bg-dark-700 transition-colors text-left">
                <ServerIcon className="h-6 w-6 text-green-400 mr-3" />
                <div>
                  <p className="font-medium text-white">Storage Nodes</p>
                  <p className="text-sm text-dark-400">Manage storage infrastructure</p>
                </div>
              </Link>
            )}

            {hasRole('admin') && (
              <Link to="/users" className="flex items-center p-4 bg-dark-800 rounded-lg hover:bg-dark-700 transition-colors text-left">
                <UsersIcon className="h-6 w-6 text-purple-400 mr-3" />
                <div>
                  <p className="font-medium text-white">User Management</p>
                  <p className="text-sm text-dark-400">Manage users and permissions</p>
                </div>
              </Link>
            )}

            <Link to="/api-keys" className="flex items-center p-4 bg-dark-800 rounded-lg hover:bg-dark-700 transition-colors text-left">
              <KeyIcon className="h-6 w-6 text-blue-400 mr-3" />
              <div>
                <p className="font-medium text-white">API Keys</p>
                <p className="text-sm text-dark-400">Manage programmatic access</p>
              </div>
            </Link>

            <Link to="/analytics" className="flex items-center p-4 bg-dark-800 rounded-lg hover:bg-dark-700 transition-colors text-left">
              <ChartBarIcon className="h-6 w-6 text-yellow-400 mr-3" />
              <div>
                <p className="font-medium text-white">Analytics</p>
                <p className="text-sm text-dark-400">View usage statistics</p>
              </div>
            </Link>

            <Link to="/settings" className="flex items-center p-4 bg-dark-800 rounded-lg hover:bg-dark-700 transition-colors text-left">
              <Cog6ToothIcon className="h-6 w-6 text-gray-400 mr-3" />
              <div>
                <p className="font-medium text-white">Settings</p>
                <p className="text-sm text-dark-400">Configure system settings</p>
              </div>
            </Link>
          </div>
        </div>

        {/* API Documentation Link */}
        <div className="mt-8 text-center">
          <a
            href="http://localhost:8080/swagger/"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center px-6 py-3 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors"
          >
            View API Documentation
          </a>
        </div>
      </div>
    </div>
  );
}