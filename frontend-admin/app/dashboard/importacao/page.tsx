'use client';

import { useState, useEffect } from 'react';
import { Upload, FileText, AlertCircle, CheckCircle2, XCircle, RefreshCw, Download } from 'lucide-react';

interface ImportError {
  line: number;
  field: string;
  message: string;
  severity: 'error' | 'warning';
}

interface ImportResult {
  success: boolean;
  total: number;
  imported: number;
  updated: number;
  failed: number;
  errors: ImportError[];
  duration: number;
}

export default function ImportacaoPage() {
  const [xmlFile, setXmlFile] = useState<File | null>(null);
  const [xlsFile, setXlsFile] = useState<File | null>(null);
  const [importing, setImporting] = useState(false);
  const [result, setResult] = useState<ImportResult | null>(null);
  const [dragActive, setDragActive] = useState(false);
  const [source, setSource] = useState<'union' | 'other'>('union');
  const [batchId, setBatchId] = useState<string | null>(null);
  const [pollingInterval, setPollingInterval] = useState<NodeJS.Timeout | null>(null);

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === "dragenter" || e.type === "dragover") {
      setDragActive(true);
    } else if (e.type === "dragleave") {
      setDragActive(false);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      handleFiles(Array.from(e.dataTransfer.files));
    }
  };

  const handleFiles = (files: File[]) => {
    files.forEach(file => {
      if (!isValidFile(file)) return;

      const extension = file.name.substring(file.name.lastIndexOf('.')).toLowerCase();
      if (extension === '.xml') {
        setXmlFile(file);
      } else if (extension === '.xls' || extension === '.xlsx') {
        setXlsFile(file);
      }
    });
    setResult(null);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>, type: 'xml' | 'xls') => {
    if (e.target.files && e.target.files[0]) {
      const selectedFile = e.target.files[0];
      if (isValidFile(selectedFile)) {
        if (type === 'xml') {
          setXmlFile(selectedFile);
        } else {
          setXlsFile(selectedFile);
        }
        setResult(null);
      }
    }
  };

  const isValidFile = (file: File): boolean => {
    const validTypes = ['text/xml', 'application/xml', 'application/vnd.ms-excel', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'];
    const validExtensions = ['.xml', '.xls', '.xlsx'];
    const extension = file.name.substring(file.name.lastIndexOf('.')).toLowerCase();

    return validTypes.includes(file.type) || validExtensions.includes(extension);
  };

  const handleImport = async () => {
    if (!xmlFile) {
      alert('Por favor, selecione pelo menos o arquivo XML');
      return;
    }

    setImporting(true);
    const startTime = Date.now();

    try {
      const tenantId = localStorage.getItem('tenant_id');
      const brokerId = localStorage.getItem('broker_id');

      if (!tenantId) {
        throw new Error('Tenant ID não encontrado');
      }

      console.log('🔐 Getting Firebase auth token...');
      // Get Firebase auth token
      const { auth } = await import('@/lib/firebase');
      const user = auth.currentUser;
      if (!user) {
        throw new Error('Usuário não autenticado');
      }
      // Force token refresh to ensure it's valid
      const token = await user.getIdToken(true);
      console.log('✅ Token obtained');

      const formData = new FormData();
      formData.append('xml', xmlFile);

      if (xlsFile) {
        formData.append('xls', xlsFile);
      }

      formData.append('source', source);
      formData.append('created_by', brokerId || 'system');

      const url = `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/import/properties`;
      console.log('📤 Sending import request to:', url);

      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
        body: formData,
      });

      console.log('📥 Response status:', response.status, response.statusText);

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: 'Erro ao importar arquivo' }));
        console.error('❌ Error response:', errorData);
        throw new Error(errorData.error || `Erro ao importar arquivo (${response.status})`);
      }

      const data = await response.json();
      console.log('✅ Import response:', data);

      // Backend returns async response with batch_id
      // Start polling for batch status
      setBatchId(data.batch_id);
      startPolling(data.batch_id, tenantId);
    } catch (error: any) {
      console.error('❌ Erro na importação:', error);
      setResult({
        success: false,
        total: 0,
        imported: 0,
        updated: 0,
        failed: 0,
        errors: [
          {
            line: 0,
            field: 'file',
            message: error.message || 'Erro desconhecido ao processar arquivo',
            severity: 'error',
          },
        ],
        duration: (Date.now() - startTime) / 1000,
      });
      setImporting(false);
    }
  };

  const resetImport = () => {
    setXmlFile(null);
    setXlsFile(null);
    setResult(null);
  };

  const removeFile = (type: 'xml' | 'xls') => {
    if (type === 'xml') {
      setXmlFile(null);
    } else {
      setXlsFile(null);
    }
  };

  const startPolling = async (batchId: string, tenantId: string) => {
    // Get auth token once for polling
    const { auth } = await import('@/lib/firebase');
    const user = auth.currentUser;
    if (!user) return;
    // Force token refresh
    const token = await user.getIdToken(true);

    // Poll every 2 seconds
    const interval = setInterval(async () => {
      try {
        const response = await fetch(
          `${process.env.NEXT_PUBLIC_API_URL}/admin/${tenantId}/import/batches/${batchId}`,
          {
            headers: {
              'Authorization': `Bearer ${token}`,
            },
          }
        );

        if (!response.ok) {
          console.error('Erro ao buscar status do batch');
          return;
        }

        const batchData = await response.json();

        // Check if batch is completed
        if (batchData.status === 'completed' || batchData.status === 'failed') {
          stopPolling();
          setImporting(false);

          // Update result with final data
          setResult({
            success: batchData.status === 'completed',
            total: batchData.total_xml_records || 0,
            imported: batchData.total_properties_created || 0,
            updated: batchData.total_properties_matched_existing || 0,
            failed: batchData.total_errors || 0,
            errors: [], // Backend should provide detailed errors
            duration: batchData.completed_at
              ? (new Date(batchData.completed_at).getTime() - new Date(batchData.started_at).getTime()) / 1000
              : 0,
          });
        }
      } catch (error) {
        console.error('Erro no polling:', error);
      }
    }, 2000);

    setPollingInterval(interval);
  };

  const stopPolling = () => {
    if (pollingInterval) {
      clearInterval(pollingInterval);
      setPollingInterval(null);
    }
  };

  // Cleanup polling on unmount
  useEffect(() => {
    return () => {
      stopPolling();
    };
  }, [pollingInterval]);

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-2">Importação de Imóveis</h1>
        <p className="text-gray-600">Importe imóveis a partir de arquivos XML ou Excel (Union)</p>
      </div>

      {/* Progress Indicator */}
      {importing && (
        <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-blue-100 rounded-full mb-4">
              <RefreshCw className="w-8 h-8 text-blue-600 animate-spin" />
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              Importando Imóveis...
            </h3>
            <p className="text-sm text-gray-600 mb-6">
              Processando arquivo XML e criando imóveis no sistema. Isso pode levar alguns minutos.
            </p>

            {/* Progress Bar */}
            <div className="w-full bg-gray-200 rounded-full h-2 mb-4">
              <div className="bg-blue-600 h-2 rounded-full animate-pulse" style={{ width: '100%' }}></div>
            </div>

            {batchId && (
              <div className="text-xs text-gray-500">
                Batch ID: {batchId}
              </div>
            )}

            <div className="mt-6 grid grid-cols-3 gap-4 text-sm">
              <div className="bg-blue-50 rounded-lg p-3">
                <p className="text-blue-600 font-medium">Lendo XML</p>
                <div className="mt-2 flex items-center justify-center">
                  <CheckCircle2 className="w-5 h-5 text-blue-600" />
                </div>
              </div>
              <div className="bg-blue-50 rounded-lg p-3">
                <p className="text-blue-600 font-medium">Criando Imóveis</p>
                <div className="mt-2 flex items-center justify-center">
                  <RefreshCw className="w-5 h-5 text-blue-600 animate-spin" />
                </div>
              </div>
              <div className="bg-gray-50 rounded-lg p-3">
                <p className="text-gray-400 font-medium">Processando Fotos</p>
                <div className="mt-2 flex items-center justify-center">
                  <div className="w-5 h-5 border-2 border-gray-300 border-t-transparent rounded-full"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Upload Area */}
      {!result && !importing && (
        <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
          {/* Source Selector */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Origem dos Dados
            </label>
            <select
              value={source}
              onChange={(e) => setSource(e.target.value as 'union' | 'other')}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="union">Union (XML + XLS opcional)</option>
              <option value="other">Outro CRM (futuro)</option>
            </select>
          </div>

          {/* Drag and Drop Zone */}
          <div
            className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors mb-4 ${
              dragActive
                ? 'border-blue-500 bg-blue-50'
                : 'border-gray-300 hover:border-gray-400'
            }`}
            onDragEnter={handleDrag}
            onDragLeave={handleDrag}
            onDragOver={handleDrag}
            onDrop={handleDrop}
          >
            <div className="inline-flex items-center justify-center w-12 h-12 bg-gray-100 rounded-full mb-3">
              <Upload className="w-6 h-6 text-gray-400" />
            </div>
            <p className="text-sm font-semibold text-gray-900 mb-1">
              Arraste os arquivos aqui
            </p>
            <p className="text-xs text-gray-500">
              {source === 'union'
                ? 'XML (obrigatório) + XLS (opcional com dados do proprietário)'
                : 'Formato a ser definido'}
            </p>
          </div>

          {/* File Selection Buttons */}
          <div className="grid grid-cols-2 gap-4 mb-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Arquivo XML <span className="text-red-500">*</span>
              </label>
              <label className="block">
                <input
                  type="file"
                  className="hidden"
                  accept=".xml"
                  onChange={(e) => handleFileChange(e, 'xml')}
                />
                <span className="w-full px-4 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors cursor-pointer inline-flex items-center justify-center gap-2">
                  <FileText className="w-4 h-4" />
                  {xmlFile ? 'Trocar XML' : 'Selecionar XML'}
                </span>
              </label>
              {xmlFile && (
                <div className="mt-2 p-2 bg-green-50 border border-green-200 rounded-lg flex items-center justify-between">
                  <div className="flex items-center gap-2 flex-1 min-w-0">
                    <CheckCircle2 className="w-4 h-4 text-green-600 flex-shrink-0" />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-green-900 truncate">{xmlFile.name}</p>
                      <p className="text-xs text-green-700">
                        {(xmlFile.size / 1024 / 1024).toFixed(2)} MB
                      </p>
                    </div>
                  </div>
                  <button
                    onClick={() => removeFile('xml')}
                    className="ml-2 p-1 text-green-600 hover:text-green-800 flex-shrink-0"
                  >
                    <XCircle className="w-5 h-5" />
                  </button>
                </div>
              )}
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Arquivo XLS/XLSX <span className="text-gray-400">(opcional)</span>
              </label>
              <label className="block">
                <input
                  type="file"
                  className="hidden"
                  accept=".xls,.xlsx"
                  onChange={(e) => handleFileChange(e, 'xls')}
                />
                <span className="w-full px-4 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors cursor-pointer inline-flex items-center justify-center gap-2">
                  <FileText className="w-4 h-4" />
                  {xlsFile ? 'Trocar XLS' : 'Selecionar XLS'}
                </span>
              </label>
              {xlsFile && (
                <div className="mt-2 p-2 bg-green-50 border border-green-200 rounded-lg flex items-center justify-between">
                  <div className="flex items-center gap-2 flex-1 min-w-0">
                    <CheckCircle2 className="w-4 h-4 text-green-600 flex-shrink-0" />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-green-900 truncate">{xlsFile.name}</p>
                      <p className="text-xs text-green-700">
                        {(xlsFile.size / 1024 / 1024).toFixed(2)} MB
                      </p>
                    </div>
                  </div>
                  <button
                    onClick={() => removeFile('xls')}
                    className="ml-2 p-1 text-green-600 hover:text-green-800 flex-shrink-0"
                  >
                    <XCircle className="w-5 h-5" />
                  </button>
                </div>
              )}
            </div>
          </div>

          {/* Import Button */}
          {xmlFile && (
            <div className="flex justify-end gap-3">
              <button
                onClick={resetImport}
                className="px-4 py-2 text-sm border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
              >
                Limpar Tudo
              </button>
              <button
                onClick={handleImport}
                disabled={importing}
                className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
              >
                {importing ? (
                  <>
                    <RefreshCw className="w-4 h-4 animate-spin" />
                    Importando...
                  </>
                ) : (
                  <>
                    <Upload className="w-4 h-4" />
                    Iniciar Importação
                  </>
                )}
              </button>
            </div>
          )}
        </div>
      )}

      {/* Results */}
      {result && (
        <div className="space-y-6">
          {/* Summary */}
          <div className="bg-white rounded-lg shadow-sm p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-bold text-gray-900">Resultado da Importação</h2>
              <button
                onClick={resetImport}
                className="px-4 py-2 text-sm border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors flex items-center gap-2"
              >
                <RefreshCw className="w-4 h-4" />
                Nova Importação
              </button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
              <div className="bg-blue-50 rounded-lg p-4">
                <div className="flex items-center gap-2 mb-1">
                  <FileText className="w-4 h-4 text-blue-600" />
                  <p className="text-sm font-medium text-blue-900">Total</p>
                </div>
                <p className="text-2xl font-bold text-blue-600">{result.total}</p>
              </div>

              <div className="bg-green-50 rounded-lg p-4">
                <div className="flex items-center gap-2 mb-1">
                  <CheckCircle2 className="w-4 h-4 text-green-600" />
                  <p className="text-sm font-medium text-green-900">Importados</p>
                </div>
                <p className="text-2xl font-bold text-green-600">{result.imported}</p>
              </div>

              <div className="bg-orange-50 rounded-lg p-4">
                <div className="flex items-center gap-2 mb-1">
                  <RefreshCw className="w-4 h-4 text-orange-600" />
                  <p className="text-sm font-medium text-orange-900">Atualizados</p>
                </div>
                <p className="text-2xl font-bold text-orange-600">{result.updated}</p>
              </div>

              <div className="bg-red-50 rounded-lg p-4">
                <div className="flex items-center gap-2 mb-1">
                  <XCircle className="w-4 h-4 text-red-600" />
                  <p className="text-sm font-medium text-red-900">Falhas</p>
                </div>
                <p className="text-2xl font-bold text-red-600">{result.failed}</p>
              </div>
            </div>

            <div className="flex items-center gap-4 text-sm text-gray-600">
              <div className="flex items-center gap-2">
                {result.success ? (
                  <CheckCircle2 className="w-5 h-5 text-green-500" />
                ) : (
                  <AlertCircle className="w-5 h-5 text-red-500" />
                )}
                <span>
                  {result.success
                    ? 'Importação concluída com sucesso'
                    : 'Importação concluída com erros'}
                </span>
              </div>
              <div>
                Tempo de processamento: {result.duration.toFixed(2)}s
              </div>
            </div>
          </div>

          {/* Errors and Warnings */}
          {result.errors.length > 0 && (
            <div className="bg-white rounded-lg shadow-sm p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold text-gray-900">
                  Erros e Avisos ({result.errors.length})
                </h3>
                <button className="px-3 py-1 text-sm text-blue-600 hover:bg-blue-50 rounded-lg transition-colors flex items-center gap-2">
                  <Download className="w-4 h-4" />
                  Exportar Log
                </button>
              </div>

              <div className="space-y-2 max-h-96 overflow-y-auto">
                {result.errors.map((error, index) => (
                  <div
                    key={index}
                    className={`p-3 rounded-lg border ${
                      error.severity === 'error'
                        ? 'bg-red-50 border-red-200'
                        : 'bg-yellow-50 border-yellow-200'
                    }`}
                  >
                    <div className="flex items-start gap-2">
                      {error.severity === 'error' ? (
                        <XCircle className="w-4 h-4 text-red-600 mt-0.5 flex-shrink-0" />
                      ) : (
                        <AlertCircle className="w-4 h-4 text-yellow-600 mt-0.5 flex-shrink-0" />
                      )}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="text-xs font-medium text-gray-500">
                            Linha {error.line}
                          </span>
                          {error.field && (
                            <>
                              <span className="text-gray-300">•</span>
                              <span className="text-xs font-medium text-gray-500">
                                Campo: {error.field}
                              </span>
                            </>
                          )}
                        </div>
                        <p className="text-sm text-gray-900">{error.message}</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Instructions */}
      {!result && !xmlFile && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
          <div className="flex gap-3">
            <AlertCircle className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
            <div>
              <h3 className="font-semibold text-blue-900 mb-2">
                Instruções para Importação Union
              </h3>
              <ul className="text-sm text-blue-800 space-y-1">
                <li>• <strong>XML obrigatório:</strong> Contém dados principais dos imóveis (endereço, características, fotos)</li>
                <li>• <strong>XLS opcional:</strong> Enriquece com dados do proprietário (nome, telefone, email, observações)</li>
                <li>• Os imóveis serão identificados por referência única</li>
                <li>• Imóveis duplicados serão detectados automaticamente</li>
                <li>• Proprietários sem dados completos serão marcados como "incompletos"</li>
                <li>• Tamanho máximo por arquivo: 50 MB</li>
              </ul>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
