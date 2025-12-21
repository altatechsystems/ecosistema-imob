# Plano de Implementação - Ecossistema Imobiliário MVP

**Data**: 2025-12-21
**Status**: 📋 100% Documentação / 0% Código Implementado
**Estimativa Total**: 310-390 horas (~2-3 meses com 1 dev sênior)

---

## 🎯 Resumo Executivo

Este é um **projeto greenfield** com documentação técnica completa (98KB de specs + 11 prompts detalhados) mas **NENHUM código fonte foi escrito ainda**.

**Arquivos existentes**:
- ✅ AI_DEV_DIRECTIVE.md (25 seções - arquitetura completa)
- ✅ 11 PROMPTs de implementação (352KB total)
- ✅ Dados de teste (XML 8.2MB + XLS 1.1MB)
- ❌ `/backend` - NÃO EXISTE
- ❌ `/frontend-public` - NÃO EXISTE
- ❌ `/frontend-admin` - NÃO EXISTE

**O que precisa ser construído**: 195 arquivos de código (~19.300 linhas)

---

## 📊 Status Atual vs. Necessário

### Backend (Go)
| Componente | Status | Arquivos Necessários |
|-----------|--------|---------------------|
| Models | ❌ 0% | 8 arquivos (Tenant, Broker, Property, Listing, Owner, Lead, etc.) |
| Repositories | ❌ 0% | 6+ arquivos (CRUD + queries complexas) |
| Services | ❌ 0% | 5+ arquivos (lógica de negócio) |
| Handlers | ❌ 0% | 7+ arquivos (40+ endpoints REST) |
| Middlewares | ❌ 0% | 3 arquivos (auth, tenant isolation, logging) |
| Import Adapters | ❌ 0% | 2 arquivos (Union XML + XLS) |
| Infraestrutura | ❌ 0% | Firebase SDK, Firestore, GCS, image processing |

### Frontend Public (Next.js)
| Componente | Status | Arquivos Necessários |
|-----------|--------|---------------------|
| Páginas | ❌ 0% | 5 páginas (Home, Buscar, Detalhe, Privacidade, Termos) |
| Componentes UI | ❌ 0% | 30+ componentes (shadcn/ui + custom) |
| API Client | ❌ 0% | 1 arquivo (axios wrapper com auth) |
| Hooks | ❌ 0% | 2+ arquivos (React Query hooks) |
| Config | ❌ 0% | package.json, next.config.js, tailwind.config.ts |

### Frontend Admin (Next.js)
| Componente | Status | Arquivos Necessários |
|-----------|--------|---------------------|
| Páginas | ❌ 0% | 10+ páginas (Dashboard, Imóveis, Leads, Parcerias, Import, Config) |
| Componentes | ❌ 0% | 40+ componentes (forms, tables, modals) |
| Auth | ❌ 0% | Firebase Auth client SDK + AuthContext |
| Middleware | ❌ 0% | Route protection |

### Infraestrutura (GCP)
| Recurso | Status | Ação Necessária |
|---------|--------|----------------|
| Firebase Project | ❌ Não criado | Criar projeto no console Firebase |
| Firestore Database | ❌ Não criado | Inicializar modo nativo |
| Cloud Storage | ❌ Não criado | Criar bucket público para imagens |
| Cloud Run | ❌ Não criado | Deploy do backend |
| Firestore Indexes | ❌ Não criado | Deploy de firestore.indexes.json |
| Security Rules | ❌ Não criado | Deploy de firestore.rules |

---

## 🗓️ Fases de Implementação

### Fase 1: Foundation & Autenticação (40-50h)

**Objetivo**: Estabelecer base do backend com autenticação multi-tenant

**Entregas**:
1. **Setup do Projeto Go** (8h)
   - Estrutura de diretórios (cmd, internal, pkg)
   - go.mod com dependências (Gin, Firebase Admin, Firestore, GCS)
   - main.go com servidor básico
   - Logging e error handling patterns

2. **Firebase & Firestore** (6h)
   - Criar projeto Firebase
   - Habilitar Authentication (email/password)
   - Criar Firestore database
   - Criar bucket GCS
   - Service account + IAM
   - Inicializar Firebase Admin SDK

3. **Core Models** (10h)
   - `internal/models/tenant.go` - Tenant + TenantSettings
   - `internal/models/broker.go` - Broker com validação de telefone
   - `internal/models/property.go` - Property (30+ campos)
   - `internal/models/listing.go` - Listing com Photo/Video
   - `internal/models/owner.go` - Owner
   - `internal/models/property_broker_role.go` - Co-corretagem
   - `internal/models/lead.go` - Lead com LGPD
   - `internal/models/activity_log.go` - Auditoria

4. **Sistema de Autenticação** (12h)
   - POST /api/v1/auth/signup (cria tenant + primeiro broker)
   - POST /api/v1/auth/login (Firebase Auth)
   - POST /api/v1/auth/refresh
   - Custom Claims (tenant_id, role)
   - `internal/middleware/auth_middleware.go`
   - `internal/middleware/tenant_middleware.go`

5. **Repositories Básicos** (8h)
   - TenantRepository (CRUD)
   - BrokerRepository (CRUD, find by email/phone)
   - PropertyRepository (CRUD base)
   - ListingRepository (CRUD base)
   - OwnerRepository (CRUD base)
   - LeadRepository (CRUD base)
   - ActivityLogRepository

6. **Handlers Básicos** (6h)
   - AuthHandler
   - TenantHandler
   - BrokerHandler
   - Health check endpoint

