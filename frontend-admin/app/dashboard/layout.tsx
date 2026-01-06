'use client';

import { useState } from 'react';
import { AuthGuard } from '@/components/auth-guard';
import { AdminSidebar } from '@/components/admin-sidebar';
import { AdminHeader } from '@/components/admin-header';

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);

  const toggleSidebar = () => {
    setIsSidebarOpen(!isSidebarOpen);
  };

  const closeSidebar = () => {
    setIsSidebarOpen(false);
  };

  return (
    <AuthGuard>
      <div className="flex h-screen bg-gray-50">
        <AdminSidebar isOpen={isSidebarOpen} onClose={closeSidebar} />
        <div className="flex-1 flex flex-col overflow-hidden">
          <AdminHeader onMenuClick={toggleSidebar} />
          <main className="flex-1 overflow-y-auto p-3 sm:p-4 md:p-6">
            {children}
          </main>
        </div>
      </div>
    </AuthGuard>
  );
}
