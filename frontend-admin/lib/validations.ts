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
  }),
  is_active: z.boolean().default(true),
  permissions: z.array(z.string()).default([]),
});

export type UserFormData = z.infer<typeof userSchema>;

// Signup validation
export const signupSchema = z.object({
  name: z.string().min(1, 'Nome completo é obrigatório'),
  email: z.string().email('Email válido é obrigatório'),
  password: z.string()
    .min(6, 'Senha deve ter no mínimo 6 caracteres')
    .regex(/[A-Z]/, 'Senha deve conter ao menos uma letra maiúscula')
    .regex(/[a-z]/, 'Senha deve conter ao menos uma letra minúscula')
    .regex(/[0-9]/, 'Senha deve conter ao menos um número'),
  phone: z.string().optional(),
  document: z.string().optional(),
  tenant_name: z.string().min(1, 'Nome da imobiliária é obrigatório'),
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
