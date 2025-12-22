'use client';

import * as React from 'react';
import Link from 'next/link';
import { PropertyCard } from '@/components/property/property-card';
import { PropertyFiltersComponent } from '@/components/property/property-filters';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Property, PropertyFilters, TransactionType, PropertyStatus, PropertyVisibility } from '@/types/property';
import { api } from '@/lib/api';
import { Search, MapPin, Home, TrendingUp, PhoneCall } from 'lucide-react';

export default function HomePage() {
  const [featuredProperties, setFeaturedProperties] = React.useState<Property[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [filters, setFilters] = React.useState<PropertyFilters>({
    status: PropertyStatus.AVAILABLE,
    visibility: PropertyVisibility.PUBLIC,
  });

  React.useEffect(() => {
    loadFeaturedProperties();
  }, []);

  const loadFeaturedProperties = async () => {
    try {
      setIsLoading(true);
      const properties = await api.getFeaturedProperties(6);
      setFeaturedProperties(properties);
    } catch (error) {
      console.error('Failed to load featured properties:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleClearFilters = () => {
    setFilters({
      status: PropertyStatus.AVAILABLE,
      visibility: PropertyVisibility.PUBLIC,
    });
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b sticky top-0 z-50">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <Link href="/" className="flex items-center gap-2">
              <Home className="w-8 h-8 text-blue-600" />
              <span className="text-2xl font-bold text-gray-900">Imobiliária</span>
            </Link>

            <nav className="hidden md:flex items-center gap-6">
              <Link href="/imoveis" className="text-gray-700 hover:text-blue-600 font-medium">
                Imóveis
              </Link>
              <Link href="/sobre" className="text-gray-700 hover:text-blue-600 font-medium">
                Sobre
              </Link>
              <Link href="/contato" className="text-gray-700 hover:text-blue-600 font-medium">
                Contato
              </Link>
              <Button variant="primary" size="sm">
                Anunciar Imóvel
              </Button>
            </nav>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="bg-gradient-to-br from-blue-600 to-blue-800 text-white">
        <div className="container mx-auto px-4 py-20">
          <div className="max-w-3xl mx-auto text-center mb-12">
            <h1 className="text-4xl md:text-5xl font-bold mb-4">
              Encontre o Imóvel dos Seus Sonhos
            </h1>
            <p className="text-xl text-blue-100">
              Milhares de imóveis para venda e aluguel em todo o Brasil
            </p>
          </div>

          {/* Quick Search */}
          <div className="max-w-5xl mx-auto">
            <PropertyFiltersComponent
              filters={filters}
              onFiltersChange={setFilters}
              onClearFilters={handleClearFilters}
              variant="horizontal"
            />
          </div>
        </div>
      </section>

      {/* Stats Section */}
      <section className="bg-white py-12 border-b">
        <div className="container mx-auto px-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 text-center">
            <div>
              <div className="flex items-center justify-center mb-2">
                <Home className="w-10 h-10 text-blue-600" />
              </div>
              <h3 className="text-3xl font-bold text-gray-900 mb-1">500+</h3>
              <p className="text-gray-600">Imóveis Disponíveis</p>
            </div>
            <div>
              <div className="flex items-center justify-center mb-2">
                <TrendingUp className="w-10 h-10 text-green-600" />
              </div>
              <h3 className="text-3xl font-bold text-gray-900 mb-1">1000+</h3>
              <p className="text-gray-600">Negócios Fechados</p>
            </div>
            <div>
              <div className="flex items-center justify-center mb-2">
                <MapPin className="w-10 h-10 text-orange-600" />
              </div>
              <h3 className="text-3xl font-bold text-gray-900 mb-1">50+</h3>
              <p className="text-gray-600">Cidades Atendidas</p>
            </div>
          </div>
        </div>
      </section>

      {/* Featured Properties */}
      <section className="py-16">
        <div className="container mx-auto px-4">
          <div className="flex items-center justify-between mb-8">
            <div>
              <h2 className="text-3xl font-bold text-gray-900 mb-2">
                Imóveis em Destaque
              </h2>
              <p className="text-gray-600">
                Conheça nossas melhores oportunidades
              </p>
            </div>
            <Link href="/imoveis">
              <Button variant="outline" size="md">
                Ver Todos
              </Button>
            </Link>
          </div>

          {isLoading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {[1, 2, 3, 4, 5, 6].map((i) => (
                <Card key={i} variant="bordered" padding="none" className="animate-pulse">
                  <div className="w-full h-56 bg-gray-200 rounded-t-lg" />
                  <div className="p-4 space-y-3">
                    <div className="h-6 bg-gray-200 rounded w-2/3" />
                    <div className="h-4 bg-gray-200 rounded w-full" />
                    <div className="h-4 bg-gray-200 rounded w-3/4" />
                  </div>
                </Card>
              ))}
            </div>
          ) : featuredProperties.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {featuredProperties.map((property) => (
                <PropertyCard key={property.id} property={property} variant="grid" />
              ))}
            </div>
          ) : (
            <Card variant="bordered" padding="lg" className="text-center py-12">
              <Search className="w-12 h-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-xl font-semibold text-gray-900 mb-2">
                Nenhum imóvel em destaque no momento
              </h3>
              <p className="text-gray-600 mb-6">
                Explore nossa lista completa de imóveis
              </p>
              <Link href="/imoveis">
                <Button variant="primary" size="md">
                  Ver Todos os Imóveis
                </Button>
              </Link>
            </Card>
          )}
        </div>
      </section>

      {/* CTA Section */}
      <section className="bg-blue-600 text-white py-16">
        <div className="container mx-auto px-4">
          <div className="max-w-3xl mx-auto text-center">
            <h2 className="text-3xl md:text-4xl font-bold mb-4">
              Precisa de Ajuda para Encontrar seu Imóvel?
            </h2>
            <p className="text-xl text-blue-100 mb-8">
              Nossa equipe de especialistas está pronta para ajudá-lo
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Link href="/contato">
                <Button variant="secondary" size="lg" leftIcon={<PhoneCall className="w-5 h-5" />}>
                  Fale Conosco
                </Button>
              </Link>
              <Link href="/imoveis">
                <Button variant="outline" size="lg" className="bg-white text-blue-600 hover:bg-gray-100">
                  Buscar Imóveis
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gray-900 text-gray-300 py-12">
        <div className="container mx-auto px-4">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-8 mb-8">
            <div>
              <div className="flex items-center gap-2 mb-4">
                <Home className="w-6 h-6 text-blue-400" />
                <span className="text-xl font-bold text-white">Imobiliária</span>
              </div>
              <p className="text-sm">
                Seu parceiro confiável para encontrar o imóvel perfeito.
              </p>
            </div>

            <div>
              <h3 className="font-semibold text-white mb-4">Links Rápidos</h3>
              <ul className="space-y-2 text-sm">
                <li><Link href="/imoveis" className="hover:text-blue-400">Imóveis</Link></li>
                <li><Link href="/sobre" className="hover:text-blue-400">Sobre Nós</Link></li>
                <li><Link href="/contato" className="hover:text-blue-400">Contato</Link></li>
              </ul>
            </div>

            <div>
              <h3 className="font-semibold text-white mb-4">Categorias</h3>
              <ul className="space-y-2 text-sm">
                <li><Link href="/imoveis?type=apartment" className="hover:text-blue-400">Apartamentos</Link></li>
                <li><Link href="/imoveis?type=house" className="hover:text-blue-400">Casas</Link></li>
                <li><Link href="/imoveis?type=commercial" className="hover:text-blue-400">Comerciais</Link></li>
              </ul>
            </div>

            <div>
              <h3 className="font-semibold text-white mb-4">Contato</h3>
              <ul className="space-y-2 text-sm">
                <li>Email: contato@imobiliaria.com</li>
                <li>Telefone: (11) 3000-0000</li>
                <li>WhatsApp: (11) 99999-9999</li>
              </ul>
            </div>
          </div>

          <div className="border-t border-gray-800 pt-8 text-center text-sm">
            <p>&copy; 2025 Imobiliária. Todos os direitos reservados.</p>
          </div>
        </div>
      </footer>
    </div>
  );
}