**Critérios de Validação**:
- ✅ Broker pode criar conta (tenant criado automaticamente)
- ✅ Broker pode fazer login (recebe JWT com tenant_id e role)
- ✅ Middleware valida tenant_id em rotas protegidas
- ✅ Acesso cross-tenant bloqueado
- ✅ ActivityLog registra eventos de auth

**Arquivos**: ~25 arquivos (~3000 linhas)

---

### 📋 Detalhamento: Arquitetura Blockchain-Ready (ActivityLog)

**Contexto Estratégico**:
A COFECI Resolução 1551/2025 (PITD - Plataformas de Intermediação de Transações Digitais) está **SUSPENSA** e não tem efeito legal vigente. Entretanto, implementaremos uma arquitetura "blockchain-ready" desde o MVP para garantir vantagem competitiva de 18-22 meses quando (e se) a regulamentação for aprovada.

**Estratégia**:
- ✅ Hash SHA-256 **ATIVO** desde o MVP (auditoria + preparação)
- ✅ Cadeia de blocos **LOCAL** (prev_hash → hash)
- ✅ Campos **RESERVADOS** para blockchain (blockchain_tx, token_id)
- ❌ Registro on-chain **INATIVO** até regulamentação

#### ActivityLog Model (internal/models/activity_log.go)

```go
package models

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"
)

// ActivityLog mantém trilha de auditoria imutável de todas as ações na plataforma.
//
// ESTRATÉGIA BLOCKCHAIN-READY:
// - Hash SHA-256 garante imutabilidade desde o MVP (compliance + preparação)
// - Campos blockchain_tx e token_id preparados para PITD (Res. 1551/2025)
// - Se regulamentação for aprovada, ativar blockchain em 2-4 semanas
// - Vantagem competitiva: 18-22 meses à frente de concorrentes
//
// STATUS REGULATÓRIO (Dezembro 2025):
// - Resolução COFECI 1551/2025 SUSPENSA (sem efeito legal)
// - blockchain_tx permanece NULL até regularização
// - Hash SHA-256 já entrega valor (auditoria, LGPD, co-corretagem)
type ActivityLog struct {
    // Identificação
    ID        string    `firestore:"id" json:"id"`
    Event     string    `firestore:"event" json:"event"` // Ex: "lead_distributed", "partnership_accepted"
    Timestamp time.Time `firestore:"timestamp" json:"timestamp"`

    // Multi-Tenancy
    TenantID  string `firestore:"tenant_id" json:"tenant_id"`
    UserID    string `firestore:"user_id" json:"user_id"` // Broker que executou a ação

    // ✅ BLOCKCHAIN-READY: Campos para auditoria imutável
    Hash         string  `firestore:"hash" json:"hash"`                   // SHA-256 do evento (ATIVO no MVP)
    PrevHash     string  `firestore:"prev_hash" json:"prev_hash"`         // Hash do evento anterior (cadeia local)
    BlockchainTx *string `firestore:"blockchain_tx,omitempty" json:"blockchain_tx,omitempty"` // NULL no MVP
    TokenID      *string `firestore:"token_id,omitempty" json:"token_id,omitempty"`           // NULL no MVP

    // Dados do evento (JSON flexível)
    Data map[string]interface{} `firestore:"data" json:"data"`

    // Metadata
    IPAddress  string            `firestore:"ip_address,omitempty" json:"ip_address,omitempty"`
    UserAgent  string            `firestore:"user_agent,omitempty" json:"user_agent,omitempty"`
    Metadata   map[string]string `firestore:"metadata,omitempty" json:"metadata,omitempty"`
}

// GenerateHash calcula SHA-256 do evento para garantir imutabilidade
// Formato: hash(event + data + timestamp + prev_hash)
func (log *ActivityLog) GenerateHash() string {
    // Serializar dados do evento de forma determinística
    dataStr := fmt.Sprintf("%v", log.Data) // Simplificado - produção deve usar JSON ordenado

    // Concatenar campos críticos
    payload := fmt.Sprintf("%s|%s|%s|%s|%s",
        log.Event,
        dataStr,
        log.Timestamp.Format(time.RFC3339Nano),
        log.PrevHash,
        log.TenantID,
    )

    // Calcular SHA-256
    hash := sha256.Sum256([]byte(payload))
    return hex.EncodeToString(hash[:])
}

// Validate verifica integridade da cadeia de blocos
func (log *ActivityLog) Validate(expectedPrevHash string) bool {
    // Verificar se prev_hash aponta para o evento anterior
    if log.PrevHash != expectedPrevHash {
        return false
    }

    // Recalcular hash e comparar com o armazenado
    calculatedHash := log.GenerateHash()
    return calculatedHash == log.Hash
}

// IsBlockchainEnabled indica se a feature de blockchain on-chain está ativa
// (monitorar regulamentação PITD - Resolução 1551/2025)
func (log *ActivityLog) IsBlockchainEnabled() bool {
    return log.BlockchainTx != nil
}

// ActivateBlockchain ativa registro on-chain (SOMENTE após regulamentação PITD)
// IMPORTANTE: NÃO chamar no MVP - campo permanece NULL até COFECI reativar Res. 1551/2025
func (log *ActivityLog) ActivateBlockchain(txHash string, tokenID string) {
    log.BlockchainTx = &txHash
    log.TokenID = &tokenID
}
```

#### ActivityLogRepository (internal/repository/activity_log_repository.go)

