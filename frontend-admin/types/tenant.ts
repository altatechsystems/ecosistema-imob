export interface Tenant {
  id: string;
  name: string;
  slug: string;

  // Contact information
  email?: string;
  phone?: string;

  // Business information
  document?: string; // CPF or CNPJ
  document_type?: 'cpf' | 'cnpj';
  business_type?: 'imobiliaria' | 'incorporadora' | 'loteadora' | 'construtora' | 'corretor_autonomo';
  creci?: string; // CRECI (Pessoa Física ou Jurídica)

  // Address
  street?: string;
  number?: string;
  complement?: string;
  neighborhood?: string;
  city?: string;
  state?: string; // UF
  zip_code?: string;
  country?: string; // default "BR"

  // Settings
  settings?: Record<string, any>;
  is_active: boolean;
  is_platform_admin?: boolean;

  // Metadata
  created_at?: string;
  updated_at?: string;
}

export interface TenantStats {
  total: number;
  active: number;
  inactive: number;
  platformAdmins: number;
}

export interface CreateTenantRequest {
  name: string;
  slug?: string; // Optional - will be generated from name if not provided
  email?: string;
  phone?: string;
  document?: string; // CPF or CNPJ
  document_type?: 'cpf' | 'cnpj';
  business_type?: 'imobiliaria' | 'incorporadora' | 'loteadora' | 'construtora' | 'corretor_autonomo';
  creci?: string;

  // Address
  street?: string;
  number?: string;
  complement?: string;
  neighborhood?: string;
  city?: string;
  state?: string;
  zip_code?: string;
  country?: string;

  settings?: Record<string, any>;
}

export interface UpdateTenantRequest {
  name?: string;
  slug?: string;
  email?: string;
  phone?: string;
  document?: string;
  document_type?: 'cpf' | 'cnpj';
  business_type?: 'imobiliaria' | 'incorporadora' | 'loteadora' | 'construtora' | 'corretor_autonomo';
  creci?: string;

  // Address
  street?: string;
  number?: string;
  complement?: string;
  neighborhood?: string;
  city?: string;
  state?: string;
  zip_code?: string;
  country?: string;

  settings?: Record<string, any>;
  is_active?: boolean;
}

export interface TenantListResponse {
  success: boolean;
  data: Tenant[];
  count: number;
}

export interface TenantResponse {
  success: boolean;
  data: Tenant;
}
