# 🏢 Ecossistema Imobiliário - MVP

Plataforma SaaS multi-tenant para ecossistema imobiliário com governança de ativo único, co-corretagem e gestão inteligente de leads.

---

## 📋 Visão Geral

Este projeto implementa um **MVP (Minimum Viable Product)** de uma plataforma imobiliária moderna que resolve problemas críticos do mercado brasileiro:

- ✅ **Imóvel Único**: Elimina duplicação de propriedades
- ✅ **Co-corretagem Estruturada**: Captador, vendedor e co-corretores claramente definidos
- ✅ **Multi-tenancy**: Múltiplas imobiliárias isoladas desde o MVP
- ✅ **Proprietário Passivo**: Confirmação de status/preço sem login
- ✅ **WhatsApp-First**: Lead gerado ANTES do redirect (rastreabilidade total)
- ✅ **SEO 100%**: Score superior a ZAP (75%) e VivaReal (80%), sitemap dinâmico, Core Web Vitals otimizados
- ✅ **Whitelabel**: Marca própria para imobiliárias (ROI 26x/ano, payback 14 dias)
- ✅ **Blockchain-Ready**: Tokenização factível em MVP+2 (3 modelos comprovados de mercado)

---

## 🏗️ Arquitetura

### Monorepo com 3 Projetos Separados:

```
ecosistema-imob/
├── backend/              # API Go + Firestore
├── frontend-public/      # Next.js 14 (Público - Usuários Finais)
├── frontend-admin/       # Next.js 14 (Admin - Corretores/Imobiliárias)
├── docs/                 # Documentação
├── prompts/              # Prompts de implementação
└── data/                 # Dados de exemplo (XML/XLS)
```

**Justificativa da Separação:**
- ✅ **Segurança**: Admin isolado, sem código sensível no público
- ✅ **Performance**: Bundle público otimizado (~150KB) para SEO
- ✅ **SEO**: Frontend público 100% SSR/SSG sem rotas protegidas
- ✅ **Deploy**: Independente (mudança no admin não afeta o público)
- ✅ **Manutenção**: Código focado, menos condicionais

---

## 🚀 Stack Tecnológica

### Backend
- **Linguagem**: Go (Golang) 1.21+
- **Framework**: Gin
- **Banco**: Google Cloud Firestore
- **Autenticação**: Firebase Authentication
- **Storage**: Google Cloud Storage (GCS)
- **Deploy**: Google Cloud Run
- **URL**: `api.example.com`

### Frontend Público (Usuários Finais)
- **Framework**: Next.js 14 (App Router)
- **Linguagem**: TypeScript 5+
- **UI**: shadcn/ui + Tailwind CSS
- **Estado**: React Query + Zustand
- **Autenticação**: ❌ SEM autenticação (público)
- **Deploy**: Vercel
- **URL**: `www.example.com`
- **Rotas**: `/`, `/buscar`, `/imovel/[slug]`

### Frontend Admin (Corretores/Imobiliárias)
- **Framework**: Next.js 14 (App Router)
- **Linguagem**: TypeScript 5+
- **UI**: shadcn/ui + Tailwind CSS
- **Estado**: React Query + Zustand
- **Autenticação**: ✅ Firebase Auth SDK (obrigatório)
- **Deploy**: Vercel
- **URL**: `app.example.com` ou `admin.example.com`
- **Rotas**: `/login`, `/imoveis`, `/leads`, `/importacao`

### Infraestrutura
- **Cloud Provider**: Google Cloud Platform (GCP)
- **CDN**: Cloud CDN
- **Monitoramento**: Cloud Logging + Vercel Analytics

---

## 📂 Estrutura do Projeto

### `/backend` - API Go
```
backend/
├── cmd/api/              # Entry point
├── internal/
│   ├── models/           # Modelos de domínio
│   ├── repositories/     # Acesso ao Firestore
│   ├── services/         # Lógica de negócio
│   ├── handlers/         # HTTP handlers (Gin)
│   ├── middleware/       # Auth, tenant isolation
│   └── adapters/         # Importação (Union XML/XLS)
├── pkg/firebase/         # Firebase Admin SDK
└── docs/decisions/       # ADRs (Architecture Decision Records)
```

### `/frontend-public` - Next.js Público
```
frontend-public/
├── app/
│   ├── page.tsx          # Homepage
│   ├── buscar/           # Busca de imóveis
│   └── imovel/[slug]/    # Detalhes do imóvel (SSR)
├── components/
│   ├── ui/               # shadcn/ui components
│   ├── property/         # PropertyCard, PropertyGallery
│   ├── search/           # SearchFilters, SearchResults
│   └── shared/           # Header, Footer
├── lib/
│   └── api.ts            # API client (backend)
└── hooks/                # React Query hooks
```

