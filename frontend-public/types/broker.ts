// Broker types for public-facing frontend

export interface Broker {
  id: string;
  tenant_id: string;
  name: string;
  email: string;
  phone?: string;
  creci: string;

  // Profile
  photo_url?: string;
  bio?: string;
  specialties?: string;
  languages?: string;
  experience?: number;
  company?: string;
  website?: string;

  // Statistics
  total_sales?: number;
  total_listings?: number;
  average_price?: number;
  rating?: number;
  review_count?: number;

  // Timestamps
  created_at?: Date | string;
  updated_at?: Date | string;
}

export interface BrokerRole {
  broker_id: string;
  broker?: Broker;
  role: 'originating_broker' | 'listing_broker' | 'co_broker';
  is_primary: boolean;
  commission_percentage?: number;
}
