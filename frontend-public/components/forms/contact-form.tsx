'use client';

import * as React from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import { Button } from '@/components/ui/button';
import { LeadChannel, CreateLeadRequest } from '@/types/lead';
import { api } from '@/lib/api';
import { isValidEmail, isValidPhone } from '@/lib/utils';

const contactFormSchema = z.object({
  name: z.string().min(3, 'Nome deve ter pelo menos 3 caracteres'),
  email: z.string().optional().refine((email) => !email || isValidEmail(email), {
    message: 'Email inválido',
  }),
  phone: z.string().min(10, 'Telefone inválido').refine(isValidPhone, {
    message: 'Telefone deve ter 10 ou 11 dígitos',
  }),
  message: z.string().optional(),
  consent: z.boolean().refine((val) => val === true, {
    message: 'Você deve aceitar os termos de privacidade',
  }),
});

type ContactFormData = z.infer<typeof contactFormSchema>;

export interface ContactFormProps {
  propertyId: string;
  propertyTitle?: string;
  channel?: LeadChannel;
  onSuccess?: () => void;
  onError?: (error: Error) => void;
}

export function ContactForm({
  propertyId,
  propertyTitle,
  channel = LeadChannel.FORM,
  onSuccess,
  onError,
}: ContactFormProps) {
  const [isSubmitting, setIsSubmitting] = React.useState(false);
  const [submitSuccess, setSubmitSuccess] = React.useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<ContactFormData>({
    resolver: zodResolver(contactFormSchema),
    defaultValues: {
      name: '',
      email: '',
      phone: '',
      message: '',
      consent: false,
    },
  });

  const onSubmit = async (data: ContactFormData) => {
    setIsSubmitting(true);
    setSubmitSuccess(false);

    try {
      const leadData: CreateLeadRequest = {
        property_id: propertyId,
        name: data.name,
        email: data.email || undefined,
        phone: data.phone,
        message: data.message || undefined,
        channel,
        consent_text: 'Autorizo o uso dos meus dados para contato conforme a LGPD.',
      };

      await api.createLead(leadData);

      setSubmitSuccess(true);
      reset();

      if (onSuccess) {
        onSuccess();
      }

      // Auto-hide success message after 5 seconds
      setTimeout(() => {
        setSubmitSuccess(false);
      }, 5000);
    } catch (error) {
      console.error('Failed to submit lead:', error);
      if (onError) {
        onError(error as Error);
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card variant="bordered" padding="lg">
      <div className="mb-6">
        <h3 className="text-2xl font-bold text-gray-900 mb-2">
          Entre em Contato
        </h3>
        {propertyTitle && (
          <p className="text-gray-600">
            Interessado em: <span className="font-medium">{propertyTitle}</span>
          </p>
        )}
      </div>

      {submitSuccess && (
        <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
          <p className="text-green-800 font-medium">
            Mensagem enviada com sucesso!
          </p>
          <p className="text-green-700 text-sm mt-1">
            Entraremos em contato em breve.
          </p>
        </div>
      )}

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <Input
          label="Nome completo *"
          placeholder="Seu nome"
          error={errors.name?.message}
          {...register('name')}
        />

        <Input
          label="Email"
          type="email"
          placeholder="seu@email.com"
          error={errors.email?.message}
          helperText="Opcional"
          {...register('email')}
        />

        <Input
          label="Telefone/WhatsApp *"
          type="tel"
          placeholder="(11) 98765-4321"
          error={errors.phone?.message}
          {...register('phone')}
        />

        <Textarea
          label="Mensagem"
          placeholder="Deixe sua mensagem ou dúvida"
          rows={4}
          error={errors.message?.message}
          helperText="Opcional"
          {...register('message')}
        />

        <Checkbox
          label="Autorizo o uso dos meus dados para contato conforme a LGPD"
          error={errors.consent?.message}
          {...register('consent')}
        />

        <Button
          type="submit"
          variant="primary"
          size="lg"
          className="w-full"
          isLoading={isSubmitting}
          disabled={isSubmitting || submitSuccess}
        >
          {isSubmitting ? 'Enviando...' : 'Enviar Mensagem'}
        </Button>

        <p className="text-xs text-gray-500 text-center">
          Ao enviar este formulário, você concorda com nossa Política de Privacidade
          e o uso dos seus dados conforme a Lei Geral de Proteção de Dados (LGPD).
        </p>
      </form>
    </Card>
  );
}
