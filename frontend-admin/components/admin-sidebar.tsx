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
  MessageSquare,
  Upload,
  BarChart3,
  Settings,
  LogOut,
} from 'lucide-react';
import { cn } from '@/lib/utils';

const menuItems = [
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
    title: 'Corretores',
    href: '/dashboard/corretores',
    icon: UserCog,
  },
  {
    title: 'Importação',
    href: '/dashboard/importacao',
    icon: Upload,
  },
  {
    title: 'Relatórios',
    href: '/dashboard/relatorios',
    icon: BarChart3,
  },
  {
    title: 'Configurações',
    href: '/dashboard/configuracoes',
    icon: Settings,
  },
];

export function AdminSidebar() {
  const pathname = usePathname();

  const handleLogout = async () => {
    try {
      await signOut(auth);
      window.location.href = '/login';
    } catch (error) {
      console.error('Logout error:', error);
    }
  };

  return (
    <aside className="w-64 bg-gray-900 text-white flex flex-col">
      {/* Logo */}
      <div className="p-6 border-b border-gray-800">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-blue-600 rounded-lg flex items-center justify-center">
            <Home className="w-6 h-6" />
          </div>
          <div>
            <h1 className="text-lg font-bold">Admin</h1>
            <p className="text-xs text-gray-400">Imobiliária</p>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4 space-y-1 overflow-y-auto">
        {menuItems.map((item) => {
          // Check if this is the exact match or a sub-route
          // Special case: Dashboard should only be active on exact match
          const isActive = item.href === '/dashboard'
            ? pathname === '/dashboard'
            : pathname === item.href || pathname.startsWith(item.href + '/');
          const Icon = item.icon;

          return (
            <Link
              key={item.href}
              href={item.href}
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
  );
}
