import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { 
  HomeIcon,
  FolderIcon, 
  ServerIcon, 
  UsersIcon, 
  ChartBarIcon,
  Cog6ToothIcon,
  ArrowRightOnRectangleIcon,
  KeyIcon
} from '@heroicons/react/24/outline';

interface LayoutProps {
  children: React.ReactNode;
}

export default function Layout({ children }: LayoutProps) {
  const { user, logout, hasRole } = useAuth();
  const location = useLocation();

  const navigation = [
    { name: 'Dashboard', href: '/', icon: HomeIcon, current: location.pathname === '/' },
    { name: 'Buckets', href: '/buckets', icon: FolderIcon, current: location.pathname === '/buckets' },
    { name: 'Storage Nodes', href: '/nodes', icon: ServerIcon, current: location.pathname === '/nodes', requiresRole: 'manager' },
    { name: 'Users', href: '/users', icon: UsersIcon, current: location.pathname === '/users', requiresRole: 'admin' },
    { name: 'API Keys', href: '/api-keys', icon: KeyIcon, current: location.pathname === '/api-keys' },
    { name: 'Analytics', href: '/analytics', icon: ChartBarIcon, current: location.pathname === '/analytics' },
    { name: 'Settings', href: '/settings', icon: Cog6ToothIcon, current: location.pathname === '/settings' },
  ];

  const handleLogout = () => {
    logout();
  };

  return (
    <div className="min-h-screen bg-dark-950 text-dark-50">
      {/* Sidebar */}
      <div className="fixed inset-y-0 left-0 z-50 w-64 bg-dark-900">
        <div className="flex h-full flex-col">
          {/* Logo */}
          <div className="flex h-16 shrink-0 items-center px-6 border-b border-dark-800">
            <h1 className="text-xl font-bold text-white">SHBucket</h1>
          </div>
          
          {/* Navigation */}
          <nav className="flex flex-1 flex-col p-4">
            <ul className="flex flex-1 flex-col gap-y-2">
              {navigation.map((item) => {
                // Check role requirements
                if (item.requiresRole && !hasRole(item.requiresRole)) {
                  return null;
                }

                return (
                  <li key={item.name}>
                    <Link
                      to={item.href}
                      className={`group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold transition-colors ${
                        item.current
                          ? 'bg-primary-600 text-white'
                          : 'text-dark-300 hover:text-white hover:bg-dark-800'
                      }`}
                    >
                      <item.icon className="h-6 w-6 shrink-0" />
                      {item.name}
                    </Link>
                  </li>
                );
              })}
            </ul>
          </nav>

          {/* User section */}
          <div className="border-t border-dark-800 p-4">
            <div className="flex items-center gap-x-4">
              <div className="flex-1">
                <p className="text-sm font-semibold text-white">{user?.username}</p>
                <p className="text-xs text-dark-400 capitalize">{user?.role}</p>
              </div>
              <button
                onClick={handleLogout}
                className="flex-none p-2 text-dark-400 hover:text-white rounded-md hover:bg-dark-800 transition-colors"
                title="Logout"
              >
                <ArrowRightOnRectangleIcon className="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Main content */}
      <div className="pl-64">
        <main className="p-8">
          {children}
        </main>
      </div>
    </div>
  );
}