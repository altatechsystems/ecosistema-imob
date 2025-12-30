'use client';

import * as React from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { useParams } from 'next/navigation';
import { Broker } from '@/types/broker';
import { Property } from '@/types/property';
import { PropertyCard } from '@/components/property/property-card';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { api } from '@/lib/api';
import {
  Home,
  Mail,
  Phone,
  Award,
  Star,
  Building2,
  Globe,
  ArrowLeft,
  MessageCircle,
  User,
} from 'lucide-react';

export default function BrokerProfilePage() {
  const params = useParams();
  const brokerId = params?.id as string;

  const [broker, setBroker] = React.useState<Broker | null>(null);
  const [properties, setProperties] = React.useState<Property[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);

  React.useEffect(() => {
    if (brokerId) {
      loadBrokerProfile();
    }
  }, [brokerId]);

  const loadBrokerProfile = async () => {
    try {
      setIsLoading(true);
      // TODO: Implement API endpoint to get public broker profile
      const brokerData = await api.getBrokerPublicProfile(brokerId);
      setBroker(brokerData);

      // Load broker's properties
      const propertiesData = await api.getBrokerProperties(brokerId);
      setProperties(propertiesData);
    } catch (error) {
      console.error('Failed to load broker profile:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handlePhoneClick = () => {
    if (broker?.phone) {
      window.open(`tel:${broker.phone}`, '_self');
    }
  };

  const handleEmailClick = () => {
    if (broker?.email) {
      window.open(`mailto:${broker.email}`, '_self');
    }
  };

  const handleWebsiteClick = () => {
    if (broker?.website) {
      window.open(broker.website, '_blank');
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50">
        <header className="bg-white border-b">
          <div className="container mx-auto px-4 py-4">
            <Link href="/" className="flex items-center gap-2">
              <Home className="w-8 h-8 text-blue-600" />
              <span className="text-2xl font-bold text-gray-900">Imobiliária</span>
            </Link>
          </div>
        </header>
        <div className="container mx-auto px-4 py-8">
          <div className="animate-pulse space-y-4">
            <div className="h-32 bg-gray-200 rounded-lg w-32" />
            <div className="h-8 bg-gray-200 rounded w-1/3" />
            <div className="h-6 bg-gray-200 rounded w-1/4" />
          </div>
        </div>
      </div>
    );
  }

  if (!broker) {
    return (
      <div className="min-h-screen bg-gray-50">
        <header className="bg-white border-b">
          <div className="container mx-auto px-4 py-4">
            <Link href="/" className="flex items-center gap-2">
              <Home className="w-8 h-8 text-blue-600" />
              <span className="text-2xl font-bold text-gray-900">Imobiliária</span>
            </Link>
          </div>
        </header>
        <div className="container mx-auto px-4 py-16 text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-4">Corretor não encontrado</h1>
          <Link href="/imoveis">
            <Button variant="primary">Ver Todos os Imóveis</Button>
          </Link>
        </div>
      </div>
    );
  }

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
            </nav>
          </div>
        </div>
      </header>

      <div className="container mx-auto px-4 py-8">
        {/* Back Button */}
        <Link
          href="/imoveis"
          className="inline-flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-6"
        >
          <ArrowLeft className="w-5 h-5" />
          Voltar para imóveis
        </Link>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Broker Profile Card */}
          <div className="lg:col-span-1">
            <Card variant="elevated" padding="lg" className="sticky top-24">
              {/* Broker Photo */}
              <div className="flex justify-center mb-6">
                {broker.photo_url ? (
                  <Image
                    src={broker.photo_url}
                    alt={broker.name}
                    width={160}
                    height={160}
                    className="rounded-full object-cover ring-4 ring-blue-100"
                  />
                ) : (
                  <div className="w-40 h-40 rounded-full bg-blue-100 flex items-center justify-center ring-4 ring-blue-200">
                    <User className="w-20 h-20 text-blue-600" />
                  </div>
                )}
              </div>

              {/* Broker Name and CRECI */}
              <div className="text-center mb-6">
                <h1 className="text-2xl font-bold text-gray-900 mb-2">{broker.name}</h1>
                <Badge variant="secondary" size="md">
                  <Award className="w-4 h-4 mr-1" />
                  CRECI {broker.creci}
                </Badge>
              </div>

              {/* Company */}
              {broker.company && (
                <div className="flex items-center justify-center gap-2 text-gray-700 mb-4">
                  <Building2 className="w-5 h-5" />
                  <span className="font-medium">{broker.company}</span>
                </div>
              )}

              {/* Statistics */}
              {(broker.total_listings || broker.experience || broker.rating) && (
                <div className="grid grid-cols-3 gap-4 mb-6 py-4 border-y">
                  {broker.total_listings !== undefined && broker.total_listings > 0 && (
                    <div className="text-center">
                      <p className="text-2xl font-bold text-gray-900">{broker.total_listings}</p>
                      <p className="text-xs text-gray-600">Imóveis</p>
                    </div>
                  )}
                  {broker.experience !== undefined && broker.experience > 0 && (
                    <div className="text-center">
                      <p className="text-2xl font-bold text-gray-900">{broker.experience}</p>
                      <p className="text-xs text-gray-600">Anos</p>
                    </div>
                  )}
                  {broker.rating !== undefined && broker.rating > 0 && (
                    <div className="text-center">
                      <div className="flex items-center justify-center gap-1 mb-1">
                        <Star className="w-5 h-5 text-yellow-500 fill-yellow-500" />
                        <p className="text-2xl font-bold text-gray-900">{broker.rating.toFixed(1)}</p>
                      </div>
                      {broker.review_count !== undefined && broker.review_count > 0 && (
                        <p className="text-xs text-gray-600">({broker.review_count} avaliações)</p>
                      )}
                    </div>
                  )}
                </div>
              )}

              {/* Contact Buttons */}
              <div className="space-y-2 mb-4">
                {broker.phone && (
                  <Button
                    variant="primary"
                    size="lg"
                    className="w-full"
                    onClick={handlePhoneClick}
                    leftIcon={<Phone className="w-5 h-5" />}
                  >
                    {broker.phone}
                  </Button>
                )}
                <Button
                  variant="secondary"
                  size="lg"
                  className="w-full"
                  onClick={handleEmailClick}
                  leftIcon={<Mail className="w-5 h-5" />}
                >
                  Enviar Email
                </Button>
                {broker.website && (
                  <Button
                    variant="ghost"
                    size="md"
                    className="w-full"
                    onClick={handleWebsiteClick}
                    leftIcon={<Globe className="w-5 h-5" />}
                  >
                    Website
                  </Button>
                )}
              </div>

              {/* Additional Info */}
              <div className="space-y-4 pt-4 border-t">
                {broker.specialties && (
                  <div>
                    <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">
                      Especialidades
                    </p>
                    <p className="text-sm text-gray-700">{broker.specialties}</p>
                  </div>
                )}
                {broker.languages && (
                  <div>
                    <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">
                      Idiomas
                    </p>
                    <p className="text-sm text-gray-700">{broker.languages}</p>
                  </div>
                )}
              </div>
            </Card>
          </div>

          {/* Main Content */}
          <div className="lg:col-span-2 space-y-8">
            {/* Bio */}
            {broker.bio && (
              <Card variant="bordered" padding="lg">
                <h2 className="text-xl font-bold text-gray-900 mb-4">Sobre</h2>
                <p className="text-gray-700 leading-relaxed whitespace-pre-line">{broker.bio}</p>
              </Card>
            )}

            {/* Broker's Properties */}
            <div>
              <h2 className="text-2xl font-bold text-gray-900 mb-6">
                Imóveis de {broker.name.split(' ')[0]}
              </h2>
              {properties.length > 0 ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                  {properties.map((property) => (
                    <PropertyCard key={property.id} property={property} variant="grid" />
                  ))}
                </div>
              ) : (
                <Card variant="bordered" padding="lg">
                  <p className="text-center text-gray-600">
                    Nenhum imóvel disponível no momento
                  </p>
                </Card>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
