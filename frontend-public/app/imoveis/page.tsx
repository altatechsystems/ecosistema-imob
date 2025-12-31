'use client';

import * as React from 'react';
import Link from 'next/link';
import { PropertyCard } from '@/components/property/property-card';
import { PropertyFiltersComponent } from '@/components/property/property-filters';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Property, PropertyFilters, PropertyStatus, PropertyVisibility } from '@/types/property';
import { api } from '@/lib/api';
import { Home, Grid, List, Search } from 'lucide-react';

export default function PropertiesPage() {
  const [properties, setProperties] = React.useState<Property[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [viewMode, setViewMode] = React.useState<'grid' | 'list'>('grid');
  const [filters, setFilters] = React.useState<PropertyFilters>({
    // Removed default filters to avoid Firestore composite index requirement
    // status: PropertyStatus.AVAILABLE,
    // visibility: PropertyVisibility.PUBLIC,
  });

  const loadProperties = React.useCallback(async () => {
    try {
      setIsLoading(true);
      const startTime = performance.now();

      const result = await api.getProperties(filters, { limit: 50 });

      const loadTime = performance.now() - startTime;
      console.log(`✅ Loaded ${result.data?.length || 0} properties in ${loadTime.toFixed(0)}ms`);

      // Use properties directly from API
      setProperties(result.data || []);

      const processingTime = performance.now() - startTime;
      console.log(`⚡ Total processing time: ${processingTime.toFixed(0)}ms`);
    } catch (error) {
      console.error('Failed to load properties:', error);
    } finally {
      setIsLoading(false);
    }
  }, [filters]);

  React.useEffect(() => {
    loadProperties();
  }, [loadProperties]);

  const handleClearFilters = () => {
    setFilters({
      // Cleared all filters to show all properties
    });
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b sticky top-0 z-50">
        <div className="container mx-auto px-3 sm:px-4 py-3 sm:py-4">
          <div className="flex items-center justify-between">
            <Link href="/" className="flex items-center gap-2">
              <Home className="w-6 h-6 sm:w-8 sm:h-8 text-blue-600" />
              <span className="text-lg sm:text-2xl font-bold text-gray-900">Imobiliária</span>
            </Link>

            <nav className="hidden md:flex items-center gap-4 lg:gap-6">
              <Link href="/imoveis" className="text-blue-600 font-semibold text-sm lg:text-base">
                Imóveis
              </Link>
              <Link href="/sobre" className="text-gray-700 hover:text-blue-600 font-medium text-sm lg:text-base">
                Sobre
              </Link>
              <Link href="/contato" className="text-gray-700 hover:text-blue-600 font-medium text-sm lg:text-base">
                Contato
              </Link>
            </nav>
          </div>
        </div>
      </header>

      <div className="container mx-auto px-3 sm:px-4 py-6 sm:py-8">
        <div className="flex flex-col lg:flex-row gap-6 sm:gap-8">
          {/* Sidebar Filters */}
          <aside className="w-full lg:w-80 flex-shrink-0">
            <PropertyFiltersComponent
              filters={filters}
              onFiltersChange={setFilters}
              onClearFilters={handleClearFilters}
              variant="sidebar"
            />
          </aside>

          {/* Main Content */}
          <main className="flex-1">
            {/* Header */}
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between mb-4 sm:mb-6 gap-3">
              <div>
                <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 mb-1 sm:mb-2">
                  Imóveis Disponíveis
                </h1>
                <p className="text-sm sm:text-base text-gray-600">
                  {isLoading ? 'Carregando...' : `${properties.length} imóveis encontrados`}
                </p>
              </div>

              {/* View Toggle */}
              <div className="hidden sm:flex items-center gap-2">
                <Button
                  variant={viewMode === 'grid' ? 'primary' : 'outline'}
                  size="sm"
                  onClick={() => setViewMode('grid')}
                >
                  <Grid className="w-4 h-4" />
                </Button>
                <Button
                  variant={viewMode === 'list' ? 'primary' : 'outline'}
                  size="sm"
                  onClick={() => setViewMode('list')}
                >
                  <List className="w-4 h-4" />
                </Button>
              </div>
            </div>

            {/* Properties Grid/List */}
            {isLoading ? (
              <div className={viewMode === 'grid'
                ? 'grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6'
                : 'space-y-6'
              }>
                {[...Array(12)].map((_, i) => (
                  <Card key={i} variant="bordered" padding="none" className="animate-pulse">
                    <div className="w-full h-48 sm:h-56 bg-gray-200 rounded-t-lg" />
                    <div className="p-3 sm:p-4 space-y-3">
                      <div className="h-4 bg-gray-200 rounded w-1/4 mb-2" />
                      <div className="h-7 bg-gray-200 rounded w-2/3 mb-3" />
                      <div className="h-6 bg-gray-200 rounded w-2/3 mb-2" />
                      <div className="h-4 bg-gray-200 rounded w-full mb-3" />
                      <div className="flex gap-3 mb-3">
                        <div className="h-4 bg-gray-200 rounded w-12" />
                        <div className="h-4 bg-gray-200 rounded w-12" />
                        <div className="h-4 bg-gray-200 rounded w-16" />
                      </div>
                      <div className="h-9 bg-gray-200 rounded w-full" />
                    </div>
                  </Card>
                ))}
              </div>
            ) : properties.length > 0 ? (
              <div className={viewMode === 'grid'
                ? 'grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6'
                : 'space-y-6'
              }>
                {properties.map((property) => (
                  <PropertyCard
                    key={property.id}
                    property={property}
                    variant={viewMode}
                  />
                ))}
              </div>
            ) : (
              <Card variant="bordered" padding="lg" className="text-center py-12">
                <Search className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                <h3 className="text-xl font-semibold text-gray-900 mb-2">
                  Nenhum imóvel encontrado
                </h3>
                <p className="text-gray-600 mb-6">
                  Tente ajustar os filtros de busca
                </p>
                <Button variant="primary" size="md" onClick={handleClearFilters}>
                  Limpar Filtros
                </Button>
              </Card>
            )}
          </main>
        </div>
      </div>
    </div>
  );
}