```go
package repository

import (
    "context"
    "fmt"
    "imovel-hub/internal/models"
    "cloud.google.com/go/firestore"
)

type ActivityLogRepository struct {
    client *firestore.Client
}

func NewActivityLogRepository(client *firestore.Client) *ActivityLogRepository {
    return &ActivityLogRepository{client: client}
}

// Create registra novo evento na cadeia de blocos local
func (r *ActivityLogRepository) Create(ctx context.Context, log *models.ActivityLog) error {
    // 1. Buscar último evento do tenant para obter prev_hash
    lastLog, err := r.GetLastLog(ctx, log.TenantID)
    if err != nil && err != ErrNotFound {
        return fmt.Errorf("failed to get last log: %w", err)
    }

    // 2. Configurar prev_hash
    if lastLog != nil {
        log.PrevHash = lastLog.Hash
    } else {
        log.PrevHash = "genesis" // Primeiro evento do tenant
    }

    // 3. Gerar hash SHA-256
    log.Hash = log.GenerateHash()

    // 4. Salvar no Firestore
    _, err = r.client.Collection("activity_logs").Doc(log.ID).Set(ctx, log)
    if err != nil {
        return fmt.Errorf("failed to create activity log: %w", err)
    }

    return nil
}

// GetLastLog retorna o último evento do tenant (para cadeia de blocos)
func (r *ActivityLogRepository) GetLastLog(ctx context.Context, tenantID string) (*models.ActivityLog, error) {
    docs, err := r.client.Collection("activity_logs").
        Where("tenant_id", "==", tenantID).
        OrderBy("timestamp", firestore.Desc).
        Limit(1).
        Documents(ctx).
        GetAll()

    if err != nil {
        return nil, err
    }

    if len(docs) == 0 {
        return nil, ErrNotFound
    }

    var log models.ActivityLog
    if err := docs[0].DataTo(&log); err != nil {
        return nil, err
    }

    return &log, nil
}

// ValidateChain verifica integridade da cadeia de blocos local (auditoria)
func (r *ActivityLogRepository) ValidateChain(ctx context.Context, tenantID string) (bool, error) {
    docs, err := r.client.Collection("activity_logs").
        Where("tenant_id", "==", tenantID).
        OrderBy("timestamp", firestore.Asc).
        Documents(ctx).
        GetAll()

    if err != nil {
        return false, err
    }

    var prevHash = "genesis"
    for _, doc := range docs {
        var log models.ActivityLog
        if err := doc.DataTo(&log); err != nil {
            return false, err
        }

        // Validar se prev_hash aponta para o anterior
        if !log.Validate(prevHash) {
            return false, fmt.Errorf("chain broken at event %s", log.ID)
        }

        prevHash = log.Hash
    }

    return true, nil
}
```

#### Exemplo de Uso (internal/services/lead_service.go)

```go
// DistributeLead distribui lead para corretor e registra na auditoria
func (s *LeadService) DistributeLead(ctx context.Context, leadID, brokerID string) error {
    // ... lógica de distribuição de lead ...

    // Registrar evento na cadeia de blocos local (auditoria)
    activityLog := &models.ActivityLog{
        ID:        uuid.New().String(),
        Event:     "lead_distributed",
        Timestamp: time.Now(),
        TenantID:  lead.TenantID,
        UserID:    brokerID,
        Data: map[string]interface{}{
            "lead_id":         lead.ID,
            "lead_name":       lead.Name,
            "lead_phone":      lead.Phone,
            "property_id":     lead.PropertyID,
            "broker_id":       brokerID,
            "broker_name":     broker.DisplayName,
            "broker_creci":    broker.CRECI,
            "distribution_source": "automatic", // ou "manual", "marketplace"
        },
        IPAddress: ctx.Value("ip_address").(string),
        UserAgent: ctx.Value("user_agent").(string),
    }

    // Hash SHA-256 será gerado automaticamente no repository
    if err := s.activityLogRepo.Create(ctx, activityLog); err != nil {
        return fmt.Errorf("failed to log activity: %w", err)
    }

    return nil
}
```

#### Firestore Index (firestore.indexes.json)

```json
{
  "indexes": [
    {
      "collectionGroup": "activity_logs",
      "queryScope": "COLLECTION",
      "fields": [
        { "fieldPath": "tenant_id", "order": "ASCENDING" },
        { "fieldPath": "timestamp", "order": "DESCENDING" }
      ]
    },
    {
      "collectionGroup": "activity_logs",
      "queryScope": "COLLECTION",
      "fields": [
        { "fieldPath": "tenant_id", "order": "ASCENDING" },
        { "fieldPath": "event", "order": "ASCENDING" },
        { "fieldPath": "timestamp", "order": "DESCENDING" }
      ]
    }
  ]
}
```

#### Vantagem Competitiva (Quando PITD for Regulamentado)

**Concorrentes (ZAP, VivaReal, CRMs)**:
1. Redesenhar schema do banco (6-12 meses)
2. Implementar hash e auditoria (3-6 meses)
3. Migrar dados históricos (2-4 meses)
4. Integrar blockchain (4-8 meses)
**Total**: 12-24 meses

**Nossa Plataforma**:
1. ✅ Schema já tem campos blockchain (prev_hash, hash, blockchain_tx)
2. ✅ Hash SHA-256 já está ativo desde o MVP
3. ✅ Cadeia de blocos local já valida integridade
4. Apenas ativar integração blockchain (2-4 semanas)
**Total**: 2-4 semanas

**First-Mover Advantage**: 18-22 meses à frente

---

### Fase 2: Sistema de Importação (50-60h)

**Objetivo**: Importar imóveis do Union CRM (XML + XLS)

**Entregas**:
1. **Import Adapters** (12h)
   - `internal/adapters/union_xml_adapter.go` - Parser do XML
   - `internal/adapters/union_xls_adapter.go` - Parser do XLS
   - Interface de adapter (extensibilidade)
   - Normalização de dados
   - Validação de campos