### `/frontend-admin` - Next.js Admin
```
frontend-admin/
├── app/
│   ├── login/            # Login Firebase Auth
│   └── (dashboard)/      # Rotas protegidas
│       ├── imoveis/      # Gestão de imóveis
│       ├── leads/        # Gestão de leads
│       └── importacao/   # Upload XML/XLS
├── components/
│   ├── ui/               # shadcn/ui components
│   ├── dashboard/        # Sidebar, DashboardHeader
│   ├── properties/       # PropertyForm, PropertyTable
│   ├── leads/            # LeadTable, LeadDetails
│   └── import/           # ImportUploader, ImportErrors
├── contexts/
│   └── AuthContext.tsx   # Firebase Auth provider
├── middleware.ts         # Route protection
└── lib/
    ├── api.ts            # API client (backend)
    └── firebase.ts       # Firebase Auth config
```

### `/docs` - Documentação
- **[AI_DEV_DIRECTIVE.md](AI_DEV_DIRECTIVE.md)**: Contrato supremo do projeto
- **[VALIDACAO_FINAL.md](VALIDACAO_FINAL.md)**: Checklist e ordem de implementação
- **[ATUALIZACOES_REALIZADAS.md](ATUALIZACOES_REALIZADAS.md)**: Documento executivo

### `/prompts` - Guias de Implementação
Sequência de implementação:
1. `09_autenticacao_multitenancy.txt` + `01_foundation_mvp.txt`
2. `02_import_deduplication.txt`
3. `03_audit_governance.txt`
4. `04_frontend_public_mvp.txt` (Frontend Público)
5. `04b_frontend_admin_mvp.txt` (Frontend Admin)
6. `10_busca_publica.txt`
7. `07_whatsapp_flow.txt`
8. `08_Property Status Confirmation.txt`
9. `06_distribuição_multicanal.txt`
10. `05_final_audit.txt`

---

## 🎯 Conceitos Fundamentais

### 1. Imóvel Único (Property)
Cada imóvel físico existe **uma única vez** no sistema, mesmo que anunciado por múltiplos corretores.

### 2. Canonical Listing
Cada Property possui **exatamente um** listing canônico exibido publicamente. Outros listings existem internamente mas não são expostos.

### 3. Co-corretagem (PropertyBrokerRole)
- **Captador** (`originating_broker`): corretor que originou o imóvel (único)
- **Vendedor** (`listing_broker`): corretor responsável por um Listing (múltiplos possíveis)
- **Co-corretor** (`co_broker`): corretor adicional na negociação (N possíveis)

### 4. Proprietário Passivo (Owner)
No MVP, o proprietário:
- ✅ Pode confirmar status/preço via link (sem login)
- ❌ NÃO tem login
- ❌ NÃO vê leads
- ❌ NÃO negocia

### 5. Multi-tenancy
Estrutura Firestore:
```
/tenants/{tenantId}/
  properties/{propertyId}
  listings/{listingId}
  brokers/{brokerId}
  leads/{leadId}
  activity_logs/{logId}
```

---

## 🔑 Regras de Negócio Críticas

### WhatsApp Flow
**OBRIGATÓRIO**: Todo clique em "Falar no WhatsApp" DEVE:
1. Criar Lead via backend (`POST /api/v1/properties/:propertyId/leads/whatsapp`)
2. Somente DEPOIS redirecionar para WhatsApp
3. Mensagem pré-preenchida com `#leadId`

### Lead
- Lead pertence ao **Property** (NUNCA diretamente ao corretor)
- Criado via página pública (formulário ou WhatsApp)
- Sem cadastro do cliente final

### Status e Preço
- Disponibilidade e preço pertencem ao **Property**
- Corretores apenas **confirmam** informações
- Validade temporal: `status_confirmed_at`, `price_confirmed_at`
- Imóveis "stale" podem ser ocultados automaticamente

---

## 🖼️ Processamento de Imagens

Pipeline automático na importação:
1. Download URL externa → GCS (temp)
2. Conversão para WebP (3 tamanhos):
   - `thumb_400.webp` (400x300)
   - `medium_800.webp` (800x600)
   - `large_1600.webp` (1600x1200)
3. Upload GCS (público)
4. Excluir original

---

## 🔍 SEO

### URLs Amigáveis
```
/imovel/{slug}
Exemplo: /imovel/apartamento-sao-paulo-jardim-europa-ap00335
```

### Meta Tags Dinâmicas
- Title: `{Tipo} {Característica} em {Bairro}, {Cidade} | {Imobiliária}`
- OpenGraph completo
- JSON-LD (schema.org/RealEstateListing)
- SSR obrigatório em `/imovel/[slug]`

---

## 🔐 Autenticação

### Firebase Authentication
- Email/senha (MVP)
- Custom Claims: `{tenant_id: "abc123", role: "admin" | "broker"}`
- JWT gerenciado automaticamente

