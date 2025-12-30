'use client';

import { useState, useEffect, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { ArrowLeft, Save, X, User, Mail, Phone, FileText, Award, Globe, MessageCircle, Camera, Trash2 } from 'lucide-react';
import { Broker } from '@/types/broker';
import ImageCropModal from '@/components/image-crop-modal';

export default function BrokerDetailPage() {
  const params = useParams();
  const router = useRouter();
  const brokerId = params?.id as string;

  const [broker, setBroker] = useState<Broker | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [isEditing, setIsEditing] = useState(false);
  const [uploadingPhoto, setUploadingPhoto] = useState(false);
  const [showCropModal, setShowCropModal] = useState(false);
  const [selectedImageFile, setSelectedImageFile] = useState<File | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (brokerId) {
      fetchBrokerDetails();
    }
  }, [brokerId]);

  const fetchBrokerDetails = async () => {
    try {
      setLoading(true);
      const tenantId = localStorage.getItem('tenant_id');

      if (!tenantId) {
        setError('Tenant ID não encontrado');
        return;
      }

      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;

      if (!user) {
        setError('Usuário não autenticado');
        router.push('/login');
        return;
      }

      const token = await user.getIdToken(true);

      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/brokers/${brokerId}`,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        }
      );

      if (!response.ok) {
        throw new Error('Erro ao buscar detalhes do corretor');
      }

      const data = await response.json();
      setBroker(data.data);
    } catch (err: any) {
      console.error('Erro ao buscar detalhes:', err);
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!broker) return;

    try {
      setSaving(true);
      const tenantId = localStorage.getItem('tenant_id');

      if (!tenantId) {
        setError('Tenant ID não encontrado');
        return;
      }

      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;

      if (!user) {
        setError('Usuário não autenticado');
        return;
      }

      const token = await user.getIdToken(true);

      // Extract only editable fields for update
      const {
        id,
        tenant_id,
        firebase_uid,
        created_at,
        updated_at,
        ...editableFields
      } = broker;

      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/brokers/${brokerId}`,
        {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(editableFields),
        }
      );

      if (!response.ok) {
        throw new Error('Erro ao atualizar corretor');
      }

      setIsEditing(false);
      await fetchBrokerDetails(); // Recarregar dados
    } catch (err: any) {
      console.error('Erro ao salvar:', err);
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  const handleChange = (field: keyof Broker, value: any) => {
    if (!broker) return;
    setBroker({ ...broker, [field]: value });
  };

  const handlePhotoUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate file type
    if (!['image/jpeg', 'image/png', 'image/webp'].includes(file.type)) {
      setError('Formato de arquivo inválido. Use JPEG, PNG ou WebP.');
      return;
    }

    // Validate file size (5MB)
    if (file.size > 5 * 1024 * 1024) {
      setError('Arquivo muito grande. O tamanho máximo é 5MB.');
      return;
    }

    // Open crop modal
    setSelectedImageFile(file);
    setShowCropModal(true);
  };

  const handleCropComplete = async (croppedFile: File) => {
    try {
      setUploadingPhoto(true);
      setError('');

      const tenantId = localStorage.getItem('tenant_id');
      if (!tenantId) {
        setError('Tenant ID não encontrado');
        return;
      }

      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;
      if (!user) {
        setError('Usuário não autenticado');
        return;
      }

      const token = await user.getIdToken(true);

      // Create form data
      const formData = new FormData();
      formData.append('file', croppedFile);

      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/brokers/${brokerId}/photo`,
        {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
          },
          body: formData,
        }
      );

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Erro ao fazer upload da foto');
      }

      const data = await response.json();

      // Update broker with new photo URL
      if (broker) {
        setBroker({ ...broker, photo_url: data.data.photo_url });
      }

      // Clear file input
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    } catch (err: any) {
      console.error('Erro ao fazer upload da foto:', err);
      setError(err.message);
    } finally {
      setUploadingPhoto(false);
    }
  };

  const handlePhotoDelete = async () => {
    if (!confirm('Tem certeza que deseja remover a foto do corretor?')) {
      return;
    }

    try {
      setUploadingPhoto(true);
      setError('');

      const tenantId = localStorage.getItem('tenant_id');
      if (!tenantId) {
        setError('Tenant ID não encontrado');
        return;
      }

      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;
      if (!user) {
        setError('Usuário não autenticado');
        return;
      }

      const token = await user.getIdToken(true);

      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/brokers/${brokerId}/photo`,
        {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        }
      );

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.error || 'Erro ao remover foto');
      }

      // Update broker to remove photo URL
      if (broker) {
        setBroker({ ...broker, photo_url: '' });
      }
    } catch (err: any) {
      console.error('Erro ao remover foto:', err);
      setError(err.message);
    } finally {
      setUploadingPhoto(false);
    }
  };

  if (loading) {
    return (
      <div className="p-6">
        <div className="bg-white rounded-lg shadow-sm p-12">
          <div className="text-center">
            <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mb-4"></div>
            <p className="text-gray-600">Carregando dados do corretor...</p>
          </div>
        </div>
      </div>
    );
  }

  if (error || !broker) {
    return (
      <div className="p-6">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <p className="text-red-600">{error || 'Corretor não encontrado'}</p>
        </div>
        <button
          onClick={() => router.push('/dashboard/corretores')}
          className="flex items-center gap-2 text-blue-600 hover:text-blue-700"
        >
          <ArrowLeft className="w-5 h-5" />
          Voltar para lista
        </button>
      </div>
    );
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <button
          onClick={() => router.push('/dashboard/corretores')}
          className="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-4"
        >
          <ArrowLeft className="w-5 h-5" />
          Voltar para lista
        </button>

        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="relative group">
              {broker.photo_url ? (
                <img
                  src={broker.photo_url}
                  alt={broker.name}
                  className="w-20 h-20 rounded-full object-cover"
                />
              ) : (
                <div className="w-20 h-20 rounded-full bg-blue-100 flex items-center justify-center">
                  <span className="text-3xl font-bold text-blue-600">
                    {broker.name.charAt(0).toUpperCase()}
                  </span>
                </div>
              )}
              {isEditing && (
                <div className="absolute inset-0 flex items-center justify-center gap-1 bg-black bg-opacity-50 rounded-full opacity-0 group-hover:opacity-100 transition-opacity">
                  <button
                    type="button"
                    onClick={() => fileInputRef.current?.click()}
                    disabled={uploadingPhoto}
                    className="p-2 bg-white rounded-full hover:bg-gray-100 transition-colors disabled:opacity-50"
                    title="Alterar foto"
                  >
                    <Camera className="w-4 h-4 text-gray-700" />
                  </button>
                  {broker.photo_url && (
                    <button
                      type="button"
                      onClick={handlePhotoDelete}
                      disabled={uploadingPhoto}
                      className="p-2 bg-white rounded-full hover:bg-gray-100 transition-colors disabled:opacity-50"
                      title="Remover foto"
                    >
                      <Trash2 className="w-4 h-4 text-red-600" />
                    </button>
                  )}
                </div>
              )}
              <input
                ref={fileInputRef}
                type="file"
                accept="image/jpeg,image/png,image/webp"
                onChange={handlePhotoUpload}
                className="hidden"
              />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-gray-900 mb-1">{broker.name}</h1>
              <p className="text-gray-600">{broker.email}</p>
              {uploadingPhoto && (
                <p className="text-sm text-blue-600 mt-1">Enviando foto...</p>
              )}
            </div>
          </div>
          <div className="flex gap-2">
            {!isEditing ? (
              <button
                onClick={() => setIsEditing(true)}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              >
                Editar
              </button>
            ) : (
              <>
                <button
                  onClick={() => {
                    setIsEditing(false);
                    fetchBrokerDetails(); // Restaurar dados originais
                  }}
                  className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
                >
                  Cancelar
                </button>
                <button
                  onClick={handleSubmit}
                  disabled={saving}
                  className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
                >
                  {saving ? 'Salvando...' : 'Salvar'}
                </button>
              </>
            )}
          </div>
        </div>
      </div>

      <form onSubmit={handleSubmit}>
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Main Content */}
          <div className="lg:col-span-2 space-y-6">
            {/* Informações Básicas */}
            <div className="bg-white rounded-lg shadow-sm p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4 flex items-center gap-2">
                <User className="w-5 h-5" />
                Informações Básicas
              </h2>

              <div className="grid grid-cols-2 gap-4">
                <div className="col-span-2">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Nome Completo *
                  </label>
                  <input
                    type="text"
                    value={broker.name || ''}
                    onChange={(e) => handleChange('name', e.target.value)}
                    disabled={!isEditing}
                    required
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-600"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Email *
                  </label>
                  <input
                    type="email"
                    value={broker.email || ''}
                    onChange={(e) => handleChange('email', e.target.value)}
                    disabled={!isEditing}
                    required
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-600"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Telefone
                  </label>
                  <input
                    type="tel"
                    value={broker.phone || ''}
                    onChange={(e) => handleChange('phone', e.target.value)}
                    disabled={!isEditing}
                    placeholder="(11) 98765-4321"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-600"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    CRECI *
                  </label>
                  <input
                    type="text"
                    value={broker.creci || ''}
                    onChange={(e) => handleChange('creci', e.target.value)}
                    disabled={!isEditing}
                    required
                    placeholder="12345-J/SP"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-600"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    CPF/CNPJ
                  </label>
                  <input
                    type="text"
                    value={broker.document || ''}
                    onChange={(e) => handleChange('document', e.target.value)}
                    disabled={!isEditing}
                    placeholder="000.000.000-00"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-600"
                  />
                </div>
              </div>
            </div>

            {/* Perfil Profissional */}
            <div className="bg-white rounded-lg shadow-sm p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4 flex items-center gap-2">
                <FileText className="w-5 h-5" />
                Perfil Profissional
              </h2>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Biografia
                  </label>
                  <textarea
                    value={broker.bio || ''}
                    onChange={(e) => handleChange('bio', e.target.value)}
                    disabled={!isEditing}
                    rows={4}
                    placeholder="Conte sobre sua experiência, especialidades e o que o diferencia..."
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-600"
                  />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Especialidades
                    </label>
                    <input
                      type="text"
                      value={broker.specialties || ''}
                      onChange={(e) => handleChange('specialties', e.target.value)}
                      disabled={!isEditing}
                      placeholder="Ex: Comprador, Vendedor, Aluguel"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-600"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Idiomas
                    </label>
                    <input
                      type="text"
                      value={broker.languages || ''}
                      onChange={(e) => handleChange('languages', e.target.value)}
                      disabled={!isEditing}
                      placeholder="Ex: Português, Inglês, Espanhol"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-600"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Anos de Experiência
                    </label>
                    <input
                      type="number"
                      value={broker.experience || 0}
                      onChange={(e) => handleChange('experience', parseInt(e.target.value) || 0)}
                      disabled={!isEditing}
                      min="0"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-600"
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Empresa/Imobiliária
                    </label>
                    <input
                      type="text"
                      value={broker.company || ''}
                      onChange={(e) => handleChange('company', e.target.value)}
                      disabled={!isEditing}
                      placeholder="Nome da empresa"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-600"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Website
                  </label>
                  <input
                    type="url"
                    value={broker.website || ''}
                    onChange={(e) => handleChange('website', e.target.value)}
                    disabled={!isEditing}
                    placeholder="https://seusite.com.br"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-600"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Certificações e Prêmios
                  </label>
                  <textarea
                    value={broker.certifications_awards || ''}
                    onChange={(e) => handleChange('certifications_awards', e.target.value)}
                    disabled={!isEditing}
                    rows={3}
                    placeholder="Liste suas certificações e prêmios..."
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-600"
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Status e Perfil */}
            <div className="bg-white rounded-lg shadow-sm p-6">
              <h3 className="font-bold text-gray-900 mb-4">Status e Perfil</h3>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Perfil/Função
                  </label>
                  <select
                    value={broker.role || 'broker'}
                    onChange={(e) => handleChange('role', e.target.value)}
                    disabled={!isEditing}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-600"
                  >
                    <option value="broker">Corretor</option>
                    <option value="manager">Gerente</option>
                    <option value="broker_admin">Admin Imobiliária</option>
                    <option value="platform_admin">Admin Plataforma</option>
                  </select>
                </div>

                <div className="flex items-center">
                  <input
                    type="checkbox"
                    checked={broker.is_active || false}
                    onChange={(e) => handleChange('is_active', e.target.checked)}
                    disabled={!isEditing}
                    className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500 disabled:opacity-50"
                  />
                  <label className="ml-2 text-sm text-gray-700">
                    Corretor ativo
                  </label>
                </div>
              </div>
            </div>

            {/* Estatísticas */}
            <div className="bg-white rounded-lg shadow-sm p-6">
              <h3 className="font-bold text-gray-900 mb-4 flex items-center gap-2">
                <Award className="w-5 h-5" />
                Estatísticas
              </h3>

              <div className="space-y-3 text-sm">
                <div className="flex items-center justify-between">
                  <span className="text-gray-600">Total de Vendas</span>
                  <span className="font-medium text-gray-900">{broker.total_sales || 0}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-gray-600">Anúncios Ativos</span>
                  <span className="font-medium text-gray-900">{broker.total_listings || 0}</span>
                </div>
                {broker.rating && broker.rating > 0 && (
                  <div className="flex items-center justify-between">
                    <span className="text-gray-600">Avaliação</span>
                    <span className="font-medium text-gray-900">
                      ⭐ {broker.rating.toFixed(1)} ({broker.review_count || 0})
                    </span>
                  </div>
                )}
                {broker.average_price && broker.average_price > 0 && (
                  <div className="flex items-center justify-between">
                    <span className="text-gray-600">Preço Médio</span>
                    <span className="font-medium text-gray-900">
                      R$ {broker.average_price.toLocaleString('pt-BR')}
                    </span>
                  </div>
                )}
              </div>
            </div>

            {/* URLs Públicas */}
            <div className="bg-white rounded-lg shadow-sm p-6">
              <h3 className="font-bold text-gray-900 mb-4 flex items-center gap-2">
                <Globe className="w-5 h-5" />
                Perfil Público
              </h3>

              <div className="space-y-3">
                <a
                  href={`/corretores/${broker.id}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="block text-sm text-blue-600 hover:text-blue-700 hover:underline"
                >
                  Ver perfil público →
                </a>
                <p className="text-xs text-gray-500">
                  Este é o perfil que os clientes verão quando visitarem seu site público
                </p>
              </div>
            </div>
          </div>
        </div>
      </form>

      {error && (
        <div className="fixed bottom-4 right-4 bg-red-50 border border-red-200 rounded-lg p-4 shadow-lg">
          <p className="text-red-600">{error}</p>
        </div>
      )}

      {/* Image Crop Modal */}
      {selectedImageFile && (
        <ImageCropModal
          isOpen={showCropModal}
          imageFile={selectedImageFile}
          onClose={() => {
            setShowCropModal(false);
            setSelectedImageFile(null);
            if (fileInputRef.current) {
              fileInputRef.current.value = '';
            }
          }}
          onCropComplete={handleCropComplete}
        />
      )}
    </div>
  );
}