2. **Lógica de Deduplicação** (8h)
   - Geração de fingerprint (hash de endereço + tipo + área)
   - Match por external_source + external_id
   - Match heurístico (similaridade de endereço)
   - Flag possible_duplicate
   - DeduplicationService

3. **Pipeline de Processamento de Imagens** (14h)
   - Download de imagens de URLs
   - Upload para GCS temp
   - Conversão para WebP (3 tamanhos: 400px, 800px, 1600px)
   - Upload de imagens processadas para GCS público
   - Exclusão de originais
   - Error handling (skip imagens falhadas)
   - Processamento concorrente (goroutines, limite 10 paralelas)
   - Integração com `github.com/disintegration/imaging`

4. **Criação de Property/Owner/Listing** (10h)
   - Criar Owner (placeholder se dados faltando)
   - Criar Property com verificação de deduplicação
   - Criar Listing com fotos
   - Atribuir canonical listing
   - Criar PropertyBrokerRole (originating_broker)
   - Handling de transações (rollback em falha)

5. **Gerenciamento de Batch** (8h)
   - Model ImportBatch
   - Subcollection import_errors
   - Tracking de status (started, in_progress, completed, failed)
   - Geração de sumário (contagens, duplicados, erros)
   - Log de erros (parsing, validação, imagens)

6. **Endpoints de Import** (8h)
   - POST /api/v1/tenants/{tenantId}/import?source=union
   - GET /api/v1/tenants/{tenantId}/import/batches
   - GET /api/v1/tenants/{tenantId}/import/batches/{batchId}
   - GET /api/v1/tenants/{tenantId}/import/batches/{batchId}/errors
   - Handling de upload (multipart/form-data)

**Critérios de Validação**:
- ✅ XML parseado corretamente (todos os campos extraídos)
- ✅ XLS enriquece dados (dados do owner)
- ✅ Imagens baixadas e convertidas para WebP (3 tamanhos)
- ✅ Imagens enviadas para GCS
- ✅ Properties criadas sem duplicação
- ✅ Listings criadas com fotos
- ✅ Canonical listing atribuído
- ✅ PropertyBrokerRole criado (originating_broker)
- ✅ Erros salvos em import_errors
- ✅ Sumário de batch mostra contagens

**Arquivos**: ~15 arquivos (~2500 linhas)

---

### Fase 3: Frontend Público - Páginas Core (50-60h)

**Objetivo**: Site público para busca e visualização de imóveis

**Entregas**:
1. **Setup do Projeto Next.js** (8h)
   - Inicializar Next.js 14 (App Router)
   - Instalar dependências (shadcn/ui, Tailwind, React Query, Zustand)
   - Configurar Tailwind com tema customizado
   - Setup shadcn/ui components
   - Configurar TypeScript
   - Setup API client
   - Variáveis de ambiente

2. **Componentes UI** (12h)
   - Instalar 15 componentes shadcn/ui (Button, Card, Input, Select, etc.)
   - LogoPlaceholder
   - PropertyCard
   - PropertyGallery (Embla Carousel para mobile)
   - SearchFilters (Sheet modal em mobile, sidebar em desktop)
   - Header (navegação responsiva)
   - Footer
   - WhatsAppButton
   - ContactForm (com consentimento LGPD)

3. **Homepage** (8h)
   - Seção hero com widget de busca
   - Grid de imóveis em destaque
   - Seções de call-to-action
   - Layout responsivo (mobile-first)
   - Branding dinâmico (fetch da API)

4. **Página de Busca** (12h)
   - SearchFilters (tipo, cidade, bairro, preço, quartos, vagas)
   - Grid de resultados (PropertyCard)
   - Paginação
   - Seletor de ordenação (recente, preço asc/desc)
   - Sync com URL params (SEO-friendly)
   - Estados de loading (Skeleton)
   - Empty state
   - Integração React Query

5. **Página de Detalhe do Imóvel** (14h)
   - Implementação SSR (getStaticProps + getStaticPaths)
   - Roteamento dinâmico de slug
   - PropertyGallery (touch-friendly em mobile, grid em desktop)
   - Seção PropertyInfo
   - Grid de características
   - Botão WhatsApp (fluxo de criação de Lead)
   - Formulário de contato
   - Meta tags dinâmicas (title, description, OpenGraph, JSON-LD)
   - Breadcrumbs
   - Botões de compartilhamento

6. **Páginas de Conformidade LGPD** (6h)
   - Página de Política de Privacidade
   - Página de Termos de Uso
   - Banner de consentimento de cookies
   - Formulários com conformidade LGPD

**Critérios de Validação**:
- ✅ Homepage carrega com imóveis em destaque
- ✅ Página de busca filtra imóveis corretamente
- ✅ Página de detalhe renderiza com SSR (view source tem meta tags)
- ✅ Meta tags presentes (title, description, OpenGraph, JSON-LD)
- ✅ Botão WhatsApp cria Lead ANTES do redirect
- ✅ Formulário de contato valida e cria Lead
- ✅ Responsivo em 6 dispositivos (iPhone SE, 12, 14 Pro Max, iPad, iPad Pro, Desktop)
- ✅ Alvos de toque mínimo 44px
- ✅ Galeria com swipe funciona em mobile
- ✅ Performance: bundle < 200KB, LCP < 2.5s

**Arquivos**: ~40 arquivos (~4000 linhas)

---

### Fase 4: Frontend Admin - Dashboard & Gestão (60-70h)

**Objetivo**: Dashboard admin para corretores gerenciarem imóveis, leads e importações