### Middlewares Go
- `AuthMiddleware`: valida Firebase token
- `TenantIsolationMiddleware`: valida tenant_id
- `AdminOnlyMiddleware`: restringe admin

---

## 📊 Auditoria

### ActivityLog
Todos os eventos críticos são registrados:
- `property_created`, `listing_created`, `canonical_assigned`
- `lead_created_whatsapp`, `lead_created_form`
- `owner_confirmed_status`, `owner_confirmed_price`
- `import_batch_started`, `import_batch_completed`

Campos obrigatórios:
- `event_id` (determinístico)
- `event_hash` (SHA256)
- `request_id` (UUID v4)
- `tenant_id`, `actor_type`, `event_type`, `timestamp`

Retenção:
- Eventos críticos: **permanente**
- Eventos operacionais: **90 dias**

---

## 🎨 Design (Frontend)

### Inspiração
Layout e hierarquia inspirados na **[Zillow](https://www.zillow.com/)**, adaptado ao mercado brasileiro:
- WhatsApp-first
- Mobile-first
- CTAs claros

### UI Components
- **shadcn/ui** + Tailwind CSS
- Componentes reutilizáveis e acessíveis
- Tema customizável por tenant

---

## 🚦 Ordem de Implementação

Ver [VALIDACAO_FINAL.md](VALIDACAO_FINAL.md) para ordem detalhada.

### Fase 1: Fundação (Backend + Auth)
- PROMPT 09 + PROMPT 01

### Fase 2: Importação
- PROMPT 02

### Fase 3: Frontend
- PROMPT 04

### Fase 4: Busca
- PROMPT 10

### Fase 5: WhatsApp + Confirmação
- PROMPT 07 + PROMPT 08

---

## 📝 Glossário

| Português | Inglês (Código) | Definição |
|-----------|----------------|-----------|
| Imóvel | Property | Ativo físico único |
| Anúncio | Listing | Versão do anúncio criada por corretor |
| Proprietário | Owner | Titular legal do imóvel |
| Corretor | Broker | Profissional/empresa com CRECI |
| Captador | Originating Broker | Corretor que captou o imóvel |
| Vendedor | Listing Broker | Corretor responsável por Listing |
| Co-corretor | Co-Broker | Corretor adicional na negociação |
| Imobiliária | Tenant | Empresa que usa a plataforma |
| Lead | Lead | Manifestação de interesse |

---

## 📜 Documentos Principais

### Documentação de Negócio
1. **[PLANO_DE_NEGOCIOS.md](PLANO_DE_NEGOCIOS.md)**: Plano de negócios completo v1.6 (SEO 100% + Whitelabel + Lançamentos + Serviços Inovadores + Tokenização)
2. **[ANALISE_CONFORMIDADE_CRECI_COFECI.md](ANALISE_CONFORMIDADE_CRECI_COFECI.md)**: Análise regulatória detalhada

### Documentação Técnica
3. **[PLANO_DE_IMPLEMENTACAO.md](PLANO_DE_IMPLEMENTACAO.md)**: Roadmap técnico (310-390h, 8 fases)
4. **[AI_DEV_DIRECTIVE.md](AI_DEV_DIRECTIVE.md)**: Contrato supremo do projeto
5. **[VALIDACAO_FINAL.md](VALIDACAO_FINAL.md)**: Checklist e ordem de implementação
6. **[ATUALIZACOES_REALIZADAS.md](ATUALIZACOES_REALIZADAS.md)**: Documento executivo de atualizações

---

## ⚠️ Escopo Proibido no MVP

- ❌ Pagamentos
- ❌ Split financeiro automático
- ❌ Área do proprietário (login)
- ❌ Exclusividade automática
- ❌ WhatsApp Business API
- ❌ Chat complexo
- ❌ Blockchain ativa

---

## ✅ Status do Projeto

**✅ VALIDADO E PRONTO PARA IMPLEMENTAÇÃO**

- Governança de negócio: ✅ Sólida
- Stack técnica: ✅ Definida
- Multi-tenancy: ✅ Desde MVP
- Co-corretagem: ✅ Bem modelada
- Processamento de imagens: ✅ Definido
- SEO: ✅ Otimizado
- Documentação: ✅ Completa

**Risco de refatoração: BAIXO** ✅

---

## 📞 Suporte

Para dúvidas sobre implementação, consultar:
1. [AI_DEV_DIRECTIVE.md](AI_DEV_DIRECTIVE.md) - Regras de negócio
2. `/prompts/*.txt` - Guias detalhados de implementação
3. [VALIDACAO_FINAL.md](VALIDACAO_FINAL.md) - Checklist completo

---

**Versão**: 1.6
**Data**: 2025-12-21
**Status**: 🚀 Pronto para Implementação | SEO 100% | Whitelabel (26x ROI) | Lançamentos (27x ROI) | Serviços Inovadores (R$ 2.3M/ano) | Tokenização Factível
