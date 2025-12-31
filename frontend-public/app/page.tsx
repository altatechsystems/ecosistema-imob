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
import { OrganizationStructuredData } from '@/components/seo/organization-structured-data';
import { MobileMenu } from '@/components/navigation/mobile-menu';

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
    <>
      {/* SEO Structured Data */}
      <OrganizationStructuredData />

      <div className="min-h-screen bg-gray-50">
        {/* Header */}
        <header className="bg-white border-b sticky top-0 z-50">
        <div className="container mx-auto px-3 sm:px-4 py-3 sm:py-4">
          <div className="flex items-center justify-between">
            <Link href="/" className="flex items-center gap-2">
              <Home className="w-6 h-6 sm:w-8 sm:h-8 text-blue-600" />
              <span className="text-lg sm:text-2xl font-bold text-gray-900">Imobiliária</span>
            </Link>

            <div className="flex items-center gap-2">
              <nav className="hidden md:flex items-center gap-4 lg:gap-6">
                <Link href="/imoveis" className="text-gray-700 hover:text-blue-600 font-medium text-sm lg:text-base">
                  Imóveis
                </Link>
                <Link href="/sobre" className="text-gray-700 hover:text-blue-600 font-medium text-sm lg:text-base">
                  Sobre
                </Link>
                <Link href="/contato" className="text-gray-700 hover:text-blue-600 font-medium text-sm lg:text-base">
                  Contato
                </Link>
                <Link href="/cadastro-imobiliaria" className="text-gray-700 hover:text-blue-600 font-medium text-sm lg:text-base hidden lg:inline">
                  Para Imobiliárias
                </Link>
                <Link href="http://localhost:3002/login" target="_blank">
                  <Button variant="outline" size="sm">
                    Login
                  </Button>
                </Link>
              </nav>

              {/* Mobile Menu */}
              <MobileMenu currentPath="/" />
            </div>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="bg-gradient-to-br from-blue-600 to-blue-800 text-white">
        <div className="container mx-auto px-3 sm:px-4 py-12 sm:py-16 md:py-20">
          <div className="max-w-3xl mx-auto text-center mb-8 sm:mb-12">
            <h1 className="text-2xl sm:text-3xl md:text-4xl lg:text-5xl font-bold mb-3 sm:mb-4">
              Encontre o Imóvel dos Seus Sonhos
            </h1>
            <p className="text-base sm:text-lg md:text-xl text-blue-100">
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
      <section className="bg-white py-8 sm:py-12 border-b">
        <div className="container mx-auto px-3 sm:px-4">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 sm:gap-8 text-center">
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
      <section className="py-8 sm:py-12 md:py-16">
        <div className="container mx-auto px-3 sm:px-4">
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between mb-6 sm:mb-8 gap-4">
            <div>
              <h2 className="text-2xl sm:text-3xl font-bold text-gray-900 mb-2">
                Imóveis em Destaque
              </h2>
              <p className="text-sm sm:text-base text-gray-600">
                Conheça nossas melhores oportunidades
              </p>
            </div>
            <Link href="/imoveis">
              <Button variant="outline" size="md" className="whitespace-nowrap">
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
      <section className="bg-blue-600 text-white py-12 sm:py-16">
        <div className="container mx-auto px-3 sm:px-4">
          <div className="max-w-3xl mx-auto text-center">
            <h2 className="text-2xl sm:text-3xl md:text-4xl font-bold mb-3 sm:mb-4">
              Precisa de Ajuda para Encontrar seu Imóvel?
            </h2>
            <p className="text-base sm:text-lg md:text-xl text-blue-100 mb-6 sm:mb-8">
              Nossa equipe de especialistas está pronta para ajudá-lo
            </p>
            <div className="flex flex-col sm:flex-row gap-3 sm:gap-4 justify-center">
              <Link href="/contato">
                <Button variant="primary" size="lg" leftIcon={<PhoneCall className="w-5 h-5" />}>
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

      {/* CTA for Real Estate Agencies */}
      <section className="bg-gradient-to-br from-gray-900 to-gray-800 text-white py-12 sm:py-16 md:py-20">
        <div className="container mx-auto px-3 sm:px-4">
          <div className="max-w-4xl mx-auto">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8 sm:gap-12 items-center">
              <div>
                <h2 className="text-2xl sm:text-3xl md:text-4xl font-bold mb-3 sm:mb-4">
                  Você é uma Imobiliária?
                </h2>
                <p className="text-base sm:text-lg md:text-xl text-gray-300 mb-4 sm:mb-6">
                  Junte-se à nossa plataforma e alcance milhares de potenciais compradores
                </p>
                <ul className="space-y-3 mb-8">
                  <li className="flex items-start gap-3">
                    <div className="w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
                      <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    </div>
                    <span className="text-gray-200">Gerenciamento completo de imóveis</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <div className="w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
                      <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    </div>
                    <span className="text-gray-200">Sistema de leads automatizado</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <div className="w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
                      <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    </div>
                    <span className="text-gray-200">Importação em massa de anúncios</span>
                  </li>
                  <li className="flex items-start gap-3">
                    <div className="w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
                      <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    </div>
                    <span className="text-gray-200">Co-corretagem facilitada entre parceiros</span>
                  </li>
                </ul>
                <div className="flex flex-col sm:flex-row gap-4">
                  <Link href="/cadastro-imobiliaria">
                    <Button variant="primary" size="lg" className="bg-blue-600 hover:bg-blue-700">
                      Cadastre-se
                    </Button>
                  </Link>
                  <Link href="http://localhost:3002/login" target="_blank">
                    <Button variant="outline" size="lg" className="border-gray-400 text-white hover:bg-gray-700">
                      Já sou cadastrado
                    </Button>
                  </Link>
                </div>
              </div>

              <div className="hidden md:block">
                <div className="bg-gradient-to-br from-blue-600 to-blue-800 rounded-2xl p-8 shadow-2xl">
                  <h3 className="text-2xl font-bold mb-6">Comece Grátis</h3>
                  <div className="space-y-4">
                    <div className="flex items-start gap-3">
                      <div className="w-8 h-8 bg-white/20 rounded-lg flex items-center justify-center flex-shrink-0">
                        <span className="text-white font-bold">1</span>
                      </div>
                      <div>
                        <p className="font-medium">Crie sua conta</p>
                        <p className="text-sm text-blue-100">Em menos de 2 minutos</p>
                      </div>
                    </div>
                    <div className="flex items-start gap-3">
                      <div className="w-8 h-8 bg-white/20 rounded-lg flex items-center justify-center flex-shrink-0">
                        <span className="text-white font-bold">2</span>
                      </div>
                      <div>
                        <p className="font-medium">Cadastre seus imóveis</p>
                        <p className="text-sm text-blue-100">Manual ou importação em massa</p>
                      </div>
                    </div>
                    <div className="flex items-start gap-3">
                      <div className="w-8 h-8 bg-white/20 rounded-lg flex items-center justify-center flex-shrink-0">
                        <span className="text-white font-bold">3</span>
                      </div>
                      <div>
                        <p className="font-medium">Receba leads qualificados</p>
                        <p className="text-sm text-blue-100">Direto no seu WhatsApp</p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gray-900 text-gray-300 py-8 sm:py-12">
        <div className="container mx-auto px-3 sm:px-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-6 sm:gap-8 mb-6 sm:mb-8">
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
    </>
  );
}