**Entregas**:
1. **Setup do Projeto Next.js** (6h)
   - Inicializar Next.js 14 (projeto separado)
   - Instalar dependências (Firebase Auth SDK, shadcn/ui, React Query, react-hook-form, zod, Sonner, react-dropzone)
   - Configurar Firebase Auth
   - Setup AuthContext
   - Middleware para proteção de rotas
   - Config Tailwind

2. **Autenticação** (10h)
   - Página de login (email/password)
   - AuthContext com listener do Firebase
   - Extração de custom claims (tenant_id, role)
   - Hook useAuth
   - Middleware (redirect para /login se não autenticado)
   - Funcionalidade de logout
   - Persistência de sessão

3. **Layout do Dashboard** (8h)
   - Navegação sidebar (responsiva, colapsável)
   - DashboardHeader (menu de usuário, notificações)
   - Layout do dashboard (sidebar + conteúdo principal)
   - Notificações badge (parcerias pendentes)
   - Navegação responsiva (menu hamburger em mobile)

4. **Gestão de Imóveis** (14h)
   - Página Meus Imóveis (tabela com filtros)
   - PropertyForm (criar/editar com 30+ campos)
   - Seletor de visibilidade (private/network/marketplace/public)
   - Campo de comissão co-corretor (condicional)
   - Uploader de fotos (drag & drop)
   - Validação de formulário (react-hook-form + zod)
   - PropertyTable (ordenável, filtrável)
   - PropertyFilters
   - VisibilityBadge
   - BrokerRoleManager (gerenciar co-corretores)

5. **Busca Interna & Parcerias** (12h)
   - Página de busca interna de imóveis (network + marketplace)
   - PropertySearchCard (com botão "Tenho cliente para este imóvel")
   - Página de solicitações de parceria (UI de aprovar/rejeitar)
   - PartnershipRequestCard
   - PartnershipTable (parcerias ativas)
   - ApprovalDialog
   - Notificações (toast ao aprovar/rejeitar)
   - Contador de badge na sidebar

6. **Gestão de Leads** (10h)
   - Página de leads (tabela filtrada por PropertyBrokerRole)
   - LeadTable (ordenável por data, status, canal)
   - LeadFilters (imóvel, status, canal, range de data)
   - Modal LeadDetails
   - LeadStatusBadge
   - Atualizar status de lead (contacted, qualified, lost)

7. **UI de Importação** (10h)
   - Página de importação (uploader drag & drop)
   - Componente ImportUploader (XML/XLS)
   - Tabela de histórico de importação
   - Componente ImportStatus (progresso do batch)
   - Componente ImportErrors (lista de erros)
   - Retry de importações falhadas

8. **Configurações/Branding** (10h)
   - Página de configurações
   - BrandingForm (nome comercial, tagline, cores)
   - LogoUploader (drag & drop, preview)
   - FaviconUploader
   - ColorPicker (primary, secondary, accent)
   - Preview de mudanças de branding
   - Ações de salvar/cancelar

**Critérios de Validação**:
- ✅ Broker pode fazer login com Firebase Auth
- ✅ Rotas protegidas (redirect para /login se não autenticado)
- ✅ Broker pode criar/editar imóveis com controle de visibilidade
- ✅ Broker pode buscar imóveis internos (network/marketplace)
- ✅ Broker pode clicar "Tenho cliente" (solicitação de parceria criada)
- ✅ Broker originador pode aprovar/rejeitar parcerias
- ✅ Badge mostra contagem de parcerias pendentes
- ✅ Broker vê leads APENAS de imóveis onde tem PropertyBrokerRole
- ✅ Broker pode fazer upload de arquivos XML/XLS
- ✅ Status de importação mostra progresso, erros
- ✅ Broker pode customizar branding (logo, cores)

**Arquivos**: ~60 arquivos (~5000 linhas)

---

### Fase 5: APIs Backend - Properties, Listings, Leads (40-50h)

**Objetivo**: Completar implementação de API backend para gestão de imóveis, listings e leads

**Entregas**:
1. **Endpoints de Property** (12h)
   - GET /api/v1/tenants/{tenantId}/properties (list com paginação)
   - POST /api/v1/tenants/{tenantId}/properties (create)
   - GET /api/v1/tenants/{tenantId}/properties/{propertyId} (get single)
   - PATCH /api/v1/tenants/{tenantId}/properties/{propertyId} (update)
   - DELETE /api/v1/tenants/{tenantId}/properties/{propertyId} (soft delete)
   - PATCH /api/v1/tenants/{tenantId}/properties/{propertyId}/primary-broker
   - GET /api/v1/tenants/{tenantId}/properties/internal-search (visibility: network/marketplace)
   - PropertyService (lógica de negócio)
   - PropertyHandler

2. **Endpoints de Listing** (8h)
   - GET /api/v1/tenants/{tenantId}/listings
   - POST /api/v1/tenants/{tenantId}/listings
   - PATCH /api/v1/tenants/{tenantId}/listings/{listingId}
   - PATCH /api/v1/tenants/{tenantId}/properties/{propertyId}/canonical-listing
   - ListingService (lógica de atribuição canônica)
   - ListingHandler

3. **Endpoints de Lead** (12h)
   - POST /api/v1/properties/{propertyId}/leads/whatsapp (público, cria lead)
   - POST /api/v1/properties/{propertyId}/leads/form (público)
   - GET /api/v1/tenants/{tenantId}/brokers/{brokerId}/leads (filtrado por PropertyBrokerRole)
   - GET /api/v1/tenants/{tenantId}/leads/{leadId}
   - PATCH /api/v1/tenants/{tenantId}/leads/{leadId} (update status)
   - LeadService (algoritmo de roteamento para corretor primário)
   - LeadHandler
   - Serviço de notificação (email + notificações dashboard)

