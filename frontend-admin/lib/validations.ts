import { z } from 'zod';

// Login validation
export const loginSchema = z.object({
  email: z.string().email('Email inválido'),
  password: z.string().min(6, 'Senha deve ter no mínimo 6 caracteres'),
});

export type LoginFormData = z.infer<typeof loginSchema>;

// User creation/update validation
export const userSchema = z.object({
  firebase_uid: z.string().min(1, 'Firebase UID é obrigatório'),
  name: z.string().min(1, 'Nome é obrigatório').max(200, 'Nome muito longo'),
  email: z.string().email('Email inválido'),
  phone: z.string().optional(),
  document: z.string().optional(),
  document_type: z.enum(['cpf', 'cnpj']).optional(),
  role: z.enum(['admin', 'manager'], {
    errorMap: () => ({ message: 'Perfil inválido' })
  } as any),
  is_active: z.boolean().default(true),
  permissions: z.array(z.string()).default([]),
});

export type UserFormData = z.infer<typeof userSchema>;

// Signup validation
export const signupSchema = z.object({
  // Tenant Type
  tenant_type: z.enum(['pf', 'pj'], {
    message: 'Selecione o tipo de negócio'
  }),

  // Tenant Info
  tenant_name: z.string().min(1, 'Nome é obrigatório'),
  document: z.string().min(11, 'CPF/CNPJ inválido'),
  business_type: z.enum([
    'corretor_autonomo',
    'imobiliaria',
    'incorporadora',
    'construtora',
    'loteadora'
  ]).optional(),
  tenant_creci: z.string().optional(),

  // Admin Info
  name: z.string().min(1, 'Nome completo é obrigatório'),
  email: z.string().email('Email válido é obrigatório'),
  password: z.string()
    .min(6, 'Senha deve ter no mínimo 6 caracteres')
    .regex(/[A-Z]/, 'Deve conter maiúscula')
    .regex(/[a-z]/, 'Deve conter minúscula')
    .regex(/[0-9]/, 'Deve conter número'),
  phone: z.string().min(10, 'Telefone inválido'),
  is_user_broker: z.boolean(),
  user_creci: z.string().optional(),
}).refine((data) => {
  // PF: tenant_creci obrigatório
  if (data.tenant_type === 'pf' && !data.tenant_creci) {
    return false;
  }

  // PF: business_type = corretor_autonomo
  if (data.tenant_type === 'pf' && data.business_type !== 'corretor_autonomo') {
    return false;
  }

  // PJ: business_type obrigatório
  if (data.tenant_type === 'pj' && !data.business_type) {
    return false;
  }

  // PJ imobiliaria: tenant_creci obrigatório
  if (data.tenant_type === 'pj' && data.business_type === 'imobiliaria' && !data.tenant_creci) {
    return false;
  }

  // is_user_broker: user_creci obrigatório
  if (data.is_user_broker && !data.user_creci) {
    return false;
  }

  return true;
}, {
  message: 'Configuração inválida',
});

export type SignupFormData = z.infer<typeof signupSchema>;

// Property import validation
export const importSchema = z.object({
  source: z.enum(['union', 'other']),
  xml: z.instanceof(File).optional(),
  xls: z.instanceof(File).optional(),
}).refine(
  (data) => data.xml || data.xls,
  { message: 'Pelo menos um arquivo (XML ou XLS) deve ser fornecido' }
);

export type ImportFormData = z.infer<typeof importSchema>;

// Owner validation
export const ownerSchema = z.object({
  name: z.string().min(1, 'Nome é obrigatório'),
  email: z.string().email('Email inválido').optional().or(z.literal('')),
  phone: z.string().optional(),
  document: z.string().optional(),
  address: z.string().optional(),
});

export type OwnerFormData = z.infer<typeof ownerSchema>;
