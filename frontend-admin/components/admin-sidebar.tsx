'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { signOut } from 'firebase/auth';
import { auth } from '@/lib/firebase';
import {
  Home,
  Building2,
  Users,
  UserCog,
  Shield,
  MessageSquare,
  Upload,
  BarChart3,
  Settings,
  LogOut,
  X,
  Calendar,
} from 'lucide-react';
import { cn } from '@/lib/utils';

// Business operations - main workflow items
const businessMenuItems = [
  {
    title: 'Dashboard',
    href: '/dashboard',
    icon: Home,
  },
  {
    title: 'Imóveis',
    href: '/dashboard/imoveis',
    icon: Building2,
  },
  {
    title: 'Leads',
    href: '/dashboard/leads',
    icon: MessageSquare,
  },
  {
    title: 'Proprietários',
    href: '/dashboard/proprietarios',
    icon: Users,
  },
  {
    title: 'Confirmações',
    href: '/dashboard/confirmacoes',
    icon: Calendar,
  },
  {
    title: 'Relatórios',
    href: '/dashboard/relatorios',
    icon: BarChart3,
  },
];

// Administrative & configuration items
const adminMenuItems = [
  {
    title: 'Corretores',
    href: '/dashboard/corretores',
    icon: UserCog,
  },
  {
    title: 'Equipe',
    href: '/dashboard/equipe',
    icon: Shield,
  },
  {
    title: 'Importação',
    href: '/dashboard/importacao',
    icon: Upload,
  },
  {
    title: 'Configurações',
    href: '/dashboard/configuracoes',
    icon: Settings,
  },
];

interface AdminSidebarProps {
  isOpen?: boolean;
  onClose?: () => void;
}

export function AdminSidebar({ isOpen = true, onClose }: AdminSidebarProps) {
  const pathname = usePathname();

  const handleLogout = async () => {
    try {
      await signOut(auth);
      window.location.href = '/login';
    } catch (error) {
      console.error('Logout error:', error);
    }
  };

  const handleLinkClick = () => {
    // Close sidebar on mobile when a link is clicked
    if (onClose) {
      onClose();
    }
  };

  return (
    <>
      {/* Backdrop for mobile and tablet */}
      {isOpen && onClose && (
        <div
          className="fixed inset-0 bg-black bg-opacity-50 z-40 xl:hidden"
          onClick={onClose}
        />
      )}

      {/* Sidebar */}
      <aside
        className={cn(
          "fixed xl:static inset-y-0 left-0 z-50 w-64 bg-gray-900 text-white flex flex-col transform transition-transform duration-300 ease-in-out",
          isOpen ? "translate-x-0" : "-translate-x-full xl:translate-x-0"
        )}
      >
      {/* Logo */}
      <div className="p-6 border-b border-gray-800">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-blue-600 rounded-lg flex items-center justify-center">
              <Home className="w-6 h-6" />
            </div>
            <div>
              <h1 className="text-lg font-bold">Admin</h1>
              <p className="text-xs text-gray-400">Imobiliária</p>
            </div>
          </div>
          {/* Close button for mobile and tablet */}
          {onClose && (
            <button
              onClick={onClose}
              className="xl:hidden p-2 hover:bg-gray-800 rounded-lg transition-colors"
              aria-label="Fechar menu"
            >
              <X className="w-5 h-5" />
            </button>
          )}
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4 overflow-y-auto">
        {/* Business Section */}
        <div className="space-y-1 mb-6">
          {businessMenuItems.map((item) => {
            const isActive = item.href === '/dashboard'
              ? pathname === '/dashboard'
              : pathname === item.href || pathname.startsWith(item.href + '/');
            const Icon = item.icon;

            return (
              <Link
                key={item.href}
                href={item.href}
                onClick={handleLinkClick}
                className={cn(
                  'flex items-center gap-3 px-4 py-3 rounded-lg transition-colors',
                  isActive
                    ? 'bg-blue-600 text-white'
                    : 'text-gray-300 hover:bg-gray-800 hover:text-white'
                )}
              >
                <Icon className="w-5 h-5" />
                <span className="font-medium">{item.title}</span>
              </Link>
            );
          })}
        </div>

        {/* Divider */}
        <div className="border-t border-gray-800 my-4"></div>

        {/* Administration Section */}
        <div className="space-y-1">
          <div className="px-4 py-2 text-xs font-semibold text-gray-500 uppercase tracking-wider">
            Administração
          </div>
          {adminMenuItems.map((item) => {
            const isActive = pathname === item.href || pathname.startsWith(item.href + '/');
            const Icon = item.icon;

            return (
              <Link
                key={item.href}
                href={item.href}
                onClick={handleLinkClick}
                className={cn(
                  'flex items-center gap-3 px-4 py-3 rounded-lg transition-colors',
                  isActive
                    ? 'bg-blue-600 text-white'
                    : 'text-gray-300 hover:bg-gray-800 hover:text-white'
                )}
              >
                <Icon className="w-5 h-5" />
                <span className="font-medium">{item.title}</span>
              </Link>
            );
          })}
        </div>
      </nav>

      {/* User / Logout */}
      <div className="p-4 border-t border-gray-800">
        <button
          onClick={handleLogout}
          className="w-full flex items-center gap-3 px-4 py-3 rounded-lg text-gray-300 hover:bg-gray-800 hover:text-white transition-colors"
        >
          <LogOut className="w-5 h-5" />
          <span className="font-medium">Sair</span>
        </button>
      </div>
    </aside>
    </>
  );
}
