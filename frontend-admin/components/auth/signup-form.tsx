'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import axios from 'axios';
import { signInWithCustomToken } from 'firebase/auth';
import { auth } from '@/lib/firebase';
import { Building, Building2, Loader2, Eye, EyeOff, ArrowLeft, ArrowRight } from 'lucide-react';
import { signupSchema } from '@/lib/validations';
import { CreciInput } from '@/components/ui/creci-input';
import { CpfInput } from '@/components/ui/cpf-input';
import { CnpjInput } from '@/components/ui/cnpj-input';
import { PhoneInput } from '@/components/ui/phone-input';
import { StateSelect } from '@/components/ui/state-select';

interface SignupFormProps {
  onSuccess?: () => void;
  redirectTo?: string;
  variant?: 'standalone' | 'embedded';
}

interface SignupFormState {
  // Step 1: Tenant Type
  tenantType: 'pf' | 'pj' | null;

  // Step 2: Tenant Info
  tenantName: string;
  document: string;
  businessType: string;
  tenantCreci: string;
  tenantCreciUf: string;

  // Step 3: Admin Info
  name: string;
  email: string;
  phone: string;
  password: string;
  confirmPassword: string;
  isUserBroker: boolean;
  userCreci: string;
  userCreciUf: string;

  // UI
  currentStep: 1 | 2 | 3;
}