4. **Endpoints de Parceria** (10h)
   - POST /api/v1/tenants/{tenantId}/properties/{propertyId}/brokers/interest
   - PATCH /api/v1/tenants/{tenantId}/properties/{propertyId}/brokers/{brokerId}/approve
   - PATCH /api/v1/tenants/{tenantId}/properties/{propertyId}/brokers/{brokerId}/reject
   - GET /api/v1/tenants/{tenantId}/brokers/{brokerId}/partnership-requests
   - GET /api/v1/tenants/{tenantId}/brokers/{brokerId}/partnerships
   - PartnershipService
   - PartnershipHandler

5. **Confirmação de Status** (8h)
   - PATCH /api/v1/tenants/{tenantId}/properties/{propertyId}/confirmations
   - Confirmação passiva do owner (baseado em link, sem login)
   - Validade temporal (status_confirmed_at, price_confirmed_at)
   - Detecção de imóvel desatualizado
   - ConfirmationService
   - ConfirmationHandler

**Critérios de Validação**:
- ✅ CRUD de Property funciona com isolamento de tenant
- ✅ Listings criados e atribuição canônica funciona
- ✅ Criação de Lead (WhatsApp + form) funciona
- ✅ Algoritmo de roteamento para corretor primário correto
- ✅ Notificações enviadas (email + dashboard)
- ✅ Fluxo de parceria funciona (solicitação → aprovar/rejeitar)
- ✅ Confirmação de status funciona (link do owner)
- ✅ Todos os endpoints respeitam isolamento de tenant

**Arquivos**: ~20 arquivos (~2000 linhas)

---

### Fase 6: Busca Pública & SEO (30-40h)

**Objetivo**: Implementar busca avançada com índices Firestore e otimização SEO

**Entregas**:
1. **Índices Firestore** (4h)
   - Criar `firestore.indexes.json`
   - Índices compostos para queries de busca:
     - `[visibility_public, city, property_type, price_amount]`
     - `[visibility_public, city, neighborhood, bedrooms]`
     - `[visibility_public, property_type, created_at]`
   - Deploy de índices para Firestore

2. **Backend de Busca** (10h)
   - GET /api/v1/properties/search (endpoint público)
   - Struct SearchFilters (9 parâmetros de filtro)
   - Query builder (queries dinâmicas Firestore)
   - Paginação (limit, offset)
   - Ordenação (recent, price_asc, price_desc)
   - SearchService
   - SearchHandler

3. **Aprimoramento Frontend de Busca** (8h)
   - Refinar SearchFilters (todos os 9 filtros)
   - Adicionar chips de filtro (mostrar filtros ativos)
   - Adicionar botão "Limpar filtros"
   - Sync de URL params (habilitar links diretos)
   - Estados de loading (skeleton)
   - Estados de erro
   - Empty state

4. **Otimização SEO** (8h)
   - Geração de sitemap (`sitemap.xml`) - dinâmico, inclui todos os imóveis públicos
   - Robots.txt
   - JSON-LD dinâmico para páginas de imóveis
   - Componente Breadcrumbs
   - Tags de canonical URL
   - Alt tags de imagem
   - Otimização Core Web Vitals (LCP, CLS, FID)

**Critérios de Validação**:
- ✅ Busca filtra imóveis corretamente (todos os 9 filtros)
- ✅ Ordenação funciona (recent, price asc/desc)
- ✅ Paginação funciona
- ✅ Sync de URL params (links de busca compartilháveis)
- ✅ Queries Firestore usam índices (sem warnings)
- ✅ Performance < 1s para busca
- ✅ Sitemap.xml gerado
- ✅ Robots.txt presente
- ✅ JSON-LD presente em páginas de imóveis
- ✅ Score Lighthouse SEO > 90

**Arquivos**: ~10 arquivos (~1000 linhas)

---

### Fase 7: Testes & Validação (20-30h)

**Objetivo**: Testes abrangentes e validação de conformidade

**Entregas**:
1. **Testes Unitários Backend** (8h)
   - Testes de repository (mocks Firestore)
   - Testes de service (lógica de negócio)
   - Testes de middleware (auth, isolamento de tenant)
   - Testes de deduplicação
   - Cobertura de testes > 70%

2. **Testes de Integração** (6h)
   - Testes de endpoint API (usar emulador Firestore)
   - Testes de fluxo de importação (XML → Property → Listing)
   - Testes de criação de lead (WhatsApp + form)
   - Testes de fluxo de parceria

3. **Testes Frontend** (6h)
   - Testes de componente (React Testing Library)
   - Testes E2E (Playwright ou Cypress)
   - Teste de SSR da página de detalhe do imóvel
   - Teste de fluxo de busca
   - Teste de fluxo de criação de lead

4. **Validação de Conformidade** (10h)
   - Checklist PROMPT 03 (validação de governança)
   - Checklist PROMPT 05 (auditoria final)
   - Revisão de conformidade LGPD
   - Testes de Firestore Security Rules
   - Testes de performance (Lighthouse)
   - Testes de acessibilidade (WCAG AA)
   - Testes cross-browser

**Critérios de Validação**:
- ✅ Todos os testes unitários passam
- ✅ Todos os testes de integração passam
- ✅ Testes E2E passam em fluxos principais
- ✅ Checklist PROMPT 03 100% OK
- ✅ Checklist PROMPT 05 100% OK
- ✅ Nenhuma violação crítica detectada
- ✅ Scores Lighthouse: Performance > 85, SEO > 90, Acessibilidade > 90

**Arquivos**: ~20 arquivos de teste (~1500 linhas)

---

### Fase 8: Deploy & Setup de Produção (20-30h)

