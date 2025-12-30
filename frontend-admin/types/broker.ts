export interface Broker {
  id: string;
  tenant_id: string;
  firebase_uid?: string;

  // Personal information
  name: string;
  email: string;
  phone?: string;

  // CRECI (Conselho Regional de Corretores de Imóveis)
  creci: string;

  // Document
  document?: string; // CPF ou CNPJ
  document_type?: 'cpf' | 'cnpj';

  // Role and status
  role?: 'platform_admin' | 'broker_admin' | 'broker' | 'manager';
  is_active: boolean;

  // Profile (Public Profile - similar to Zillow)
  photo_url?: string;
  bio?: string; // Biografia do corretor
  specialties?: string; // Ex: "Buyer's Agent, Listing Agent"
  languages?: string; // Ex: "Português, Inglês, Espanhol"
  experience?: number; // Anos de experiência
  company?: string; // Nome da empresa/imobiliária
  website?: string; // Website pessoal
  social_media?: string; // Links redes sociais (JSON)

  // Statistics (computed/cached for performance)
  total_sales?: number; // Total de vendas
  total_listings?: number; // Total de anúncios ativos
  average_price?: number; // Preço médio de vendas
  rating?: number; // Avaliação média (0-5)
  review_count?: number; // Número de avaliações
  last_sale_date?: string; // Data da última venda
  service_areas?: string; // Áreas de atendimento (JSON array)
  certifications_awards?: string; // Certificações e prêmios

  // Metadata
  created_at?: string;
  updated_at?: string;
}

export interface BrokerStats {
  total: number;
  active: number;
  inactive: number;
  byRole: {
    platform_admin: number;
    broker_admin: number;
    broker: number;
    manager: number;
  };
}

export enum BrokerRole {
  PLATFORM_ADMIN = 'platform_admin',
  BROKER_ADMIN = 'broker_admin',
  BROKER = 'broker',
  MANAGER = 'manager',
}

export enum BrokerSpecialty {
  BUYERS_AGENT = "Buyer's Agent",
  LISTING_AGENT = "Listing Agent",
  RENTAL_AGENT = 'Rental Agent',
  COMMERCIAL = 'Commercial',
  LUXURY = 'Luxury',
  FIRST_TIME_BUYERS = 'First Time Buyers',
  INVESTMENT_PROPERTIES = 'Investment Properties',
}