export function SignupForm({
  onSuccess,
  redirectTo = '/dashboard',
  variant = 'standalone'
}: SignupFormProps) {
  const router = useRouter();
  const [formData, setFormData] = useState<SignupFormState>({
    tenantType: null,
    tenantName: '',
    document: '',
    businessType: '',
    tenantCreci: '',
    tenantCreciUf: '',
    name: '',
    email: '',
    phone: '',
    password: '',
    confirmPassword: '',
    isUserBroker: false,
    userCreci: '',
    userCreciUf: '',
    currentStep: 1,
  });
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleFieldChange = (field: keyof SignupFormState, value: any) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleNameChange = (value: string) => {
    setFormData((prev) => ({
      ...prev,
      name: value,
      tenantName: value,
    }));
  };

  const handleSubmit = async () => {
    setError('');
    setLoading(true);

    try {
      // Validate password confirmation
      if (formData.password !== formData.confirmPassword) {
        setError('As senhas não coincidem');
        setLoading(false);
        return;
      }

      // Build request payload based on tenant type
      const payload: any = {
        email: formData.email,
        password: formData.password,
        name: formData.name,
        phone: formData.phone.replace(/\D/g, ''), // Remove formatting (only numbers)
        tenant_name: formData.tenantName,
        tenant_type: formData.tenantType,
        document: formData.document.replace(/\D/g, ''), // Remove formatting (only numbers)
      };

      if (formData.tenantType === 'pf') {
        // PF: corretor autônomo
        payload.business_type = 'corretor_autonomo';
        // Concatena CRECI com UF (ex: "12345-F/SP")
        payload.tenant_creci = formData.tenantCreci && formData.tenantCreciUf
          ? `${formData.tenantCreci}/${formData.tenantCreciUf}`
          : formData.tenantCreci;
        payload.is_user_broker = true;
        payload.user_creci = formData.tenantCreci && formData.tenantCreciUf
          ? `${formData.tenantCreci}/${formData.tenantCreciUf}`
          : formData.tenantCreci; // Same CRECI for PF
      } else if (formData.tenantType === 'pj') {
        // PJ: empresa
        payload.business_type = formData.businessType;
        payload.tenant_creci = formData.tenantCreci && formData.tenantCreciUf
          ? `${formData.tenantCreci}/${formData.tenantCreciUf}`
          : undefined;
        payload.is_user_broker = formData.isUserBroker;
        payload.user_creci = formData.isUserBroker && formData.userCreci && formData.userCreciUf
          ? `${formData.userCreci}/${formData.userCreciUf}`
          : undefined;
      }

      console.log('Payload being sent:', payload);

      // 1. Criar tenant e usuário no backend
      const signupResponse = await axios.post(
        `${process.env.NEXT_PUBLIC_API_URL}/auth/signup`,
        payload
      );

      const data = signupResponse.data;

      // 2. Sign in with custom token from backend
      await signInWithCustomToken(auth, data.firebase_token);

      // 3. Store tenant info in localStorage
      localStorage.setItem('tenant_id', data.tenant_id);
      localStorage.setItem('broker_id', data.broker_id);
      localStorage.setItem('broker_role', data.user.role);
      localStorage.setItem('broker_name', data.user.name);

      // 4. Callback de sucesso (se fornecido)
      if (onSuccess) {
        onSuccess();
      }

      // 5. Redirecionar
      router.push(redirectTo);
    } catch (err: any) {
      if (err.errors) {
        // Zod validation error
        setError(err.errors[0]?.message || 'Dados inválidos');
      } else if (err.response?.data?.error) {
        setError(err.response.data.error);
      } else if (err.response?.status === 409) {
        setError('Email já cadastrado. Faça login ou use outro email.');
      } else if (err.response?.status === 400) {
        setError('Dados inválidos. Verifique os campos e tente novamente.');
      } else {
        setError('Erro ao criar conta. Tente novamente.');
      }
    } finally {
      setLoading(false);
    }
  };

  const goToNextStep = () => {
    setError('');
    if (formData.currentStep === 1 && !formData.tenantType) {
      setError('Selecione o tipo de negócio');
      return;
    }
    if (formData.currentStep < 3) {
      setFormData({ ...formData, currentStep: (formData.currentStep + 1) as 1 | 2 | 3 });
    }
  };

  const goToPrevStep = () => {
    setError('');
    if (formData.currentStep > 1) {
      setFormData({ ...formData, currentStep: (formData.currentStep - 1) as 1 | 2 | 3 });
    }
  };

  const containerClasses = variant === 'standalone'
    ? 'min-h-screen bg-gradient-to-br from-blue-600 to-blue-800 flex items-center justify-center p-4'
    : '';

  const cardClasses = variant === 'standalone'
    ? 'bg-white rounded-2xl shadow-2xl p-8 w-full max-w-md'
    : 'w-full';

  return (
    <div className={containerClasses}>
      <div className={cardClasses}>
        {variant === 'standalone' && (
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-blue-100 rounded-full mb-4">
              <Building2 className="w-8 h-8 text-blue-600" />
            </div>
            <h1 className="text-2xl font-bold text-gray-900 mb-2">
              Cadastre-se
            </h1>
            <p className="text-gray-600">
              Comece a gerenciar seus imóveis hoje mesmo
            </p>
          </div>
        )}

        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-sm text-red-600">{error}</p>
          </div>
        )}

        {/* Step Indicator */}
        {formData.tenantType && (
          <div className="mb-8 flex items-center justify-center">
            {[1, 2, 3].map((step) => {
              // PF only has 2 steps
              if (formData.tenantType === 'pf' && step === 3) return null;

              const isActive = formData.currentStep === step;
              const isCompleted = formData.currentStep > step;

              return (
                <div key={step} className="flex items-center">
                  <div
                    className={`w-10 h-10 rounded-full flex items-center justify-center font-medium ${
                      isActive
                        ? 'bg-blue-600 text-white'
                        : isCompleted
                        ? 'bg-green-500 text-white'
                        : 'bg-gray-200 text-gray-600'
                    }`}
                  >
                    {step}
                  </div>
                  {step < (formData.tenantType === 'pf' ? 2 : 3) && (
                    <div
                      className={`w-16 h-1 mx-2 ${
                        isCompleted ? 'bg-green-500' : 'bg-gray-200'
                      }`}
                    />
                  )}
                </div>
              );
            })}
          </div>
        )}

        <div className="space-y-4">
          {/* STEP 1: Select Tenant Type */}
          {formData.currentStep === 1 && (
            <div className="space-y-6">
              <div className="text-center mb-6">
                <h2 className="text-xl font-bold text-gray-900 mb-2">
                  Que tipo de negócio você representa?
                </h2>
                <p className="text-sm text-gray-600">
                  Selecione a opção que melhor descreve seu perfil
                </p>
              </div>

              <div className="grid gap-4">
                <button
                  type="button"
                  onClick={() => handleFieldChange('tenantType', 'pf')}
                  className={`p-6 border-2 rounded-lg transition-all hover:shadow-md ${
                    formData.tenantType === 'pf'
                      ? 'border-blue-600 bg-blue-50'
                      : 'border-gray-200 hover:border-blue-300'
                  }`}
                >
                  <div className="flex items-start space-x-4">
                    <Building className="w-8 h-8 text-blue-600 flex-shrink-0" />
                    <div className="text-left">
                      <h3 className="font-semibold text-gray-900 mb-1">
                        Corretor Autônomo
                      </h3>
                      <p className="text-sm text-gray-600">
                        Você é um corretor individual com CRECI-F (Pessoa Física)
                      </p>
                    </div>
                  </div>
                </button>

                <button
                  type="button"
                  onClick={() => handleFieldChange('tenantType', 'pj')}
                  className={`p-6 border-2 rounded-lg transition-all hover:shadow-md ${
                    formData.tenantType === 'pj'
                      ? 'border-blue-600 bg-blue-50'
                      : 'border-gray-200 hover:border-blue-300'
                  }`}
                >
                  <div className="flex items-start space-x-4">
                    <Building2 className="w-8 h-8 text-blue-600 flex-shrink-0" />
                    <div className="text-left">
                      <h3 className="font-semibold text-gray-900 mb-1">
                        Imobiliária / Empresa
                      </h3>
                      <p className="text-sm text-gray-600">
                        Sua empresa possui CNPJ (Imobiliária, Incorporadora, Construtora, Loteadora)
                      </p>
                    </div>
                  </div>
                </button>
              </div>

              <button
                type="button"
                onClick={goToNextStep}
                disabled={!formData.tenantType}
                className="w-full bg-blue-600 text-white py-3 px-4 rounded-lg font-medium hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors flex items-center justify-center"
              >
                Próximo
                <ArrowRight className="w-5 h-5 ml-2" />
              </button>
            </div>
          )}

          {/* STEP 2: Tenant Info (differs by type) */}
          {formData.currentStep === 2 && formData.tenantType === 'pf' && (
            <div className="space-y-4">
              <div className="mb-6">
                <h2 className="text-xl font-bold text-gray-900 mb-2">
                  Dados do Corretor Autônomo
                </h2>
                <p className="text-sm text-gray-600">
                  Preencha suas informações pessoais e profissionais
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Nome Completo *
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => handleNameChange(e.target.value)}
                  placeholder="João Silva"
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Email *
                </label>
                <input
                  type="email"
                  value={formData.email}
                  onChange={(e) => handleFieldChange('email', e.target.value)}
                  placeholder="seu@email.com"
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Telefone (WhatsApp) *
                </label>
                <PhoneInput
                  value={formData.phone}
                  onChange={(val) => handleFieldChange('phone', val)}
                  placeholder="(11) 98765-4321"
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <CpfInput
                label="CPF *"
                value={formData.document}
                onChange={(val) => handleFieldChange('document', val)}
              />

              <StateSelect
                label="Estado do CRECI *"
                value={formData.tenantCreciUf}
                onChange={(val) => handleFieldChange('tenantCreciUf', val)}
                required
              />

              <CreciInput
                type="F"
                label="CRECI-F *"
                value={formData.tenantCreci}
                onChange={(val) => handleFieldChange('tenantCreci', val)}
                helperText="Número do seu CRECI (ex: 12345-F)"
                required
              />

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Senha *
                </label>
                <div className="relative">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    value={formData.password}
                    onChange={(e) => handleFieldChange('password', e.target.value)}
                    placeholder="Mínimo 6 caracteres"
                    className="w-full px-4 py-2 pr-10 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500"
                  >
                    {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                  </button>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Confirmar Senha *
                </label>
                <div className="relative">
                  <input
                    type={showConfirmPassword ? 'text' : 'password'}
                    value={formData.confirmPassword}
                    onChange={(e) => handleFieldChange('confirmPassword', e.target.value)}
                    placeholder="Digite a senha novamente"
                    className="w-full px-4 py-2 pr-10 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <button
                    type="button"
                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500"
                  >
                    {showConfirmPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                  </button>
                </div>
              </div>

              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  onClick={goToPrevStep}
                  className="flex-1 border border-gray-300 text-gray-700 py-3 px-4 rounded-lg font-medium hover:bg-gray-50 transition-colors flex items-center justify-center"
                >
                  <ArrowLeft className="w-5 h-5 mr-2" />
                  Voltar
                </button>
                <button
                  type="button"
                  onClick={handleSubmit}
                  disabled={loading}
                  className="flex-1 bg-blue-600 text-white py-3 px-4 rounded-lg font-medium hover:bg-blue-700 disabled:bg-gray-400 transition-colors flex items-center justify-center"
                >
                  {loading ? (
                    <>
                      <Loader2 className="w-5 h-5 mr-2 animate-spin" />
                      Criando...
                    </>
                  ) : (
                    'Criar Conta'
                  )}
                </button>
              </div>
            </div>
          )}

          {/* STEP 2: PJ Company Info */}
          {formData.currentStep === 2 && formData.tenantType === 'pj' && (
            <div className="space-y-4">
              <div className="mb-6">
                <h2 className="text-xl font-bold text-gray-900 mb-2">
                  Informações da Empresa
                </h2>
                <p className="text-sm text-gray-600">
                  Preencha os dados da sua empresa
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Tipo de Empresa *
                </label>
                <div className="space-y-2">
                  {[
                    { value: 'imobiliaria', label: 'Imobiliária', note: '(requer CRECI-J)' },
                    { value: 'incorporadora', label: 'Incorporadora', note: '(CRECI-J opcional)' },
                    { value: 'construtora', label: 'Construtora', note: '(CRECI-J opcional)' },
                    { value: 'loteadora', label: 'Loteadora', note: '(CRECI-J opcional)' },
                  ].map((type) => (
                    <label key={type.value} className="flex items-center p-3 border rounded-lg cursor-pointer hover:bg-gray-50">
                      <input
                        type="radio"
                        name="businessType"
                        value={type.value}
                        checked={formData.businessType === type.value}
                        onChange={(e) => handleFieldChange('businessType', e.target.value)}
                        className="w-4 h-4 text-blue-600"
                      />
                      <span className="ml-3 text-sm font-medium text-gray-900">{type.label}</span>
                      <span className="ml-2 text-xs text-gray-500">{type.note}</span>
                    </label>
                  ))}
                </div>
                <p className="mt-2 text-xs text-gray-500">
                  ℹ️ Apenas imobiliárias precisam de CRECI-J obrigatório. Incorporadoras, construtoras e loteadoras podem optar por ter CRECI-J se quiserem também intermediar imóveis de terceiros.
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Nome da Empresa *
                </label>
                <input
                  type="text"
                  value={formData.tenantName}
                  onChange={(e) => handleFieldChange('tenantName', e.target.value)}
                  placeholder="Imobiliária XYZ Ltda"
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <CnpjInput
                label="CNPJ *"
                value={formData.document}
                onChange={(val) => handleFieldChange('document', val)}
              />

              {formData.businessType === 'imobiliaria' && (
                <>
                  <StateSelect
                    label="Estado do CRECI *"
                    value={formData.tenantCreciUf}
                    onChange={(val) => handleFieldChange('tenantCreciUf', val)}
                    required
                  />
                  <CreciInput
                    type="J"
                    label="CRECI-J da Empresa *"
                    value={formData.tenantCreci}
                    onChange={(val) => handleFieldChange('tenantCreci', val)}
                    helperText="Número do CRECI da imobiliária (ex: 12345-J)"
                    required
                  />
                </>
              )}

              {formData.businessType && formData.businessType !== 'imobiliaria' && (
                <>
                  <StateSelect
                    label="Estado do CRECI (opcional)"
                    value={formData.tenantCreciUf}
                    onChange={(val) => handleFieldChange('tenantCreciUf', val)}
                    required={false}
                  />
                  <CreciInput
                    type="J"
                    label="CRECI-J da Empresa (opcional)"
                    value={formData.tenantCreci}
                    onChange={(val) => handleFieldChange('tenantCreci', val)}
                    helperText="Opcional - apenas necessário se sua empresa também atuar como imobiliária intermediando imóveis de terceiros"
                    required={false}
                  />
                </>
              )}

              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  onClick={goToPrevStep}
                  className="flex-1 border border-gray-300 text-gray-700 py-3 px-4 rounded-lg font-medium hover:bg-gray-50 transition-colors flex items-center justify-center"
                >
                  <ArrowLeft className="w-5 h-5 mr-2" />
                  Voltar
                </button>
                <button
                  type="button"
                  onClick={goToNextStep}
                  className="flex-1 bg-blue-600 text-white py-3 px-4 rounded-lg font-medium hover:bg-blue-700 transition-colors flex items-center justify-center"
                >
                  Próximo
                  <ArrowRight className="w-5 h-5 ml-2" />
                </button>
              </div>
            </div>
          )}

          {/* STEP 3: PJ Admin Info */}
          {formData.currentStep === 3 && formData.tenantType === 'pj' && (
            <div className="space-y-4">
              <div className="mb-6">
                <h2 className="text-xl font-bold text-gray-900 mb-2">
                  Seus Dados (Administrador Principal)
                </h2>
                <p className="text-sm text-gray-600">
                  Preencha suas informações pessoais
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Nome Completo *
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => handleFieldChange('name', e.target.value)}
                  placeholder="João Silva"
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Email *
                </label>
                <input
                  type="email"
                  value={formData.email}
                  onChange={(e) => handleFieldChange('email', e.target.value)}
                  placeholder="seu@email.com"
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Telefone (WhatsApp) *
                </label>
                <PhoneInput
                  value={formData.phone}
                  onChange={(val) => handleFieldChange('phone', val)}
                  placeholder="(11) 98765-4321"
                  className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Senha *
                </label>
                <div className="relative">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    value={formData.password}
                    onChange={(e) => handleFieldChange('password', e.target.value)}
                    placeholder="Mínimo 6 caracteres"
                    className="w-full px-4 py-2 pr-10 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500"
                  >
                    {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                  </button>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Confirmar Senha *
                </label>
                <div className="relative">
                  <input
                    type={showConfirmPassword ? 'text' : 'password'}
                    value={formData.confirmPassword}
                    onChange={(e) => handleFieldChange('confirmPassword', e.target.value)}
                    placeholder="Digite a senha novamente"
                    className="w-full px-4 py-2 pr-10 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                  <button
                    type="button"
                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500"
                  >
                    {showConfirmPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                  </button>
                </div>
              </div>

              <div className="border-t border-gray-200 pt-4">
                <label className="flex items-center space-x-3 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={formData.isUserBroker}
                    onChange={(e) => handleFieldChange('isUserBroker', e.target.checked)}
                    className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                  />
                  <span className="text-sm font-medium text-gray-700">
                    Sou corretor de imóveis (tenho CRECI)
                  </span>
                </label>

                {formData.isUserBroker && (
                  <div className="mt-4 space-y-3">
                    <StateSelect
                      label="Estado do seu CRECI *"
                      value={formData.userCreciUf}
                      onChange={(val) => handleFieldChange('userCreciUf', val)}
                      required
                    />
                    <CreciInput
                      type="F"
                      label="Seu CRECI *"
                      value={formData.userCreci}
                      onChange={(val) => handleFieldChange('userCreci', val)}
                      helperText="Seu CRECI individual (ex: 12345-F)"
                      required
                    />
                  </div>
                )}

                {!formData.isUserBroker && (
                  <p className="mt-2 ml-7 text-xs text-gray-500">
                    Você terá acesso administrativo sem aparecer como corretor
                  </p>
                )}
              </div>

              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  onClick={goToPrevStep}
                  className="flex-1 border border-gray-300 text-gray-700 py-3 px-4 rounded-lg font-medium hover:bg-gray-50 transition-colors flex items-center justify-center"
                >
                  <ArrowLeft className="w-5 h-5 mr-2" />
                  Voltar
                </button>
                <button
                  type="button"
                  onClick={handleSubmit}
                  disabled={loading}
                  className="flex-1 bg-blue-600 text-white py-3 px-4 rounded-lg font-medium hover:bg-blue-700 disabled:bg-gray-400 transition-colors flex items-center justify-center"
                >
                  {loading ? (
                    <>
                      <Loader2 className="w-5 h-5 mr-2 animate-spin" />
                      Criando...
                    </>
                  ) : (
                    'Criar Conta'
                  )}
                </button>
              </div>
            </div>
          )}
        </div>

        {variant === 'standalone' && (
          <div className="mt-6 text-center">
            <p className="text-sm text-gray-600">
              Já tem uma conta?{' '}
              <a href="/login" className="text-blue-600 hover:text-blue-700 font-medium">
                Faça login
              </a>
            </p>
          </div>
        )}

        {variant === 'standalone' && (
          <div className="mt-6 pt-6 border-t border-gray-200">
            <p className="text-xs text-gray-500 text-center">
              Ao criar uma conta, você concorda com nossos{' '}
              <a href="/termos" className="text-blue-600 hover:underline">
                Termos de Uso
              </a>{' '}
              e{' '}
              <a href="/privacidade" className="text-blue-600 hover:underline">
                Política de Privacidade
              </a>
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