**Objetivo**: Deploy para produção (GCP + Vercel)

**Entregas**:
1. **Deploy Backend** (8h)
   - Dockerfile para backend
   - Config Cloud Build (`cloudbuild.yaml`)
   - Deploy Cloud Run
   - Setup de variáveis de ambiente (secrets)
   - Setup bucket Cloud Storage (acesso público)
   - Deploy de índices Firestore
   - Deploy de Firestore Security Rules
   - Setup de service account

2. **Deploy Frontend** (6h)
   - Projeto Vercel para frontend-public
   - Projeto Vercel para frontend-admin
   - Setup de variáveis de ambiente
   - Domínios customizados (opcional)
   - Deploy para produção

3. **Monitoramento & Logging** (6h)
   - Setup Cloud Logging
   - Rastreamento de erros (Sentry ou GCP Error Reporting)
   - Vercel Analytics
   - Monitoramento de uptime
   - Monitoramento de performance

**Critérios de Validação**:
- ✅ Backend deployed no Cloud Run
- ✅ Frontend público deployed no Vercel
- ✅ Frontend admin deployed no Vercel
- ✅ Todos os serviços acessíveis
- ✅ Certificados SSL válidos
- ✅ Logging funcionando
- ✅ Dashboards de monitoramento configurados

**Arquivos**: ~5 arquivos de config (~300 linhas)

---

## 📈 Resumo de Estimativa de Esforço

| Fase | Descrição | Horas | Arquivos de Código | Linhas de Código |
|------|-----------|-------|-------------------|------------------|
| 1 | Foundation & Auth | 40-50 | 25 | 3.000 |
| 2 | Sistema de Importação | 50-60 | 15 | 2.500 |
| 3 | Frontend Público | 50-60 | 40 | 4.000 |
| 4 | Frontend Admin | 60-70 | 60 | 5.000 |
| 5 | APIs Backend | 40-50 | 20 | 2.000 |
| 6 | Busca & SEO | 30-40 | 10 | 1.000 |
| 7 | Testes | 20-30 | 20 | 1.500 |
| 8 | Deploy | 20-30 | 5 | 300 |
| **TOTAL** | **MVP Completo** | **310-390** | **195** | **19.300** |

**Timeline Estimado**:
- **1 Dev Full-Stack Sênior**: 10-12 semanas (2,5-3 meses)
- **2 Devs** (1 backend, 1 frontend): 6-8 semanas (1,5-2 meses)
- **3 Devs** (1 backend, 2 frontend): 4-6 semanas (1-1,5 meses)

---

## 🔄 Dependências Críticas

### Dependências Sequenciais (Deve Ser Construído Primeiro)

```
Fase 1 (Foundation + Auth) → BLOQUEIA TODAS AS OUTRAS FASES
    ↓
Fase 2 (Import) → Requer models da Fase 1
    ↓
Fase 5 (Backend APIs) → Requer models da Fase 1
    ↓
Fase 3 (Frontend Público) → Requer endpoints da Fase 5
    ↓
Fase 4 (Frontend Admin) → Requer endpoints da Fase 5
    ↓
Fase 6 (Busca & SEO) → Requer Fase 3 + Fase 5
    ↓
Fase 7 (Testes) → Requer todas as features
    ↓
Fase 8 (Deploy) → Requer todas as features
```

### Oportunidades de Trabalho Paralelo

Após Fase 1 completar:
- Fase 2 (Import) e Fase 5 (Backend APIs) podem trabalhar em paralelo (serviços diferentes)

Após Fase 5 completar:
- Fase 3 (Frontend Público) e Fase 4 (Frontend Admin) podem trabalhar em paralelo (projetos diferentes)

---

## ⚠️ Áreas de Risco

### Risco Alto

1. **Índices Compostos Firestore** (Fase 6)
   - **Risco**: Índices faltantes causam falhas de query em runtime
   - **Mitigação**: Deploy de `firestore.indexes.json` cedo, testar queries com dados de exemplo
   - **Esforço se perdido**: 4-8 horas para debug + redeploy

2. **Performance de Processamento de Imagens** (Fase 2)
   - **Risco**: Download + conversão de 100+ imagens por batch demora muito, timeouts
   - **Mitigação**: Processamento concorrente (goroutines), limite de 10 paralelas, handling de timeout
   - **Esforço se perdido**: 8-12 horas para refatorar para processamento assíncrono

3. **Vazamento de Dados Multi-Tenancy** (Fase 1)
   - **Risco**: Validação de tenant_id faltando em query → vazamento de dados entre tenants
   - **Mitigação**: Middleware SEMPRE extrai tenant_id, todas as queries DEVEM filtrar por tenant_id, Security Rules forçam isolamento
   - **Esforço se perdido**: CRÍTICO - auditoria de segurança completa + refatoração (40+ horas)

4. **Fluxo de Lead WhatsApp** (Fase 5)
   - **Risco**: Lead não criado antes do redirect (rastreamento perdido)
   - **Mitigação**: Frontend aguarda resposta da API antes do redirect, estado de loading, handling de timeout
   - **Esforço se perdido**: 6-10 horas para refatorar fluxo frontend

### Risco Médio

5. **Performance SSR** (Fase 3)
   - **Risco**: Páginas de detalhe de imóvel lentas para renderizar (> 3s) devido a múltiplas chamadas API
   - **Mitigação**: Endpoint único de backend retorna Property + Listing + Photos em uma chamada, usar ISR (Incremental Static Regeneration)
   - **Esforço se perdido**: 8-12 horas para otimizar queries backend + caching frontend

