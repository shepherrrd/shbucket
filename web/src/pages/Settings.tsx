import { useState } from 'react';
import { Link } from 'react-router-dom';
import toast from 'react-hot-toast';
import {
  Cog6ToothIcon,
  ArrowLeftIcon,
  ShieldCheckIcon,
  ServerIcon,
  GlobeAltIcon,
} from '@heroicons/react/24/outline';

export default function Settings() {
  const [activeTab, setActiveTab] = useState<'general' | 'security' | 'storage' | 'api'>('general');
  const [settings, setSettings] = useState({
    general: {
      appName: 'SHBucket',
      appDescription: 'Self-hosted S3 storage solution',
      timezone: 'UTC',
      defaultFileExpiry: '7',
      maxUploadSize: '100',
    },
    security: {
      sessionTimeout: '24',
      maxLoginAttempts: '5',
      passwordMinLength: '8',
      requireStrongPasswords: true,
      enableTwoFactor: false,
    },
    storage: {
      defaultBucketQuota: '1000',
      compressionEnabled: true,
      encryptionEnabled: false,
      versioningEnabled: false,
      backupEnabled: true,
      backupRetention: '30',
    },
    api: {
      rateLimitEnabled: true,
      requestsPerMinute: '100',
      enableApiKeys: true,
      enableCors: true,
      corsOrigins: 'http://localhost:3000',
      swaggerEnabled: true,
    },
  });

  const handleSettingChange = (category: keyof typeof settings, key: string, value: string | boolean) => {
    setSettings(prev => ({
      ...prev,
      [category]: {
        ...prev[category],
        [key]: value,
      },
    }));
  };

  const handleSave = () => {
    // In a real app, this would send the settings to the backend
    toast.success('Settings saved successfully!');
  };

  const handleReset = () => {
    if (confirm('Are you sure you want to reset all settings to default values?')) {
      // Reset to default values
      toast.success('Settings reset to defaults');
    }
  };

  const tabs = [
    { id: 'general' as const, name: 'General', icon: Cog6ToothIcon },
    { id: 'security' as const, name: 'Security', icon: ShieldCheckIcon },
    { id: 'storage' as const, name: 'Storage', icon: ServerIcon },
    { id: 'api' as const, name: 'API', icon: GlobeAltIcon },
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
                <h1 className="text-3xl font-bold text-white">Settings</h1>
                <p className="text-dark-400">Configure system settings</p>
              </div>
            </div>
            <div className="flex space-x-3">
              <button
                onClick={handleReset}
                className="px-4 py-2 bg-dark-700 hover:bg-dark-600 text-white font-medium rounded-lg transition-colors"
              >
                Reset to Defaults
              </button>
              <button
                onClick={handleSave}
                className="px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white font-medium rounded-lg transition-colors"
              >
                Save Changes
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Main content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex">
          {/* Sidebar */}
          <div className="w-64 mr-8">
            <nav className="space-y-2">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`w-full flex items-center px-4 py-2 text-left rounded-lg transition-colors ${
                    activeTab === tab.id
                      ? 'bg-primary-600 text-white'
                      : 'bg-dark-800 text-dark-300 hover:bg-dark-700 hover:text-white'
                  }`}
                >
                  <tab.icon className="h-5 w-5 mr-3" />
                  {tab.name}
                </button>
              ))}
            </nav>
          </div>

          {/* Content */}
          <div className="flex-1">
            <div className="bg-dark-900 rounded-lg border border-dark-700 p-6">
              {/* General Settings */}
              {activeTab === 'general' && (
                <div>
                  <h3 className="text-lg font-semibold text-white mb-6">General Settings</h3>
                  <div className="space-y-6">
                    <div>
                      <label className="block text-sm font-medium text-dark-300 mb-2">
                        Application Name
                      </label>
                      <input
                        type="text"
                        value={settings.general.appName}
                        onChange={(e) => handleSettingChange('general', 'appName', e.target.value)}
                        className="w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-dark-300 mb-2">
                        Description
                      </label>
                      <textarea
                        value={settings.general.appDescription}
                        onChange={(e) => handleSettingChange('general', 'appDescription', e.target.value)}
                        rows={3}
                        className="w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-dark-300 mb-2">
                        Timezone
                      </label>
                      <select
                        value={settings.general.timezone}
                        onChange={(e) => handleSettingChange('general', 'timezone', e.target.value)}
                        className="w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-primary-500"
                      >
                        <option value="UTC">UTC</option>
                        <option value="America/New_York">Eastern Time</option>
                        <option value="America/Chicago">Central Time</option>
                        <option value="America/Denver">Mountain Time</option>
                        <option value="America/Los_Angeles">Pacific Time</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-dark-300 mb-2">
                        Default File Expiry (days)
                      </label>
                      <input
                        type="number"
                        value={settings.general.defaultFileExpiry}
                        onChange={(e) => handleSettingChange('general', 'defaultFileExpiry', e.target.value)}
                        min="1"
                        className="w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-dark-300 mb-2">
                        Max Upload Size (MB)
                      </label>
                      <input
                        type="number"
                        value={settings.general.maxUploadSize}
                        onChange={(e) => handleSettingChange('general', 'maxUploadSize', e.target.value)}
                        min="1"
                        className="w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
                      />
                    </div>
                  </div>
                </div>
              )}

              {/* Security Settings */}
              {activeTab === 'security' && (
                <div>
                  <h3 className="text-lg font-semibold text-white mb-6">Security Settings</h3>
                  <div className="space-y-6">
                    <div>
                      <label className="block text-sm font-medium text-dark-300 mb-2">
                        Session Timeout (hours)
                      </label>
                      <input
                        type="number"
                        value={settings.security.sessionTimeout}
                        onChange={(e) => handleSettingChange('security', 'sessionTimeout', e.target.value)}
                        min="1"
                        className="w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-dark-300 mb-2">
                        Max Login Attempts
                      </label>
                      <input
                        type="number"
                        value={settings.security.maxLoginAttempts}
                        onChange={(e) => handleSettingChange('security', 'maxLoginAttempts', e.target.value)}
                        min="1"
                        max="10"
                        className="w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-dark-300 mb-2">
                        Minimum Password Length
                      </label>
                      <input
                        type="number"
                        value={settings.security.passwordMinLength}
                        onChange={(e) => handleSettingChange('security', 'passwordMinLength', e.target.value)}
                        min="4"
                        max="32"
                        className="w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
                      />
                    </div>
                    <div className="flex items-center">
                      <input
                        type="checkbox"
                        checked={settings.security.requireStrongPasswords}
                        onChange={(e) => handleSettingChange('security', 'requireStrongPasswords', e.target.checked)}
                        className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-dark-600 rounded bg-dark-800"
                      />
                      <label className="ml-2 text-sm text-dark-300">
                        Require strong passwords (uppercase, lowercase, numbers, symbols)
                      </label>
                    </div>
                    <div className="flex items-center">
                      <input
                        type="checkbox"
                        checked={settings.security.enableTwoFactor}
                        onChange={(e) => handleSettingChange('security', 'enableTwoFactor', e.target.checked)}
                        className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-dark-600 rounded bg-dark-800"
                      />
                      <label className="ml-2 text-sm text-dark-300">
                        Enable two-factor authentication
                      </label>
                    </div>
                  </div>
                </div>
              )}

              {/* Storage Settings */}
              {activeTab === 'storage' && (
                <div>
                  <h3 className="text-lg font-semibold text-white mb-6">Storage Settings</h3>
                  <div className="space-y-6">
                    <div>
                      <label className="block text-sm font-medium text-dark-300 mb-2">
                        Default Bucket Quota (MB)
                      </label>
                      <input
                        type="number"
                        value={settings.storage.defaultBucketQuota}
                        onChange={(e) => handleSettingChange('storage', 'defaultBucketQuota', e.target.value)}
                        min="1"
                        className="w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
                      />
                    </div>
                    <div className="flex items-center">
                      <input
                        type="checkbox"
                        checked={settings.storage.compressionEnabled}
                        onChange={(e) => handleSettingChange('storage', 'compressionEnabled', e.target.checked)}
                        className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-dark-600 rounded bg-dark-800"
                      />
                      <label className="ml-2 text-sm text-dark-300">
                        Enable file compression
                      </label>
                    </div>
                    <div className="flex items-center">
                      <input
                        type="checkbox"
                        checked={settings.storage.encryptionEnabled}
                        onChange={(e) => handleSettingChange('storage', 'encryptionEnabled', e.target.checked)}
                        className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-dark-600 rounded bg-dark-800"
                      />
                      <label className="ml-2 text-sm text-dark-300">
                        Enable encryption at rest
                      </label>
                    </div>
                    <div className="flex items-center">
                      <input
                        type="checkbox"
                        checked={settings.storage.versioningEnabled}
                        onChange={(e) => handleSettingChange('storage', 'versioningEnabled', e.target.checked)}
                        className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-dark-600 rounded bg-dark-800"
                      />
                      <label className="ml-2 text-sm text-dark-300">
                        Enable file versioning by default
                      </label>
                    </div>
                    <div className="flex items-center">
                      <input
                        type="checkbox"
                        checked={settings.storage.backupEnabled}
                        onChange={(e) => handleSettingChange('storage', 'backupEnabled', e.target.checked)}
                        className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-dark-600 rounded bg-dark-800"
                      />
                      <label className="ml-2 text-sm text-dark-300">
                        Enable automatic backups
                      </label>
                    </div>
                    {settings.storage.backupEnabled && (
                      <div>
                        <label className="block text-sm font-medium text-dark-300 mb-2">
                          Backup Retention (days)
                        </label>
                        <input
                          type="number"
                          value={settings.storage.backupRetention}
                          onChange={(e) => handleSettingChange('storage', 'backupRetention', e.target.value)}
                          min="1"
                          className="w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
                        />
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* API Settings */}
              {activeTab === 'api' && (
                <div>
                  <h3 className="text-lg font-semibold text-white mb-6">API Settings</h3>
                  <div className="space-y-6">
                    <div className="flex items-center">
                      <input
                        type="checkbox"
                        checked={settings.api.rateLimitEnabled}
                        onChange={(e) => handleSettingChange('api', 'rateLimitEnabled', e.target.checked)}
                        className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-dark-600 rounded bg-dark-800"
                      />
                      <label className="ml-2 text-sm text-dark-300">
                        Enable rate limiting
                      </label>
                    </div>
                    {settings.api.rateLimitEnabled && (
                      <div>
                        <label className="block text-sm font-medium text-dark-300 mb-2">
                          Requests per minute per IP
                        </label>
                        <input
                          type="number"
                          value={settings.api.requestsPerMinute}
                          onChange={(e) => handleSettingChange('api', 'requestsPerMinute', e.target.value)}
                          min="1"
                          className="w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
                        />
                      </div>
                    )}
                    <div className="flex items-center">
                      <input
                        type="checkbox"
                        checked={settings.api.enableApiKeys}
                        onChange={(e) => handleSettingChange('api', 'enableApiKeys', e.target.checked)}
                        className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-dark-600 rounded bg-dark-800"
                      />
                      <label className="ml-2 text-sm text-dark-300">
                        Enable API key authentication
                      </label>
                    </div>
                    <div className="flex items-center">
                      <input
                        type="checkbox"
                        checked={settings.api.enableCors}
                        onChange={(e) => handleSettingChange('api', 'enableCors', e.target.checked)}
                        className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-dark-600 rounded bg-dark-800"
                      />
                      <label className="ml-2 text-sm text-dark-300">
                        Enable CORS
                      </label>
                    </div>
                    {settings.api.enableCors && (
                      <div>
                        <label className="block text-sm font-medium text-dark-300 mb-2">
                          CORS Origins (comma-separated)
                        </label>
                        <input
                          type="text"
                          value={settings.api.corsOrigins}
                          onChange={(e) => handleSettingChange('api', 'corsOrigins', e.target.value)}
                          placeholder="http://localhost:3000, https://myapp.com"
                          className="w-full px-3 py-2 bg-dark-800 border border-dark-600 rounded-md text-white placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
                        />
                      </div>
                    )}
                    <div className="flex items-center">
                      <input
                        type="checkbox"
                        checked={settings.api.swaggerEnabled}
                        onChange={(e) => handleSettingChange('api', 'swaggerEnabled', e.target.checked)}
                        className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-dark-600 rounded bg-dark-800"
                      />
                      <label className="ml-2 text-sm text-dark-300">
                        Enable Swagger UI documentation
                      </label>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}