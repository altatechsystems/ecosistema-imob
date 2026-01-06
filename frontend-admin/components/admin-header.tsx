'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/hooks/use-auth';
import { TenantSelector } from '@/components/tenant-selector';
import { Bell, Search, Menu } from 'lucide-react';

interface AdminHeaderProps {
  onMenuClick?: () => void;
}

export function AdminHeader({ onMenuClick }: AdminHeaderProps) {
  const { user } = useAuth();
  const [isPlatformAdmin, setIsPlatformAdmin] = useState(false);

  useEffect(() => {
    // Check if user is platform admin from localStorage
    // This will be set during login
    const platformAdminFlag = localStorage.getItem('is_platform_admin');
    setIsPlatformAdmin(platformAdminFlag === 'true');
  }, []);

  return (
    <header className="bg-white border-b border-gray-200 px-4 md:px-6 py-4">
      <div className="flex items-center justify-between">
        {/* Hamburger menu for mobile and tablet */}
        {onMenuClick && (
          <button
            onClick={onMenuClick}
            className="xl:hidden p-2 mr-3 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors"
            aria-label="Abrir menu"
          >
            <Menu className="w-6 h-6" />
          </button>
        )}

        {/* Search */}
        <div className="flex-1 max-w-2xl">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
            <input
              type="text"
              placeholder="Buscar imóveis, leads, proprietários..."
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>

        {/* Right side */}
        <div className="flex items-center gap-4 ml-6">
          {/* Tenant Selector (only for platform admins) */}
          <TenantSelector isPlatformAdmin={isPlatformAdmin} />

          {/* Notifications */}
          <button className="relative p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors">
            <Bell className="w-5 h-5" />
            <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full"></span>
          </button>

          {/* User menu */}
          <div className="flex items-center gap-3 pl-4 border-l border-gray-200">
            <div className="w-10 h-10 bg-blue-600 rounded-full flex items-center justify-center">
              <span className="text-white font-semibold">
                {user?.email?.charAt(0).toUpperCase() || 'U'}
              </span>
            </div>
            <div className="hidden md:block">
              <p className="text-sm font-medium text-gray-900">
                {user?.displayName || 'Usuário'}
              </p>
              <p className="text-xs text-gray-500">
                {user?.email || 'usuario@example.com'}
              </p>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}