6. **Custom Claims Firebase Auth** (Fase 1)
   - **Risco**: Custom claims não definidos corretamente (tenant_id faltando) → autorização quebrada
   - **Mitigação**: Teste unitário de lógica de claims, validar claims em middleware, mensagens de erro claras
   - **Esforço se perdido**: 4-8 horas para debug + fix

7. **Flexibilidade do Parser XLS** (Fase 2)
   - **Risco**: Nomes de colunas no XLS não batem com formato esperado → importação falha
   - **Mitigação**: Auto-detecção de nomes de colunas (fuzzy matching), mensagens de erro detalhadas, permitir mapeamento manual (futuro)
   - **Esforço se perdido**: 6-10 horas para refatorar parser

### Risco Baixo

8. **Responsividade Mobile** (Fase 3, 4)
   - **Risco**: UI quebra em dispositivos mobile
   - **Mitigação**: Testar em 6 dispositivos (checklist PROMPT 04), usar design mobile-first
   - **Esforço se perdido**: 10-15 horas para corrigir problemas de layout

9. **Validação de Consentimento LGPD** (Fase 3, 5)
   - **Risco**: Formulários não coletam consentimento → não-conformidade legal
   - **Mitigação**: Checkbox de consentimento LGPD em todos os formulários, backend valida consent_given = true
   - **Esforço se perdido**: 4-6 horas para adicionar checkboxes + validação

---

## 🎯 Ordem de Implementação Recomendada

### Opção 1: Waterfall (Sequencial)
**Melhor para**: Dev solo ou time pequeno com paralelização limitada

```
Semana 1-2:   Fase 1 (Foundation + Auth)
Semana 3-4:   Fase 2 (Import)
Semana 5-6:   Fase 5 (Backend APIs)
Semana 7-8:   Fase 3 (Frontend Público)
Semana 9-10:  Fase 4 (Frontend Admin)
Semana 11:    Fase 6 (Busca & SEO)
Semana 12:    Fase 7 (Testes) + Fase 8 (Deploy)
```

### Opção 2: Paralela (Recomendado para 2+ Devs)
**Melhor para**: Time com especialistas backend + frontend

```
Semana 1-2:   Fase 1 (Foundation + Auth) - TIME TODO
Semana 3-4:   Fase 2 (Import) + Fase 5 (Backend APIs) - PARALELO
Semana 5-6:   Fase 3 (Frontend Público) + Fase 4 (Frontend Admin) - PARALELO
Semana 7:     Fase 6 (Busca & SEO) - Colaboração Backend + Frontend
Semana 8:     Fase 7 (Testes) + Fase 8 (Deploy) - TIME TODO
```

### Opção 3: MVP-First (Mais Rápido para Demo)
**Melhor para**: Precisa mostrar demo funcionando rapidamente

```
Semana 1-2:   Fase 1 (Foundation + Auth)
Semana 3:     Fase 5 (Backend APIs) - CRUD de Property apenas
Semana 4:     Fase 3 (Frontend Público) - Página de detalhe apenas
Semana 5:     Fase 2 (Import) - Import manual via API
Semana 6:     Fase 4 (Frontend Admin) - Lista básica de imóveis
[DEMO PRONTO]
Semana 7-8:   Completar features Fase 3, 4, 5
Semana 9:     Fase 6 (Busca & SEO)
Semana 10:    Fase 7 (Testes) + Fase 8 (Deploy)
```

---

## 📁 Arquivos Críticos para Começar Implementação

Baseado na análise abrangente, aqui estão os 5 arquivos mais críticos para começar:

1. **`backend/cmd/api/main.go`**
   Ponto de entrada para todo o backend, configura servidor, inicializa Firebase, registra rotas. Sem isso, nada roda.

2. **`backend/internal/models/property.go`**
   Define o model Property core com 30+ campos. É a fundação de todo o sistema (princípio de Property Uniqueness). Todos os repositories, services e handlers dependem disso.

3. **`backend/internal/middleware/auth_middleware.go`**
   Valida tokens JWT Firebase, extrai claims de tenant_id e role, força isolamento de tenant. Crítico para segurança e multi-tenancy.

4. **`prompts/01_foundation_mvp.txt`**
   Especificação completa para Fase 1, inclui todas as definições de model, padrões de repository e contratos de endpoint. É o blueprint de implementação para a fundação.

5. **`prompts/09_autenticacao_multitenancy.txt`**
   Especificação completa para autenticação e multi-tenancy, inclui setup Firebase, custom claims, padrões de isolamento de tenant e security rules. Essencial para começar corretamente.

---

## 📌 Próximos Passos Imediatos

Quando estiver pronto para começar a implementação:

1. **Criar estrutura de diretórios**:
   ```bash
   mkdir -p backend/{cmd/api,internal/{models,repositories,services,handlers,middleware,adapters,utils},pkg/firebase}
   mkdir -p frontend-public/{app,components,lib,hooks,types,public}
   mkdir -p frontend-admin/{app,components,contexts,hooks,lib}
   ```

2. **Inicializar projeto Go**:
   ```bash
   cd backend
   go mod init github.com/altatechsystems/ecosistema-imob-backend
   ```

3. **Criar projeto Firebase**:
   - Acessar console.firebase.google.com
   - Criar novo projeto
   - Habilitar Authentication (Email/Password)
   - Criar Firestore database (modo nativo)
   - Criar bucket Cloud Storage
   - Baixar service account JSON

4. **Seguir Fase 1 do plano**:
   - Começar com `backend/cmd/api/main.go`
   - Implementar models (8 arquivos)
   - Implementar auth middleware
   - Criar endpoints básicos

---

**Documento gerado em**: 2025-12-21
**Por**: Claude Code (Análise de Implementação)
**Agent ID para retomar**: adef293
