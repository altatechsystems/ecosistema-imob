# AI_DEV_DIRECTIVE.md
## Diretiva Técnica e de Produto — MVP Ecossistema Imobiliário

## 1. Propósito do Documento
Este documento define as regras estruturais, limites de escopo e princípios fundamentais do MVP da plataforma de ecossistema imobiliário.
Ele é o contrato supremo do projeto.

## 2. Princípios Invioláveis do MVP
### 2.1 Imóvel Único
Cada imóvel físico existe uma única vez (Property). É proibida duplicação.

### 2.2 Separação Conceitual
- Property: ativo único
- Listing: anúncio por corretor
- Owner: proprietário (passivo no MVP)

### 2.3 Multi-tenancy Obrigatório
Desde o MVP, a plataforma deve suportar múltiplas imobiliárias (tenants) com isolamento completo de dados.
- Cada tenant possui corretores, imóveis e leads isolados
- Compartilhamento de imóveis entre tenants via co-corretagem explícita
- Sem compartilhamento acidental de dados

## 3. Proprietário (Owner)
- Titular legal do imóvel
- Passivo no MVP
- Sem login, telas ou leads

## 4. Co-corretagem
- Múltiplos corretores por imóvel
- Papéis distintos:
  - **Captador**: corretor que originou/captou o imóvel (único por Property)
  - **Vendedor (Listing Broker)**: corretor responsável por um Listing específico (pode haver múltiplos Listings)
  - **Co-corretor**: corretor adicional na negociação/venda (comum no mercado brasileiro)
- Um Property pode ter:
  - 1 captador (obrigatório)
  - N vendedores (1 por Listing)
  - N co-corretores (adicionados durante negociação)
- Sem split financeiro no MVP (apenas registro de comissão)

## 5. Canonical Listing
Cada Property possui exatamente um canonical_listing_id.
Somente o canonical listing é exibido publicamente.

## 6. Distribuição Multicanal (Evolução Pós-MVP)
O Property e seu canonical listing podem ser distribuídos por múltiplos canais
(portais, links diretos, redes sociais), sempre preservando:
- a unicidade do imóvel
- a exibição de um único anúncio público
- a associação de leads ao Property
Canais externos são tratados como meios de distribuição, não como cadastros independentes.

## 7. Leads
- Lead pertence ao Property
- Criado via página pública
- Sem cadastro do cliente final

## 8. WhatsApp como Canal de Atendimento (Regra Estrutural do MVP)

No MVP, o WhatsApp é tratado **exclusivamente como canal de comunicação**, e **não** como sistema de entrada de dados, origem primária do lead ou substituto do funil da plataforma.

### 8.1 Regra Fundamental

**Todo contato iniciado via WhatsApp deve gerar um Lead na plataforma antes do redirecionamento para o aplicativo.**

É **explicitamente proibido** no MVP:
- redirecionar o usuário para o WhatsApp sem registrar previamente o lead;
- permitir que o lead exista apenas na conversa externa;
- tratar o WhatsApp como origem primária do lead;
- associar o lead diretamente a um corretor, em vez de ao imóvel.

### 8.2 Fluxo Obrigatório no MVP

O fluxo correto e obrigatório é:

1. O usuário visualiza a página pública do imóvel.
2. Ao clicar em **“Falar no WhatsApp”**, a plataforma:
   - cria um **Lead** associado ao `property_id`;
   - registra `channel = whatsapp`;
   - registra dados de origem disponíveis (UTM, página, campanha).
3. Somente após a criação do Lead, o usuário é redirecionado ao WhatsApp, com uma **mensagem pré-preenchida**, contendo:
   - identificação do imóvel;
   - identificador do lead (ex.: `#L12345`).

A conversa ocorre normalmente no WhatsApp pessoal do corretor ou da imobiliária.

### 8.3 Limitações do MVP

No MVP é **proibido**:
- integração com WhatsApp Business API;
- automação de mensagens;
- bots ou fluxos automáticos;
- roteamento inteligente de atendentes;
- armazenamento de mensagens da conversa na plataforma.

### 8.4 Evolução Futura (Fora do Escopo do MVP)

Integrações com **WhatsApp Business API** são consideradas **evolução futura**, voltadas a cenários de maior volume, múltiplos atendentes e necessidade de histórico centralizado, e **não fazem parte do escopo do MVP**.

### 8.5 Justificativa

Essa abordagem garante:
- rastreabilidade completa dos leads;
- preservação do imóvel como ativo central do sistema;
- aderência ao comportamento do mercado imobiliário brasileiro;
- simplicidade técnica e baixo custo no MVP;
- preparação para evolução futura sem refatoração estrutural.

### 8.6 Checklist de Conformidade

- [ ] Todo clique em WhatsApp cria um Lead antes do redirect  
- [ ] Lead pertence ao Property  
- [ ] WhatsApp nunca é origem primária do dado  
- [ ] Não existe integração com WhatsApp Business no MVP  

## 9. Importação
- Arquitetura por adapters (source)
- MVP: apenas Union XML
- Normalização + deduplicação obrigatórias

## 10. Auditoria
- ActivityLog obrigatório
- Dois gates: pós-backend e pós-frontend

## 11. Frontend
- Público: /imovel/[propertyId], apenas canonical
- Privado: gestão de imóveis, leads e importação

## 12. Blockchain-ready
Estrutura preparada, sem blockchain ativa no MVP.

## 13. Escopo Proibido
Pagamentos, split financeiro, área do proprietário, exclusividade automática.

## 14. Stack Tecnológica do MVP

### 14.1 Backend
- **Linguagem**: Go (Golang) 1.21+
- **Framework Web**:
  - Gin (recomendado para MVP - leve, rápido, comunidade grande)
  - OU Fiber (se preferir performance extrema e API Express-like)
- **Banco de Dados**: Google Cloud Firestore (NoSQL nativo do Firebase)
  - Vantagens: serverless, auto-scaling, integração Firebase Auth, queries indexadas
  - Estrutura de coleções: `/tenants/{tenantId}/properties/{propertyId}`
- **ORM/ODM**:
  - Firebase Admin SDK for Go (oficial, tipo-seguro)
  - Camada de abstração customizada para queries complexas
- **Autenticação**: Firebase Authentication
  - Suporte nativo a email/senha, Google, links mágicos
  - JWT gerenciado automaticamente
  - Custom claims para multi-tenancy (tenant_id, role)
- **Storage**: Google Cloud Storage (GCS)
  - Vantagens: CDN integrado, versionamento, lifecycle policies
  - Estrutura: `gs://{bucket}/tenants/{tenantId}/properties/{propertyId}/photos/{photoId}.webp`
  - **NÃO usar Cloud Filestore** (POSIX filesystem, overkill para imagens)

### 14.2 Frontend
- **Framework**: Next.js 14+ (App Router)
- **Linguagem**: TypeScript 5+
- **UI Components**: shadcn/ui + Tailwind CSS
  - Justificativa: componentes reutilizáveis, acessíveis, customizáveis
- **Gerenciamento de Estado**:
  - React Query (TanStack Query) para server state
  - Zustand para client state (filtros, UI temporário)
- **Autenticação**: Firebase Auth SDK (client-side)
- **Imagens**: next/image com loader customizado para GCS

### 14.3 Infraestrutura
- **Hospedagem Frontend**: Vercel (deploy automático via GitHub)
- **Hospedagem Backend**: Google Cloud Run
  - Serverless, auto-scaling, paga por uso
  - Deploy via GitHub Actions → Artifact Registry → Cloud Run
- **CDN**: Cloud CDN (GCP) para imagens
- **Domínio/DNS**: Cloud DNS ou Vercel DNS
- **Monitoramento**:
  - Backend: Cloud Logging + Cloud Monitoring
  - Frontend: Vercel Analytics

### 14.4 Processamento de Imagens
- **Pipeline de Importação**:
  1. Download da URL externa (XML) → Cloud Storage (original)
  2. Processamento síncrono via `imaging/draw` (Go) ou Cloud Functions
  3. Gerar WebP em 3 tamanhos:
     - `thumb_400.webp` (400x300, listagens)
     - `medium_800.webp` (800x600, carrossel)
     - `large_1600.webp` (1600x1200, fullscreen)
  4. Excluir original após conversão (economia de storage)
- **Formato**: WebP (85% de qualidade, melhor compressão que JPEG)
- **Lazy Loading**: obrigatório no frontend (`loading="lazy"`)

### 14.5 Multi-tenancy
- **Estratégia**: Database-per-tenant via subcoleções Firestore
  - Estrutura: `/tenants/{tenantId}/properties/...`
  - Segurança via Firestore Security Rules
- **Identificação**:
  - Firebase Custom Claims: `{tenant_id: "abc123", role: "broker"}`
  - Middleware Go valida tenant_id em todas as requests
- **Isolamento**: queries SEMPRE filtram por tenant_id

## 15. URL e SEO

### 15.1 Estrutura de URLs Públicas
- Padrão: `/imovel/{slug}`
- Slug gerado automaticamente: `{tipo}-{cidade}-{bairro}-{ref}`
  - Exemplo: `/imovel/apartamento-sao-paulo-jardim-europa-ap00335`
  - Normalização: lowercase, sem acentos, hífens
- Fallback interno: campo `slug` (unique index no Firestore)
- Redirect 301 se slug mudar (manter SEO)

### 15.2 Meta Tags Obrigatórias
- Title dinâmico: `{Tipo} {Característica} em {Bairro}, {Cidade} | {TenantName}`
- Meta description: primeiros 155 chars da descrição do imóvel
- OpenGraph completo (og:image, og:title, og:description, og:url)
- JSON-LD schema.org/RealEstateListing

## 16. Glossário Técnico (Padronização)

**Sempre utilizar os termos em português nos documentos de negócio e em inglês no código:**

| Português (Docs) | Inglês (Código) | Definição |
|------------------|-----------------|-----------|
| Imóvel | Property | Ativo físico único (apartamento, casa, terreno) |
| Anúncio | Listing | Versão do anúncio criada por um corretor |
| Proprietário | Owner | Titular legal do imóvel (pessoa física/jurídica) |
| Corretor | Broker | Profissional/empresa com CRECI que opera imóveis |
| Captador | Originating Broker | Corretor que captou/originou o imóvel |
| Vendedor | Listing Broker | Corretor responsável por um Listing específico |
| Co-corretor | Co-Broker | Corretor adicional em uma negociação |
| Lead | Lead | Manifestação de interesse em um imóvel |
| Imobiliária | Tenant | Entidade/empresa que usa a plataforma (multi-tenancy) |
| Ator | Actor | Qualquer entidade que executa ação (User, System, Owner) |

**IMPORTANTE**: No código Go, usar sempre termos em inglês. Em prompts de implementação, usar português para clareza de negócio.

## 17. Auditoria e Logs

### 17.1 ActivityLog
- Coleção Firestore: `/tenants/{tenantId}/activity_logs/{logId}`
- Campos obrigatórios:
  - `event_id` (determinístico: hash de propertyId + action + timestamp_bucket_5min)
  - `event_hash` (SHA256 do payload normalizado)
  - `request_id` (UUID v4 por request HTTP)
  - `tenant_id` (isolamento)
  - `actor_type` (user | system | owner)
  - `actor_id`
  - `event_type` (ex: property_created, lead_created_whatsapp)
  - `timestamp`
  - `metadata` (map flexível)

### 17.2 Retenção de Logs
- **Eventos críticos** (retenção permanente):
  - property_created, listing_created, canonical_assigned
  - lead_created, owner_confirmed_status
- **Eventos operacionais** (90 dias):
  - import_batch_*, property_status_confirmed (rotineiro)
  - owner_confirmation_link_created
- **Purge**: Cloud Scheduler → Cloud Function mensal (deletar logs > 90 dias via batch)

## 18. Tratamento de Erros na Importação

### 18.1 Estratégia de Erro (MVP)
- **Erro de parsing XML/XLS**: pular registro + salvar em `import_errors` subcoleção
- **Campos obrigatórios faltando**:
  - Criar Property com valores default + flag `data_completeness: "incomplete"`
  - Owner placeholder se necessário
- **Erro de deduplicação**: marcar `possible_duplicate: true` + prosseguir
- **Storage**:
  - Subcoleção: `/tenants/{tenantId}/import_batches/{batchId}/errors/{errorId}`
  - Campos: `record_data`, `error_message`, `error_type`, `timestamp`
- **UI privada**: exibir erros do último batch + botão "Revisar erros"
- **Resolução manual**: corretor pode editar Property depois e remover flag `incomplete`

## 19. Estrutura de Diretórios do Projeto

O projeto será dividido em **TRÊS pastas separadas na raiz**:

```
ecosistema-imob/
├── backend/                    # Projeto Go (API)
│   ├── cmd/
│   │   └── api/
│   │       └── main.go
│   ├── internal/
│   │   ├── models/
│   │   ├── repositories/
│   │   ├── services/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   ├── adapters/
│   │   └── utils/
│   ├── pkg/
│   │   └── firebase/
│   ├── docs/
│   │   └── decisions/
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile
│   ├── cloudbuild.yaml
│   └── README.md
│
├── frontend-public/            # Projeto Next.js PÚBLICO (usuários finais)
│   ├── app/
│   │   ├── page.tsx           # Homepage
│   │   ├── buscar/
│   │   │   └── page.tsx       # Busca de imóveis
│   │   └── imovel/
│   │       └── [slug]/
│   │           └── page.tsx   # Detalhes do imóvel (SSR)
│   ├── components/
│   │   ├── ui/                # shadcn/ui components
│   │   ├── property/          # PropertyCard, PropertyGallery, etc.
│   │   ├── search/            # SearchFilters, SearchResults
│   │   └── shared/            # Header, Footer
│   ├── lib/
│   │   ├── api.ts             # API client (backend)
│   │   └── firebase.ts        # Firebase config (APENAS Analytics, SEM Auth)
│   ├── hooks/
│   │   └── use-properties.ts  # React Query hooks
│   ├── types/
│   ├── public/
│   ├── next.config.js
│   ├── tailwind.config.ts
│   ├── package.json
│   └── README.md
│
├── frontend-admin/             # Projeto Next.js ADMIN (corretores/imobiliárias)
│   ├── app/
│   │   ├── login/
│   │   │   └── page.tsx       # Login Firebase Auth
│   │   ├── (dashboard)/       # Grupo protegido
│   │   │   ├── layout.tsx     # Dashboard layout
│   │   │   ├── page.tsx       # Overview
│   │   │   ├── imoveis/
│   │   │   │   ├── page.tsx   # Lista de imóveis
│   │   │   │   └── [id]/
│   │   │   │       └── page.tsx # Editar imóvel
│   │   │   ├── leads/
│   │   │   │   └── page.tsx   # Gestão de leads
│   │   │   └── importacao/
│   │   │       └── page.tsx   # Upload XML/XLS
│   │   └── api/
│   ├── components/
│   │   ├── ui/                # shadcn/ui components
│   │   ├── dashboard/         # Sidebar, DashboardHeader
│   │   ├── properties/        # PropertyForm, PropertyTable
│   │   ├── leads/             # LeadTable, LeadDetails
│   │   └── import/            # ImportUploader, ImportErrors
│   ├── lib/
│   │   ├── api.ts             # API client (backend)
│   │   └── firebase.ts        # Firebase Auth config
│   ├── contexts/
│   │   └── AuthContext.tsx    # Firebase Auth provider
│   ├── hooks/
│   │   ├── use-auth.ts        # Auth hook
│   │   └── use-properties.ts  # React Query hooks
│   ├── middleware.ts           # Route protection
│   ├── types/
│   ├── public/
│   ├── next.config.js
│   ├── tailwind.config.ts
│   ├── package.json
│   └── README.md
│
├── docs/                       # Documentação geral
│   ├── AI_DEV_DIRECTIVE.md    # Contrato supremo
│   ├── VALIDACAO_FINAL.md
│   └── ATUALIZACOES_REALIZADAS.md
│
├── prompts/                    # Prompts de implementação
│   ├── 01_foundation_mvp.txt
│   ├── 02_import_deduplication.txt
│   ├── 03_audit_governance.txt
│   ├── 04_frontend_public_mvp.txt      # Frontend público (RENOMEAR)
│   ├── 04b_frontend_admin_mvp.txt      # Frontend admin (CRIAR NOVO)
│   ├── 05_final_audit.txt
│   ├── 06_distribuição_multicanal.txt
│   ├── 07_whatsapp_flow.txt
│   ├── 08_Property Status Confirmation.txt
│   ├── 09_autenticacao_multitenancy.txt
│   └── 10_busca_publica.txt
│
├── data/                       # Dados de exemplo
│   ├── 914802.xml
│   └── univen-imoveis_20-12-2025_18_12_15.xls
│
└── README.md                   # README principal do monorepo
```

### Justificativa da Separação em TRÊS Projetos:

**Backend (Go):**
- Deploy independente no Cloud Run
- Versionamento independente
- Testes independentes
- CI/CD separado

**Frontend Público (Next.js):**
- Deploy independente no Vercel (ou subdomínio: `www.example.com`)
- **SSR obrigatório** para SEO
- **SEM autenticação** (apenas exibição de imóveis públicos)
- **Bundle otimizado** (apenas componentes de busca/exibição)
- Acesso: usuários finais navegando imóveis
- CI/CD separado (Vercel automático)

**Frontend Admin (Next.js):**
- Deploy independente no Vercel (subdomínio: `app.example.com` ou `admin.example.com`)
- **CSR prioritário** (Client-Side Rendering) - performance não é crítica
- **Firebase Auth obrigatório** (login + proteção de rotas)
- **Bundle com componentes complexos** (tabelas, formulários, upload)
- Acesso: corretores e admins da imobiliária
- CI/CD separado (Vercel automático)
- Middleware protege TODAS as rotas (exceto `/login`)

### Vantagens da Separação Frontend Público vs Admin:

**1. Segurança:**
- Frontend público NUNCA possui código de autenticação ou rotas admin
- Impossível vazar tokens ou lógica de negócio sensível
- Admin completamente isolado em subdomínio separado

**2. Performance:**
- Frontend público: bundle minimalista (~150KB) - crítico para SEO
- Frontend admin: bundle maior permitido (~300KB) - UX mais rica

**3. SEO:**
- Frontend público 100% otimizado para SSR/SSG
- Nenhuma rota protegida interferindo com crawlers

**4. Deploy e Desenvolvimento:**
- Deploys independentes (mudança no admin não afeta o público)
- Times diferentes podem trabalhar em paralelo
- Rollback independente em caso de bugs

**5. Escalabilidade:**
- Frontend público pode ter cache agressivo (CDN)
- Frontend admin pode ter rate limiting mais restritivo

**6. Manutenção:**
- Código mais simples e focado em cada projeto
- Menos condicionais do tipo "if (isAdmin)"
- Testes mais diretos

### Comunicação Backend ↔ Frontends:
- Ambos os frontends chamam o mesmo backend via `NEXT_PUBLIC_API_URL` (env var)
- Produção:
  - Frontend Público: `www.example.com` → Backend: `api.example.com`
  - Frontend Admin: `app.example.com` → Backend: `api.example.com`
- Desenvolvimento:
  - Frontend Público: `localhost:3000` → Backend: `localhost:8080`
  - Frontend Admin: `localhost:3001` → Backend: `localhost:8080`
- Backend valida autenticação via Firebase token (apenas frontend admin envia tokens)

## 20. Regras de Distribuição de Leads e Co-Corretagem

### 20.1 Papéis de Corretores (PropertyBrokerRole)

#### **Captador (Originating Broker)** - O "Dono" do Ativo
```
Definição: Corretor que trouxe o imóvel do proprietário para a plataforma

Direitos:
- Cria o Property (ownership do ativo)
- Define visibilidade inicial (private, network, marketplace, public)
- Pode criar Listing canônico
- Define % de comissão para co-corretagem
- Pode revogar visibilidade a qualquer momento
- É SEMPRE is_primary por padrão (recebe leads primeiro)
- Aparece como "Captador" no dashboard admin

Responsabilidades:
- Manter dados do Property atualizados
- Responder ao proprietário
- Confirmar status/preço periodicamente (PROMPT 08)
- Definir política de compartilhamento

Criação:
- Automática na importação (PROMPT 02)
- Automática no cadastro manual de Property
- BrokerID = corretor autenticado
```

#### **Vendedor (Selling Broker)** - O Closer
```
Definição: Corretor que TEM um cliente interessado e busca imóveis para parceria

Como entra:
- Vê imóvel na busca interna do admin (visibilidade: network ou marketplace)
- Clica botão "Tenho um cliente para este imóvel"
- Sistema cria PropertyBrokerRole automaticamente
- Notifica o captador (aprovação automática no MVP)

Direitos:
- Pode criar seu próprio Listing (não-canônico, interno)
- Recebe notificação de novos leads do Property
- Pode atualizar status de negociação do lead
- Compartilha comissão conforme % definida pelo captador

Limitações:
- NÃO pode editar dados do Property (apenas captador)
- NÃO pode mudar visibilidade
- Seu Listing é interno (não é o canônico público)
- NÃO pode ser is_primary (a menos que captador transfira)
```

#### **Co-corretor (Co-Broker)** - Apoiador Passivo
```
Definição: Corretor adicional na negociação (indicação, parceiro)

Como entra:
- Adicionado manualmente por captador ou vendedor
- Ex: "Fulano me indicou o cliente, vou dar 20% para ele"

Direitos:
- Visualiza detalhes do Property no admin
- Recebe notificação de progresso da negociação
- Comissão definida manualmente (split futuro)

Limitações:
- NÃO pode editar nada
- Papel passivo (apenas registro para auditoria/comissão)
- NÃO pode ser is_primary
```

### 20.2 Visibilidade Escalonada de Properties

O **captador** controla quem pode ver o imóvel através do campo `Property.visibility`:

#### **Level 1: Private (Privado)**
```
Visível para: Apenas o captador
Uso: Imóvel recém-captado, ainda não validado pelo proprietário
Busca interna (admin): NÃO aparece para outros corretores
Frontend público: NÃO aparece
Leads: Apenas captador recebe
```

#### **Level 2: Network (Rede)**
```
Visível para: Todos os corretores do MESMO tenant (imobiliária)
Uso: Captador quer compartilhar com sua própria equipe
Busca interna (admin): Aparece com badge "Rede"
Frontend público: NÃO aparece
Botão disponível: "Tenho um cliente para este imóvel"
Leads: Captador recebe (is_primary), mas outros podem se tornar selling_broker
```

#### **Level 3: Marketplace**
```
Visível para: Todos os corretores de TODOS os tenants (co-corretagem aberta)
Uso: Captador quer máximo alcance via parceria
Busca interna (admin): Aparece com badge "Marketplace" + % comissão oferecida
Frontend público: NÃO aparece (ainda não público)
Botão disponível: "Tenho um cliente para este imóvel"
Leads: Captador recebe, sistema notifica todos os selling_brokers ativos
Comissão: % definida pelo captador (ex: "ofereço 40% para quem vender")
```

#### **Level 4: Public (Público)**
```
Visível para: Internet (Google, SEO, redes sociais)
Uso: Canonical Listing ativo, imóvel pronto para divulgação
Busca interna (admin): Aparece (qualquer corretor pode virar selling_broker)
Frontend público: APARECE (único nível visível publicamente)
Leads: Vai para is_primary (captador por padrão)
SEO: Indexado, slug amigável, meta tags
```

**Regra de Ouro:** Apenas 1 Property, mas visibilidade controlada pelo captador. Elimina duplicação.

### 20.3 Fluxo "Tenho um Cliente" (Co-Corretagem)

#### **Cenário:** Vendedor tem cliente de compra, busca imóvel compatível

**Passo 1: Busca Interna (Dashboard Admin)**
```
Vendedor autenticado usa busca no admin:
- Filtros: tipo, cidade, bairro, preço, quartos, etc.
- Visibilidade: APENAS network (seu tenant) OU marketplace (todos)
- Resultado: Properties com visibilidade adequada
```

**Passo 2: Manifestar Interesse**
```
Vendedor clica botão: [Tenho um cliente para este imóvel]

Backend cria PropertyBrokerRole:
{
  property_id: "prop123",
  broker_id: "broker456", // vendedor
  tenant_id: "tenant_do_vendedor",
  role: "selling_broker",
  status: "pending_approval", // ⭐ AGUARDA APROVAÇÃO DO CAPTADOR
  commission_percentage: X, // copiado de Property.co_broker_commission
  is_primary: false,
  created_at: now()
}

Backend notifica captador:
- Email: "Corretor X solicitou parceria no imóvel Y. Clique para aprovar/rejeitar."
- Dashboard: badge "1 solicitação pendente"
- ActivityLog: property_selling_broker_requested

Captador DEVE aprovar ou rejeitar:
- Aprovar → status: "active" (vendedor passa a receber notificações de leads)
- Rejeitar → status: "rejected" (vendedor não tem mais acesso ao Property)
- Timeout: 7 dias sem resposta → auto-rejeição

MVP: Aprovação MANUAL (captador tem controle total)
Benefícios: Evita banalização, spam e conflitos
```

**Passo 3: Lead Chega (Frontend Público)**
```
Usuário final clica "Falar no WhatsApp" no site público

Backend:
1. Cria Lead associado ao Property
2. Busca corretor primary (algoritmo 20.4)
3. Retorna WhatsApp do primary (captador por padrão)
4. Notifica TODOS os PropertyBrokerRole ativos:
   - Captador (is_primary: true) → WhatsApp redirect
   - Vendedores (selling_broker) → Email + Dashboard notification
   - Co-corretores (co_broker) → Dashboard notification

Frontend público:
- Redireciona usuário para WhatsApp do primary
- Mensagem pré-preenchida com lead_id
```

**Passo 4: Atendimento e Fechamento**
```
MVP: Qualquer corretor (captador ou vendedor) pode responder o lead
Futuro: Exclusividade temporária, SLA, escalação

Fechamento (fora do MVP, mas preparado):
- Corretor marca lead como "vendido"
- Sistema registra quem fechou (ActivityLog)
- Split de comissão (futuro) usa PropertyBrokerRole.commission_percentage
```

### 20.4 Algoritmo de Seleção do Primary Broker

**Objetivo:** Determinar qual corretor recebe o lead primeiro (WhatsApp redirect).

```go
func GetPrimaryBroker(propertyID string) (*Broker, error) {
    // 1. Buscar PropertyBrokerRole com is_primary = true
    roles := QueryPropertyBrokerRoles(propertyID)

    for _, role := range roles {
        if role.IsPrimary && role.Broker.Phone != "" {
            return role.Broker, nil
        }
    }

    // 2. Fallback: captador (originating_broker)
    for _, role := range roles {
        if role.Role == "originating_broker" && role.Broker.Phone != "" {
            return role.Broker, nil
        }
    }

    // 3. Fallback: primeiro selling_broker (ordenado por created_at)
    sellingBrokers := FilterByRole(roles, "selling_broker")
    SortByCreatedAt(sellingBrokers) // ASC

    for _, role := range sellingBrokers {
        if role.Broker.Phone != "" {
            return role.Broker, nil
        }
    }

    // 4. Erro crítico (não deve acontecer se validações estiverem corretas)
    return nil, errors.New("no_primary_broker_found")
}
```

**Regras:**
- Prioridade: is_primary > originating_broker > selling_broker (mais antigo)
- Phone obrigatório (validado no cadastro do Broker)
- Se nenhum corretor tem phone válido → erro (frontend exibe formulário)

### 20.5 Notificação Multi-Corretor

**Objetivo:** Garantir que TODOS os corretores envolvidos saibam do lead.

#### **Lead via WhatsApp:**
```go
func NotifyLeadCreated(lead Lead) error {
    // 1. Primary recebe WhatsApp redirect (tempo real)
    primary, err := GetPrimaryBroker(lead.PropertyID)
    if err != nil {
        return err
    }
    // Frontend redireciona usuário final para WhatsApp do primary
    // (notificação implícita via mensagem do cliente)

    // 2. Buscar TODOS os outros corretores ativos
    allBrokers := GetAllActiveBrokers(lead.PropertyID)

    for _, broker := range allBrokers {
        if broker.ID == primary.ID {
            continue // primary já foi notificado via WhatsApp
        }

        // 2a. Email assíncrono
        SendEmail(broker.Email, EmailTemplate{
            Subject: "Novo lead no imóvel " + lead.PropertyAddress,
            Body: "Cliente interessado via WhatsApp. Lead #" + lead.ID,
            CTA: "Ver no dashboard"
        })

        // 2b. Notificação dashboard (Firestore real-time)
        CreateNotification(broker.ID, Notification{
            Type: "new_lead",
            LeadID: lead.ID,
            PropertyID: lead.PropertyID,
            Message: "Novo lead via WhatsApp",
            CreatedAt: time.Now()
        })
    }

    // 3. ActivityLog para auditoria
    LogEvent(ActivityLog{
        EventType: "lead_created_whatsapp",
        LeadID: lead.ID,
        PropertyID: lead.PropertyID,
        Metadata: map[string]interface{}{
            "primary_broker_id": primary.ID,
            "notified_brokers": GetBrokerIDs(allBrokers),
            "total_notified": len(allBrokers)
        }
    })

    return nil
}
```

#### **Lead via Formulário:**
```go
func NotifyFormLead(lead Lead) error {
    primary, _ := GetPrimaryBroker(lead.PropertyID)

    // 1. Primary recebe email IMEDIATO (alta prioridade)
    SendEmail(primary.Email, EmailTemplate{
        Subject: "[URGENTE] Novo contato via formulário",
        Body: lead.Message,
        ClientInfo: lead.Name + " - " + lead.Phone,
        CTA: "Responder agora"
    })

    // 2. Outros corretores: dashboard apenas (menos urgente)
    allBrokers := GetAllActiveBrokers(lead.PropertyID)
    for _, broker := range allBrokers {
        if broker.ID != primary.ID {
            CreateNotification(broker.ID, Notification{
                Type: "new_form_lead",
                LeadID: lead.ID,
                Message: "Novo contato via formulário",
                CreatedAt: time.Now()
            })
        }
    }

    return nil
}
```

### 20.6 Permissões de Visualização de Leads (Dashboard Admin)

**Regra:** Corretor vê leads de Properties onde ele possui PropertyBrokerRole ativo.

#### **Backend (Go):**
```go
// GET /api/v1/tenants/:tenantId/brokers/:brokerId/leads
func GetBrokerLeads(tenantID, brokerID string) ([]Lead, error) {
    // 1. Buscar todos PropertyBrokerRole do corretor
    roles := QueryPropertyBrokerRoles(
        "broker_id = ? AND tenant_id = ?",
        brokerID,
        tenantID
    )

    // 2. Extrair property_ids
    propertyIDs := ExtractPropertyIDs(roles)

    // 3. Buscar leads desses Properties
    leads := QueryLeads(
        "property_id IN (?) AND tenant_id = ?",
        propertyIDs,
        tenantID
    )

    return leads, nil
}
```

#### **Frontend (Next.js):**
```typescript
// Dashboard Admin - LeadTable
const { data: leads } = useQuery({
  queryKey: ['leads', tenantId, brokerId],
  queryFn: async () => {
    // Backend filtra automaticamente por PropertyBrokerRole
    return api.get(`/tenants/${tenantId}/brokers/${brokerId}/leads`)
  }
})

// Filtros disponíveis (client-side após fetch):
// - Por Property
// - Por status (new, contacted, qualified, lost)
// - Por channel (whatsapp, form)
// - Por data
```

**Regra de Negócio:**
- Captador vê TODOS os leads do Property (sempre)
- Vendedor vê TODOS os leads do Property (cooperação total)
- Co-corretor vê TODOS os leads (transparência)
- Admin do tenant vê TODOS os leads do tenant

### 20.7 Campo Phone Obrigatório no Broker

**Objetivo:** Garantir que todo corretor pode receber leads via WhatsApp.

#### **Modelo Broker (Go):**
```go
type Broker struct {
    ID       string `firestore:"id" json:"id"`
    TenantID string `firestore:"tenant_id" json:"tenant_id"`

    Name  string `firestore:"name" json:"name" validate:"required"`
    Email string `firestore:"email" json:"email" validate:"required,email"`

    // ⭐ OBRIGATÓRIO - Formato E.164
    Phone string `firestore:"phone" json:"phone" validate:"required,e164"`
    // Exemplo: +5511999999999 (Brasil)

    Role      string    `firestore:"role" json:"role"` // admin | broker
    Status    string    `firestore:"status" json:"status"` // active | inactive
    CreatedAt time.Time `firestore:"created_at" json:"created_at"`
    UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}
```

#### **Validação no Signup (PROMPT 09):**
```go
func CreateBroker(data BrokerInput) error {
    // Validar phone obrigatório
    if data.Phone == "" {
        return errors.New("phone_required")
    }

    // Validar formato E.164 (+5511999999999)
    if !isValidE164(data.Phone) {
        return errors.New("invalid_phone_format")
    }

    // Validar phone único por tenant (opcional, mas recomendado)
    existing := FindBrokerByPhone(data.TenantID, data.Phone)
    if existing != nil {
        return errors.New("phone_already_registered")
    }

    // Criar broker
    broker := &Broker{
        ID:       uuid.New().String(),
        TenantID: data.TenantID,
        Name:     data.Name,
        Email:    data.Email,
        Phone:    data.Phone,
        Role:     "broker",
        Status:   "active",
        CreatedAt: time.Now(),
    }

    return SaveBroker(broker)
}
```

#### **Tratamento de Erro (WhatsApp Flow):**
```go
func GetWhatsAppNumber(propertyID string) (string, error) {
    primary, err := GetPrimaryBroker(propertyID)

    if err != nil || primary.Phone == "" {
        // Fallback: buscar próximo corretor com phone válido
        roles := GetAllActiveBrokers(propertyID)
        for _, role := range roles {
            if role.Broker.Phone != "" {
                return role.Broker.Phone, nil
            }
        }

        // Erro crítico: nenhum corretor tem phone
        return "", errors.New("no_phone_available")
    }

    return primary.Phone, nil
}

// Frontend (Next.js):
if (error.code === "no_phone_available") {
    alert("Imóvel indisponível para contato via WhatsApp no momento. Use o formulário abaixo.")
    // Exibir apenas formulário de contato
}
```

### 20.8 Mudança de Primary Broker

**Objetivo:** Permitir que captador transfira atendimento de leads para outro corretor.

#### **Endpoint:**
```
PATCH /api/v1/tenants/:tenantId/properties/:propertyId/primary-broker

Body:
{
  "new_primary_broker_id": "broker123"
}

Headers:
Authorization: Bearer <firebase_jwt>
```

#### **Regras de Permissão:**
- Apenas **captador** (originating_broker) pode alterar
- OU **admin** do tenant
- Novo primary DEVE ter PropertyBrokerRole ativo no Property
- Apenas 1 corretor pode ter is_primary: true (transação atômica)

#### **Implementação (Go):**
```go
func ChangePrimaryBroker(propertyID, newBrokerID, requestingBrokerID string) error {
    // 1. Validar permissão
    requestingRole := GetBrokerRole(propertyID, requestingBrokerID)
    if requestingRole.Role != "originating_broker" && !IsAdmin(requestingBrokerID) {
        return errors.New("permission_denied")
    }

    // 2. Validar novo primary existe
    newRole := GetBrokerRole(propertyID, newBrokerID)
    if newRole == nil {
        return errors.New("broker_not_found_in_property")
    }

    // 3. Transação Firestore (atômico)
    err := firestore.RunTransaction(ctx, func(tx *firestore.Transaction) error {
        // 3a. Remover is_primary de TODOS
        allRoles := QueryPropertyBrokerRoles(propertyID)
        for _, role := range allRoles {
            tx.Update(role.Ref, []firestore.Update{
                {Path: "is_primary", Value: false},
            })
        }

        // 3b. Setar is_primary no novo
        tx.Update(newRole.Ref, []firestore.Update{
            {Path: "is_primary", Value: true},
            {Path: "updated_at", Value: time.Now()},
        })

        return nil
    })

    if err != nil {
        return err
    }

    // 4. ActivityLog
    LogEvent(ActivityLog{
        EventType: "primary_broker_changed",
        PropertyID: propertyID,
        ActorID: requestingBrokerID,
        Metadata: map[string]interface{}{
            "old_primary": requestingRole.BrokerID,
            "new_primary": newBrokerID,
        }
    })

    // 5. Notificar novo primary
    SendEmail(newRole.Broker.Email, EmailTemplate{
        Subject: "Você agora é o corretor principal",
        Body: "Você receberá os leads do imóvel X",
    })

    return nil
}
```

### 20.9 Cadastro Direto pelo Proprietário (Evolução Futura)

**Cenário:** Proprietário cadastra seu próprio imóvel no frontend público (self-service).

#### **Regras Estruturais:**
```
Quando: Proprietário clica "Anunciar meu imóvel" no site público
Fluxo:
1. Formulário simplificado (endereço, tipo, preço, fotos)
2. Criação de Owner (sem login, apenas dados básicos)
3. Criação de Property com status: "pending_broker_assignment"
4. Sistema atribui "Ecossistema Imob" como captador (tenant_id especial)

PropertyBrokerRole criado automaticamente:
{
  property_id: "prop123",
  broker_id: "broker_ecosystem", // corretor da plataforma
  tenant_id: "tenant_ecosystem", // tenant da plataforma
  role: "originating_broker",
  is_primary: true,
  commission_percentage: 100, // plataforma fica com 100% até venda
  created_at: now()
}

Visibilidade inicial: "marketplace" (disponível para todos os corretores)
Comissão de co-corretagem: definida pela plataforma (ex: 50%)
```

#### **Vantagens Estratégicas:**
- Plataforma atua como **marketplace neutro + imobiliária ativa**
- Proprietários sem corretor podem anunciar
- Corretores têm acesso a carteira exclusiva da plataforma
- Receita dupla: SaaS + comissão de vendas
- Network effect: mais imóveis → mais corretores → mais imóveis

#### **Implementação (Pós-MVP):**
- Frontend público: formulário "Anunciar Grátis"
- Backend: endpoint público (sem autenticação)
- Owner: criado sem login, apenas email/phone
- Property: criado com tenant_id da plataforma
- Notificação: corretores veem novo imóvel na busca marketplace

**IMPORTANTE:** No MVP, cadastro é apenas via corretores autenticados. Self-service é evolução futura.

### 20.10 Resumo Executivo - Distribuição de Leads

**Problema Resolvido:** "Como distribuir leads de forma justa quando múltiplos corretores estão envolvidos?"

**Solução:**
1. ✅ **Lead pertence ao Property** (não ao corretor)
2. ✅ **Primary recebe primeiro** (WhatsApp redirect)
3. ✅ **Todos são notificados** (email + dashboard)
4. ✅ **Visibilidade escalonada** (private → network → marketplace → public)
5. ✅ **Co-corretagem transparente** (% definida, auditada, preparada para split)
6. ✅ **Fallbacks claros** (algoritmo determinístico)
7. ✅ **Permissões por role** (captador > vendedor > co-corretor)

**Diferencial Competitivo:**
- Portais: Lead vendido 5x (competição)
- Seu modelo: Lead compartilhado 1x (cooperação)
- Resultado: Confiança + network effect + marketplace justo

## 21. Conformidade com LGPD (Lei Geral de Proteção de Dados)

### 21.1 Contexto Legal

A LGPD (Lei nº 13.709/2018) regula o tratamento de dados pessoais no Brasil. Este projeto coleta e processa dados de:
- **Proprietários** (owners): CPF, nome, email, telefone, endereço do imóvel
- **Corretores** (brokers): CPF/CNPJ, nome, email, telefone, CRECI
- **Leads** (potenciais compradores): nome, email, telefone, mensagens

**IMPORTANTE**: Violações à LGPD podem resultar em multas de até 2% do faturamento (limitado a R$ 50 milhões por infração) + danos à reputação.

### 21.2 Princípios da LGPD Aplicados

#### **Finalidade**
```
Dados coletados apenas para propósitos específicos e legítimos:
- Proprietários: gestão de imóveis + comunicação sobre vendas
- Corretores: autenticação + distribuição de leads + comissões
- Leads: atendimento comercial + negociação imobiliária

❌ PROIBIDO: Usar dados para marketing sem consentimento explícito
❌ PROIBIDO: Compartilhar dados com terceiros não autorizados
```

#### **Adequação**
```
Tratamento compatível com finalidades informadas ao titular:
✅ Lead fornece telefone → usado para contato comercial via WhatsApp
✅ Proprietário fornece CPF → usado para contrato de venda
❌ ERRADO: Usar telefone do lead para vender outros serviços
```

#### **Necessidade**
```
Coletar apenas dados estritamente necessários:
✅ OBRIGATÓRIO: Nome, email, telefone (comunicação essencial)
✅ OPCIONAL: CPF (apenas quando necessário para contrato)
❌ DESNECESSÁRIO: Data de nascimento, estado civil, renda (MVP)
```

#### **Transparência**
```
Titular deve saber como seus dados são usados:
✅ Política de Privacidade clara e acessível (link no footer)
✅ Termos de Consentimento explícitos (checkbox obrigatório)
✅ Dashboard para proprietários/leads consultarem seus dados
```

#### **Segurança**
```
Proteção técnica e administrativa:
✅ HTTPS obrigatório (TLS 1.3)
✅ Firestore Security Rules (isolamento por tenant)
✅ Firebase Auth (autenticação segura)
✅ Hashing de eventos (SHA256 para auditoria)
✅ Logs de acesso (ActivityLog)
```

### 21.3 Base Legal (Artigo 7º da LGPD)

#### **Consentimento (Art. 7º, I)**
```
Aplicável a: Leads (contato comercial)

Implementação:
- Checkbox no formulário de contato (não pré-marcado)
- Texto: "Concordo em receber contato sobre este imóvel"
- Consentimento registrado no Lead.consent_given: true
- Possibilidade de revogar consentimento (direito do titular)

Modelo Lead (ATUALIZADO):
type Lead struct {
    // ... campos existentes

    // 🆕 LGPD - Consentimento
    ConsentGiven   bool      `firestore:"consent_given" json:"consent_given"`
    ConsentText    string    `firestore:"consent_text" json:"consent_text"` // Texto exibido
    ConsentDate    time.Time `firestore:"consent_date" json:"consent_date"`
    ConsentIP      string    `firestore:"consent_ip" json:"consent_ip"` // IP do usuário
    ConsentRevoked bool      `firestore:"consent_revoked" json:"consent_revoked"`
    RevokedAt      time.Time `firestore:"revoked_at,omitempty" json:"revoked_at,omitempty"`
}
```

#### **Execução de Contrato (Art. 7º, V)**
```
Aplicável a: Proprietários e Corretores (relação contratual)

Justificativa:
- Corretor precisa dos dados do proprietário para intermediar venda
- Plataforma precisa dos dados do corretor para autenticação/comissões
- Imóvel precisa estar cadastrado para ser anunciado

Não requer consentimento explícito (necessário para cumprimento do contrato)
```

#### **Legítimo Interesse (Art. 7º, IX)**
```
Aplicável a: Logs de auditoria, prevenção de fraude

Exemplos:
- ActivityLog: necessário para segurança e compliance
- Detecção de duplicação: proteção da qualidade do marketplace
- Logs de acesso: prevenção de abuso

ATENÇÃO: Legítimo interesse NÃO pode sobrepor direitos do titular
Exemplo: Enviar marketing sem consentimento = ILEGAL
```

### 21.4 Direitos dos Titulares (Artigos 17 a 22)

#### **Confirmação e Acesso (Art. 18, I e II)**
```
Titular pode pedir: "Vocês têm meus dados? Quais?"

Endpoint obrigatório:
GET /api/v1/data-subject-request?email={email}&type=access

Resposta:
{
  "data_found": true,
  "categories": ["lead", "owner"],
  "details": {
    "lead": {
      "id": "lead123",
      "created_at": "2024-01-15",
      "property": "Apartamento em São Paulo",
      "data": {
        "name": "João Silva",
        "email": "joao@example.com",
        "phone": "+5511999999999",
        "message": "Tenho interesse neste imóvel"
      }
    }
  }
}

Prazo: 15 dias (Art. 19, §3º)
```

#### **Correção (Art. 18, III)**
```
Titular pode pedir: "Meu telefone está errado, corrijam"

Endpoint:
PATCH /api/v1/data-subject-request
{
  "email": "joao@example.com",
  "type": "correction",
  "field": "phone",
  "new_value": "+5511888888888"
}

Backend:
- Valida identidade (email + código de verificação)
- Atualiza dado
- Registra correção no ActivityLog
```

#### **Anonimização/Exclusão (Art. 18, IV e VI)**
```
Titular pode pedir: "Deletem meus dados"

Endpoint:
DELETE /api/v1/data-subject-request
{
  "email": "joao@example.com",
  "type": "deletion",
  "reason": "Não tenho mais interesse"
}

Regras de Retenção:
✅ Lead sem negociação em andamento → DELETAR imediatamente
⚠️ Lead com negociação ativa → ANONIMIZAR (manter histórico)
⚠️ Proprietário com imóvel vendido → ANONIMIZAR (obrigação fiscal por 5 anos)
❌ Dados fiscais/contratuais → NÃO deletar (base legal: obrigação legal)

Anonimização:
- Nome → "Titular Anonimizado {hash}"
- Email → "anonimizado_{hash}@deleted.local"
- Telefone → "+55119999XXXX"
- CPF → "XXX.XXX.XXX-XX"
```

#### **Portabilidade (Art. 18, V)**
```
Titular pode pedir: "Quero meus dados em formato legível"

Endpoint:
GET /api/v1/data-subject-request/export?email={email}

Resposta: JSON ou CSV com todos os dados
```

#### **Revogação de Consentimento (Art. 18, IX)**
```
Titular pode pedir: "Não quero mais ser contatado"

Implementação:
- Botão "Não quero mais receber contatos" em emails
- Link: /api/v1/consent/revoke?token={lead_id_hash}
- Backend: Lead.consent_revoked = true
- Corretor NÃO pode mais contatar (LGPD + Lei do Spam)
```

### 21.5 Implementação Técnica

#### **Modelo Owner (ATUALIZADO - LGPD)**
```go
type Owner struct {
    ID       string `firestore:"id" json:"id"`
    TenantID string `firestore:"tenant_id" json:"tenant_id"`

    Name  string `firestore:"name" json:"name" validate:"required"`
    Email string `firestore:"email" json:"email" validate:"required,email"`
    Phone string `firestore:"phone" json:"phone" validate:"required,e164"`

    // CPF apenas quando necessário (contrato de venda)
    CPF string `firestore:"cpf,omitempty" json:"cpf,omitempty" validate:"omitempty,cpf"`

    // 🆕 LGPD - Consentimento e Origem
    ConsentGiven     bool      `firestore:"consent_given" json:"consent_given"`
    ConsentText      string    `firestore:"consent_text" json:"consent_text"`
    ConsentDate      time.Time `firestore:"consent_date" json:"consent_date"`
    ConsentOrigin    string    `firestore:"consent_origin" json:"consent_origin"` // "broker" | "self_service"

    // 🆕 LGPD - Anonimização
    IsAnonymized     bool      `firestore:"is_anonymized" json:"is_anonymized"`
    AnonymizedAt     time.Time `firestore:"anonymized_at,omitempty" json:"anonymized_at,omitempty"`
    AnonymizationReason string `firestore:"anonymization_reason,omitempty" json:"anonymization_reason,omitempty"`

    CreatedAt time.Time `firestore:"created_at" json:"created_at"`
    UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}
```

#### **Política de Retenção**
```go
// Cloud Scheduler (mensal)
func AnonymizeInactiveData() {
    // 1. Leads sem resposta há 2 anos
    leads := QueryLeads("status = 'new' AND created_at < ?", time.Now().AddDate(-2, 0, 0))
    for _, lead := range leads {
        AnonymizeLead(lead.ID)
    }

    // 2. Proprietários que removeram imóveis há 5 anos (obrigação fiscal cumprida)
    owners := QueryOwners("last_property_removed_at < ?", time.Now().AddDate(-5, 0, 0))
    for _, owner := range owners {
        AnonymizeOwner(owner.ID)
    }
}

func AnonymizeLead(leadID string) {
    lead := GetLead(leadID)

    lead.Name = "Titular Anonimizado " + HashID(leadID)[:8]
    lead.Email = "anonimizado_" + HashID(leadID)[:8] + "@deleted.local"
    lead.Phone = "+5511999900000"
    lead.Message = "[MENSAGEM REMOVIDA - LGPD]"
    lead.IsAnonymized = true
    lead.AnonymizedAt = time.Now()
    lead.AnonymizationReason = "retention_policy"

    UpdateLead(lead)

    // ActivityLog para auditoria
    LogEvent(ActivityLog{
        EventType: "data_anonymized",
        EntityType: "lead",
        EntityID: leadID,
        Metadata: map[string]interface{}{
            "reason": "retention_policy",
            "original_created_at": lead.CreatedAt,
        }
    })
}
```

#### **Endpoints de LGPD (OBRIGATÓRIOS)**
```
POST   /api/v1/lgpd/data-subject-request (criar solicitação)
GET    /api/v1/lgpd/data-subject-request/:id (consultar status)
GET    /api/v1/lgpd/export?email={email}&token={verification_code} (exportar dados)
DELETE /api/v1/lgpd/delete?email={email}&token={verification_code} (deletar/anonimizar)
POST   /api/v1/lgpd/consent/revoke (revogar consentimento)
```

### 21.6 Documentação Obrigatória

#### **Política de Privacidade (Frontend Público)**
```
URL: /politica-de-privacidade
Link obrigatório: Footer de TODAS as páginas

Conteúdo mínimo:
1. Identificação do Controlador (empresa + CNPJ + endereço + email DPO)
2. Tipos de dados coletados (nome, email, telefone, CPF quando aplicável)
3. Finalidades do tratamento (intermediação imobiliária, contato comercial)
4. Base legal (consentimento, execução de contrato, legítimo interesse)
5. Compartilhamento de dados (corretores autorizados, não há venda a terceiros)
6. Direitos do titular (acesso, correção, exclusão, portabilidade, revogação)
7. Como exercer direitos (email: lgpd@example.com)
8. Prazo de retenção (2 anos para leads inativos, 5 anos após venda)
9. Segurança (Firebase Auth, HTTPS, Firestore Security Rules)
10. Cookies (se aplicável - Google Analytics, etc.)
11. Alterações na política (data de última atualização)
```

#### **Termos de Consentimento (Formulários)**
```html
<!-- Formulário de Contato -->
<form>
  <input name="name" required />
  <input name="email" required />
  <input name="phone" required />
  <textarea name="message"></textarea>

  <!-- ⭐ LGPD - Consentimento Explícito -->
  <label>
    <input type="checkbox" name="consent" required />
    Concordo com a <a href="/politica-de-privacidade" target="_blank">
      Política de Privacidade
    </a> e autorizo o uso dos meus dados para contato sobre este imóvel.
  </label>

  <button type="submit">Enviar</button>
</form>
```

### 21.7 DPO (Data Protection Officer)

#### **Obrigatoriedade (Art. 41)**
```
Empresa DEVE ter DPO se:
- Tratamento de dados é atividade principal (SIM - plataforma imobiliária)
- Tratamento regular e sistemático de dados (SIM - leads contínuos)
- Grande volume de dados sensíveis (TALVEZ - depende da escala)

MVP: Recomendado ter DPO ou consultor externo
Produção: OBRIGATÓRIO (conforme crescimento)
```

#### **Responsabilidades do DPO**
```
1. Aceitar reclamações de titulares (email: lgpd@example.com)
2. Prestar esclarecimentos sobre tratamento de dados
3. Orientar colaboradores sobre boas práticas
4. Atuar como canal de comunicação com ANPD (Autoridade Nacional)
5. Elaborar Relatório de Impacto (RIPD) quando necessário
```

### 21.8 Registro de Atividades de Tratamento (Art. 37)

```
Obrigatório para controladores (mesmo sem DPO formal)

Estrutura:
1. Categoria de dados: Leads
   - Dados: nome, email, telefone, mensagem
   - Finalidade: Atendimento comercial
   - Base legal: Consentimento
   - Compartilhamento: Corretor responsável pelo imóvel
   - Retenção: 2 anos após último contato

2. Categoria de dados: Proprietários
   - Dados: nome, email, telefone, CPF (opcional)
   - Finalidade: Gestão de imóveis + contratos
   - Base legal: Execução de contrato
   - Compartilhamento: Corretor captador + vendedor (co-corretagem)
   - Retenção: 5 anos após venda (obrigação fiscal)

3. Categoria de dados: Corretores
   - Dados: nome, email, telefone, CRECI, CPF/CNPJ
   - Finalidade: Autenticação + distribuição de leads + comissões
   - Base legal: Execução de contrato
   - Compartilhamento: Interno (plataforma)
   - Retenção: Enquanto ativo + 5 anos após inativação
```

### 21.9 Incidentes de Segurança (Art. 48)

#### **Obrigação de Notificação**
```
Prazo: "em prazo razoável" (interpretado como 72h pela maioria dos juristas)
Quem notificar:
1. ANPD (Autoridade Nacional) - SEMPRE
2. Titular afetado - SE houver risco ou dano relevante

Exemplos:
- Vazamento de senhas → Notificar ANPD + titulares
- Acesso não autorizado a emails → Notificar ANPD + titulares
- Backup corrompido (sem vazamento) → Notificar apenas ANPD
```

#### **Implementação**
```go
// Em caso de incidente
func NotifyDataBreach(incident DataBreachIncident) {
    // 1. Log interno
    LogEvent(ActivityLog{
        EventType: "data_breach_detected",
        Severity: "critical",
        Metadata: incident,
    })

    // 2. Notificar DPO/Admin imediatamente
    SendAlert("DPO", "LGPD: Incidente de Segurança Detectado")

    // 3. Avaliar severidade
    if incident.AffectsPersonalData {
        // 4. Notificar ANPD (email oficial + formulário web)
        NotifyANPD(incident)

        // 5. Notificar titulares afetados (se risco relevante)
        if incident.HighRisk {
            affectedUsers := GetAffectedUsers(incident)
            for _, user := range affectedUsers {
                SendEmail(user.Email, EmailTemplate{
                    Subject: "IMPORTANTE: Incidente de Segurança",
                    Body: "Informamos que...", // Transparência total
                })
            }
        }
    }
}
```

### 21.10 Checklist de Conformidade LGPD

#### **Antes do MVP (Obrigatório)**
- [ ] Política de Privacidade publicada (frontend público)
- [ ] Termos de Consentimento em todos os formulários
- [ ] Campos de consentimento no modelo Lead
- [ ] Endpoint de revogação de consentimento
- [ ] HTTPS obrigatório (certificado SSL/TLS)
- [ ] Firestore Security Rules ativas
- [ ] ActivityLog funcionando (auditoria)

#### **MVP (Recomendado)**
- [ ] DPO nomeado (ou consultor externo)
- [ ] Email lgpd@example.com ativo
- [ ] Endpoint de acesso aos dados (data subject request)
- [ ] Endpoint de exclusão/anonimização
- [ ] Registro de Atividades de Tratamento documentado

#### **Pós-MVP (Antes de Produção)**
- [ ] Política de Retenção implementada (Cloud Scheduler)
- [ ] Relatório de Impacto (RIPD) elaborado
- [ ] Processo de notificação de incidentes testado
- [ ] Treinamento da equipe sobre LGPD
- [ ] Revisão jurídica da Política de Privacidade
- [ ] Termos de Uso completos

### 21.11 Penalidades por Não Conformidade

#### **Advertência**
```
Primeira infração leve ou correção rápida
Exemplo: Política de Privacidade desatualizada
```

#### **Multa Simples**
```
Até 2% do faturamento (limitado a R$ 50 milhões)
Exemplo: Não atender solicitação de exclusão no prazo
```

#### **Multa Diária**
```
Até R$ 50 milhões (total)
Exemplo: Continuar tratando dados após ordem de suspensão
```

#### **Bloqueio/Eliminação de Dados**
```
ANPD pode ordenar exclusão imediata
Exemplo: Tratamento sem base legal
```

#### **Suspensão do Banco de Dados**
```
Impede operação da plataforma
Exemplo: Incidente grave sem notificação
```

### 21.12 Impacto no Projeto

**Modelos Atualizados:**
- `Owner` → campos de consentimento + anonimização
- `Lead` → campos de consentimento + revogação + IP
- `Broker` → campos de consentimento (CRECI é dado sensível)

**Endpoints Novos:**
- `POST /api/v1/lgpd/data-subject-request`
- `GET /api/v1/lgpd/export`
- `DELETE /api/v1/lgpd/delete`
- `POST /api/v1/lgpd/consent/revoke`

**Frontend:**
- Página `/politica-de-privacidade`
- Checkbox de consentimento em formulários
- Link "Não quero mais receber contatos" em emails

**Backend:**
- Cloud Scheduler para anonimização automática
- Função de notificação de incidentes
- Logs de acesso aos dados (auditoria)

**Custo Adicional Estimado:**
- Consultor LGPD: R$ 2.000 - R$ 5.000 (one-time)
- DPO terceirizado: R$ 1.000 - R$ 3.000/mês
- Revisão jurídica: R$ 3.000 - R$ 8.000 (one-time)

**IMPORTANTE**: Não conformidade com LGPD pode **inviabilizar o negócio**. Investimento em compliance é obrigatório, não opcional.

## 22. Identidade Visual e Design System (MVP)

### 22.1 Contexto

O projeto ainda **não possui nome definitivo, logo ou identidade visual estabelecida**.

Para garantir um **produto profissional desde o MVP**, devemos implementar:
- Design system provisório baseado em **referências do mercado** (Zillow, Zap Imóveis, QuintoAndar)
- Elementos visuais **placeholder** que possam ser facilmente substituídos
- UX/UI moderno, clean e **retenção de usuários** como prioridade

### 22.2 Referências de Design

**Inspirações (mercado americano e brasileiro):**
- **Zillow** (EUA): design clean, hierarquia visual clara, cards bem espaçados
- **Redfin** (EUA): navegação intuitiva, filtros visuais, mapas integrados
- **Zap Imóveis** (BR): layout familiar ao público brasileiro, CTA's evidentes
- **QuintoAndar** (BR): onboarding suave, microinterações, confiança visual

**Princípios de Design:**
1. **Clean e Espaçoso**: evitar poluição visual, breathing room
2. **Hierarquia Clara**: títulos, subtítulos, texto corpo bem definidos
3. **CTA Visível**: botões de ação principais sempre evidentes
4. **Mobile-First**: design responsivo prioritário (70% do tráfego imobiliário é mobile)
5. **Performance**: fast loading, imagens otimizadas, lazy loading

### 22.3 Design System Provisório (shadcn/ui + Tailwind)

**Paleta de Cores (Placeholder):**
```css
/* Primary - Azul Confiança (inspirado em Zillow) */
--primary: 214 100% 50%        /* #0066FF - Azul vibrante */
--primary-foreground: 0 0% 100% /* Branco */

/* Secondary - Cinza Neutro */
--secondary: 220 13% 91%        /* #E8EAED - Cinza claro */
--secondary-foreground: 220 9% 46% /* #6B7280 - Cinza médio */

/* Accent - Verde Sucesso (conversão) */
--accent: 142 71% 45%           /* #22C55E - Verde */
--accent-foreground: 0 0% 100%  /* Branco */

/* Destructive - Vermelho Alerta */
--destructive: 0 84% 60%        /* #EF4444 - Vermelho */

/* Background */
--background: 0 0% 100%         /* Branco puro */
--foreground: 222 47% 11%       /* Quase preto */

/* Muted (backgrounds secundários) */
--muted: 220 13% 95%            /* #F9FAFB - Cinza muito claro */
--muted-foreground: 220 9% 46%  /* #6B7280 */

/* Border */
--border: 220 13% 91%           /* #E8EAED */
--radius: 0.5rem                /* 8px - cantos levemente arredondados */
```

**Tipografia:**
```tsx
/* Font Stack (Next.js + Tailwind) */
import { Inter, Poppins } from 'next/font/google'

// Body text - legibilidade
const inter = Inter({ subsets: ['latin'] })

// Headings - impacto visual
const poppins = Poppins({
  subsets: ['latin'],
  weight: ['400', '500', '600', '700']
})

/* Tamanhos */
--text-xs: 0.75rem     /* 12px */
--text-sm: 0.875rem    /* 14px */
--text-base: 1rem      /* 16px */
--text-lg: 1.125rem    /* 18px */
--text-xl: 1.25rem     /* 20px */
--text-2xl: 1.5rem     /* 24px */
--text-3xl: 1.875rem   /* 30px */
--text-4xl: 2.25rem    /* 36px */
```

**Componentes Core (shadcn/ui):**
```bash
# Essenciais para o MVP
npx shadcn-ui@latest add button
npx shadcn-ui@latest add card
npx shadcn-ui@latest add input
npx shadcn-ui@latest add select
npx shadcn-ui@latest add dialog
npx shadcn-ui@latest add dropdown-menu
npx shadcn-ui@latest add table
npx shadcn-ui@latest add badge
npx shadcn-ui@latest add avatar
npx shadcn-ui@latest add skeleton
npx shadcn-ui@latest add toast
npx shadcn-ui@latest add checkbox
npx shadcn-ui@latest add label
```

### 22.4 Logo e Branding Placeholder

**Logo Temporário:**
```tsx
// components/ui/logo-placeholder.tsx
export function LogoPlaceholder({ variant = "full" }: { variant?: "full" | "icon" }) {
  if (variant === "icon") {
    return (
      <div className="w-10 h-10 bg-primary rounded-lg flex items-center justify-center">
        <Home className="w-6 h-6 text-primary-foreground" />
      </div>
    )
  }

  return (
    <div className="flex items-center gap-2">
      <div className="w-10 h-10 bg-primary rounded-lg flex items-center justify-center">
        <Home className="w-6 h-6 text-primary-foreground" />
      </div>
      <div className="flex flex-col">
        <span className="text-xl font-bold text-foreground">ImóvelHub</span>
        <span className="text-xs text-muted-foreground">Marketplace Imobiliário</span>
      </div>
    </div>
  )
}
```

**Ícone de Favoritos (fallback):**
```tsx
// public/favicon.svg (provisório)
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
  <rect width="100" height="100" fill="#0066FF" rx="20"/>
  <path d="M30 50 L50 30 L70 50 L70 75 L30 75 Z" fill="white"/>
</svg>
```

**Nome Provisório:**
- **ImóvelHub** (placeholder)
- Pode ser facilmente substituído via variável de ambiente:

```env
# .env.local
NEXT_PUBLIC_APP_NAME="ImóvelHub"
NEXT_PUBLIC_APP_TAGLINE="Marketplace Imobiliário"
```

### 22.5 UX/UI - Melhores Práticas

**1. Página Inicial (Pública)**

```tsx
// Layout inspirado em Zillow
<Hero>
  {/* CTA principal: busca de imóveis */}
  <SearchBar
    placeholder="Buscar por cidade, bairro ou referência..."
    onSearch={handleSearch}
  />
  {/* Filtros rápidos */}
  <QuickFilters options={['Comprar', 'Alugar', 'Terrenos']} />
</Hero>

<FeaturedProperties>
  {/* Grid de cards com fotos grandes */}
  <PropertyCard
    image={...}
    price={...}
    location={...}
    specs={{ bedrooms, bathrooms, area }}
    onFavorite={...}
    onContact={...}
  />
</FeaturedProperties>
```

**2. Card de Imóvel (Padrão de Mercado)**

```tsx
<Card className="overflow-hidden hover:shadow-lg transition-shadow">
  {/* Imagem com badge de status */}
  <div className="relative aspect-video">
    <Image src={coverPhoto} fill className="object-cover" />
    <Badge className="absolute top-2 right-2">Disponível</Badge>
    <Button
      variant="ghost"
      size="icon"
      className="absolute top-2 left-2 bg-white/80"
      onClick={handleFavorite}
    >
      <Heart />
    </Button>
  </div>

  {/* Conteúdo */}
  <CardContent className="p-4">
    <div className="flex justify-between items-start">
      <div>
        <h3 className="font-semibold text-xl">R$ 450.000</h3>
        <p className="text-sm text-muted-foreground">
          Apartamento • 3 quartos • 2 banheiros • 85m²
        </p>
        <p className="text-sm text-muted-foreground mt-1">
          Jardim Paulista, São Paulo - SP
        </p>
      </div>
    </div>

    <div className="flex gap-2 mt-4">
      <Button size="sm" className="flex-1">
        Ver Detalhes
      </Button>
      <Button size="sm" variant="outline">
        <MessageCircle className="w-4 h-4" />
      </Button>
    </div>
  </CardContent>
</Card>
```

**3. Skeleton Loading (Perceived Performance)**

```tsx
// Enquanto carrega dados, mostrar skeleton
<PropertyCardSkeleton />

function PropertyCardSkeleton() {
  return (
    <Card>
      <Skeleton className="aspect-video" />
      <CardContent className="p-4 space-y-3">
        <Skeleton className="h-6 w-32" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-3/4" />
        <div className="flex gap-2">
          <Skeleton className="h-9 flex-1" />
          <Skeleton className="h-9 w-9" />
        </div>
      </CardContent>
    </Card>
  )
}
```

**4. Formulários (LGPD-compliant + UX)**

```tsx
<form onSubmit={handleSubmit}>
  <Input
    label="Nome completo"
    placeholder="João Silva"
    required
  />
  <Input
    label="Email"
    type="email"
    placeholder="joao@example.com"
    required
  />
  <Input
    label="Telefone"
    type="tel"
    placeholder="(11) 99999-9999"
    required
  />
  <Textarea
    label="Mensagem"
    placeholder="Gostaria de agendar uma visita..."
  />

  {/* ⭐ LGPD - Checkbox de consentimento */}
  <div className="flex items-start gap-2 p-4 bg-muted rounded-lg">
    <Checkbox
      id="consent"
      required
      checked={consentGiven}
      onCheckedChange={setConsentGiven}
    />
    <Label htmlFor="consent" className="text-sm">
      Concordo com a{' '}
      <Link href="/politica-de-privacidade" className="underline">
        Política de Privacidade
      </Link>{' '}
      e autorizo o uso dos meus dados para contato sobre este imóvel.
    </Label>
  </div>

  <Button
    type="submit"
    size="lg"
    className="w-full"
    disabled={!consentGiven}
  >
    Enviar Mensagem
  </Button>
</form>
```

**5. Admin Dashboard (PROMPT 04b)**

```tsx
// Layout inspirado em Vercel/Linear
<DashboardLayout>
  <Sidebar>
    <LogoPlaceholder variant="icon" />
    <Nav items={[
      { label: 'Imóveis', icon: Home, href: '/admin/properties' },
      { label: 'Leads', icon: Users, href: '/admin/leads' },
      { label: 'Corretores', icon: UserCheck, href: '/admin/brokers' },
      { label: 'Co-corretagem', icon: Handshake, href: '/admin/partnerships' },
    ]} />
  </Sidebar>

  <Main>
    <Header>
      <h1>Imóveis</h1>
      <Button>Adicionar Imóvel</Button>
    </Header>

    {/* Filtros + Tabela */}
    <Filters />
    <DataTable
      columns={propertyColumns}
      data={properties}
      onRowClick={handleEdit}
    />
  </Main>
</DashboardLayout>
```

### 22.6 Microinterações e Feedback Visual

**Princípios:**
1. **Feedback Imediato**: toda ação do usuário tem resposta visual (loading, sucesso, erro)
2. **Animações Sutis**: transições suaves (150-300ms), sem exageros
3. **Estados Claros**: hover, active, disabled, loading sempre visíveis

**Exemplos:**

```tsx
// Botão com loading state
<Button disabled={isLoading}>
  {isLoading ? (
    <>
      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
      Salvando...
    </>
  ) : (
    'Salvar Imóvel'
  )}
</Button>

// Toast de sucesso
toast({
  title: "Imóvel cadastrado!",
  description: "O imóvel foi adicionado com sucesso.",
  variant: "success"
})

// Hover em cards
<Card className="transition-all hover:scale-[1.02] hover:shadow-lg cursor-pointer">
```

### 22.7 Responsividade (Mobile-First)

**Breakpoints (Tailwind padrão):**
```
sm: 640px   (tablets pequenos)
md: 768px   (tablets)
lg: 1024px  (desktops)
xl: 1280px  (desktops grandes)
2xl: 1536px (ultra-wide)
```

**Grid de Imóveis (responsivo):**
```tsx
<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
  {properties.map(property => <PropertyCard key={property.id} {...property} />)}
</div>
```

### 22.8 Assets Provisórios

**Imagens Placeholder (quando não houver foto):**
```tsx
// components/property-image-placeholder.tsx
export function PropertyImagePlaceholder({ type }: { type: PropertyType }) {
  const placeholderUrl = `https://placehold.co/800x600/0066FF/FFFFFF?text=${
    type === 'apartment' ? 'Apartamento' :
    type === 'house' ? 'Casa' :
    type === 'land' ? 'Terreno' : 'Comercial'
  }`

  return (
    <Image
      src={placeholderUrl}
      alt="Imagem em breve"
      fill
      className="object-cover"
    />
  )
}
```

**Ícones (Lucide React):**
```tsx
import {
  Home, Building, TreePine, Store, // Tipos de imóveis
  Bed, Bath, Car, Ruler,           // Especificações
  Heart, Share2, Phone, Mail,      // Ações
  MapPin, Search, Filter,          // Navegação
  ChevronRight, ChevronLeft,       // Carrosséis
  Upload, Trash, Edit, Check       // Admin
} from 'lucide-react'
```

### 22.9 Preparação para Rebranding

**Variáveis Centralizadas:**
```typescript
// lib/branding.ts
export const branding = {
  name: process.env.NEXT_PUBLIC_APP_NAME || "ImóvelHub",
  tagline: process.env.NEXT_PUBLIC_APP_TAGLINE || "Marketplace Imobiliário",
  logo: process.env.NEXT_PUBLIC_LOGO_URL || "/logo-placeholder.svg",
  favicon: process.env.NEXT_PUBLIC_FAVICON_URL || "/favicon.svg",
  primaryColor: process.env.NEXT_PUBLIC_PRIMARY_COLOR || "#0066FF",
  email: process.env.NEXT_PUBLIC_CONTACT_EMAIL || "contato@imovelhub.com.br",
  phone: process.env.NEXT_PUBLIC_CONTACT_PHONE || "+55 11 99999-9999"
}
```

**Quando houver marca definitiva:**
1. Atualizar variáveis de ambiente
2. Substituir logo/favicon no `/public`
3. Ajustar paleta de cores em `tailwind.config.ts`
4. **Zero refatoração de código necessária**

### 22.10 Checklist de Qualidade Visual

**Antes de entregar o MVP:**
- [ ] Logo placeholder implementado e substituível
- [ ] Paleta de cores aplicada consistentemente
- [ ] Tipografia hierárquica (headings vs body)
- [ ] Todos os botões têm estados hover/active/disabled
- [ ] Loading states em todas as ações assíncronas
- [ ] Skeleton loaders em carregamentos de dados
- [ ] Toasts para feedback de sucesso/erro
- [ ] Cards de imóveis seguem padrão de mercado (foto grande, preço destacado)
- [ ] Formulários LGPD-compliant (checkbox visível)
- [ ] Responsividade testada em mobile/tablet/desktop
- [ ] Imagens otimizadas (WebP, lazy loading)
- [ ] Favicon e meta tags configurados

**Impacto Esperado:**
- ✅ Produto visualmente **comparável a Zillow/QuintoAndar** desde o MVP
- ✅ **Retenção de usuários** por UX moderna e fluida
- ✅ **Facilidade de rebranding** quando marca definitiva for criada
- ✅ **Profissionalismo** que inspira confiança em corretores e proprietários

## 23. Otimização Automática de Mídia (Fotos e Vídeos)

### 23.1 Problema Identificado

**Realidade do mercado brasileiro:**
- **Raros imóveis** são fotografados por profissionais
- **Corretores tiram fotos próprias** com qualidade inconsistente:
  - Problemas de iluminação (fotos escuras, superexpostas)
  - Enquadramento incorreto (cortes ruins, ângulos ruins)
  - Falta de sequência lógica (sem "tour" organizado)
  - Ausência de foto de capa atrativa
- **Vídeos para redes sociais** são comuns, mas não aproveitados nas plataformas

**Impacto:**
- Anúncios com fotos ruins **convertem 60% menos** (dados Zillow)
- Usuários saem da plataforma rapidamente se as fotos não impressionam
- Desperdício de conteúdo (vídeos do Instagram/TikTok não usados)

**Oportunidade:**
- **Diferencial competitivo**: otimização automática de fotos via IA
- **Suporte a vídeos**: integração com conteúdo de redes sociais
- **Melhor conversão**: imóveis mais atraentes = mais leads

### 23.2 Solução Proposta - Otimização Automática de Fotos (IA)

**Pipeline de Processamento (Google Cloud):**

```
Upload → Cloud Storage → Cloud Functions → Vision AI → Processamento → GCS Final
```

#### Etapa 1: Upload e Detecção (Cloud Vision API)

```go
// Análise automática da foto
type PhotoAnalysis struct {
    Labels          []string  // "kitchen", "bedroom", "living_room", "bathroom"
    Quality         float64   // 0.0 - 1.0 (brightness, sharpness, composition)
    SafeSearch      bool      // Filtrar conteúdo inapropriado
    HasPeople       bool      // Detectar pessoas (remover por privacidade)
    DominantColors  []string  // Para harmonia visual
    SuggestedOrder  int       // Ordem sugerida no tour
}
```

**Categorização Automática:**
```go
// Cloud Function: analyze-photo
func AnalyzePhoto(ctx context.Context, file *storage.ObjectAttrs) error {
    // 1. Vision API - Label Detection
    labels := visionClient.DetectLabels(file.URL)

    // 2. Classificar tipo de ambiente
    roomType := ClassifyRoom(labels)
    // "living_room" → ordem: 1 (sala primeiro no tour)
    // "kitchen" → ordem: 2
    // "bedroom" → ordem: 3
    // "bathroom" → ordem: 4
    // "exterior" → ordem: 0 (sempre primeiro - fachada)

    // 3. Avaliar qualidade técnica
    quality := EvaluateQuality(file.URL)
    // brightness: muito escuro? muito claro?
    // sharpness: desfocado?
    // composition: horizonte torto?

    // 4. Salvar metadata no Firestore
    SavePhotoMetadata(PhotoAnalysis{
        Labels:         labels,
        Quality:        quality,
        RoomType:       roomType,
        SuggestedOrder: GetRoomOrder(roomType),
    })
}
```

#### Etapa 2: Melhorias Automáticas (opcional - MVP++)

**Opção A: Google Cloud Vision AI + AutoML (treinamento customizado)**
- Treinar modelo para detectar "boa foto de imóvel"
- Sugerir cortes automáticos (crop para destacar ambiente)

**Opção B: Integração com APIs de Terceiros (MVP)**
- **Remove.bg API**: remover objetos indesejados (pessoas, lixo)
- **ImgBB ou Cloudinary**: ajustes automáticos de:
  - Brightness/Contrast (correção de iluminação)
  - Auto-straighten (corrigir horizonte torto)
  - Sharpening (melhorar nitidez)

**Exemplo de Pipeline (Cloud Function):**

```go
// Cloud Function: enhance-photo
func EnhancePhoto(ctx context.Context, file *storage.ObjectAttrs) error {
    // 1. Download da foto original
    originalURL := file.URL

    // 2. Análise de qualidade
    analysis := AnalyzePhoto(originalURL)

    // 3. Aplicar melhorias SE necessário
    if analysis.Quality < 0.6 {
        // Iluminação ruim → ajustar brightness/contrast
        enhanced := AdjustBrightness(originalURL, analysis.Brightness)

        // Horizonte torto → auto-straighten
        if analysis.IsTilted {
            enhanced = StraightenImage(enhanced)
        }

        // Desfocado → tentar sharpening leve
        if analysis.Sharpness < 0.5 {
            enhanced = ApplySharpening(enhanced)
        }

        // Salvar versão melhorada
        SaveEnhancedPhoto(enhanced, "enhanced_" + file.Name)
    }

    // 4. Gerar variantes (thumb, medium, large) - WebP
    GenerateVariants(file.URL)
}
```

#### Etapa 3: Ordenação Inteligente (Tour Virtual)

**Sequência Lógica Sugerida:**
```
1. Fachada/Exterior (primeira impressão)
2. Sala de estar (ambiente principal)
3. Cozinha
4. Quartos (ordem decrescente de tamanho)
5. Banheiros
6. Áreas extras (varanda, quintal, garagem)
```

**Implementação (Backend):**

```go
// internal/services/photo_service.go
func SuggestPhotoOrder(photos []Photo) []Photo {
    // 1. Classificar por tipo de ambiente (Vision AI labels)
    categorized := CategorizePhotos(photos)

    // 2. Ordenar por prioridade
    ordered := []Photo{}
    ordered = append(ordered, categorized["exterior"]...)      // Fachada primeiro
    ordered = append(ordered, categorized["living_room"]...)   // Sala
    ordered = append(ordered, categorized["kitchen"]...)       // Cozinha
    ordered = append(ordered, categorized["bedroom"]...)       // Quartos
    ordered = append(ordered, categorized["bathroom"]...)      // Banheiros
    ordered = append(ordered, categorized["other"]...)         // Resto

    // 3. Atualizar campo `order` de cada foto
    for i, photo := range ordered {
        photo.Order = i
        photo.IsCover = (i == 0) // Primeira foto = capa
    }

    return ordered
}
```

**Interface Admin (PROMPT 04b):**
```tsx
// Corretor pode:
// 1. Ver ordem sugerida pela IA
// 2. Drag-and-drop para reordenar manualmente
// 3. Ver análise de qualidade de cada foto
<PhotoManager>
  <PhotoGrid sortable onReorder={handleReorder}>
    {photos.map((photo, idx) => (
      <PhotoCard
        photo={photo}
        suggestedOrder={photo.suggestedOrder}
        currentOrder={idx}
        quality={photo.quality}
        roomType={photo.roomType}
      />
    ))}
  </PhotoGrid>

  <Button onClick={applyAISuggestions}>
    Aplicar Ordem Sugerida por IA
  </Button>
</PhotoManager>
```

### 23.3 Suporte a Vídeos (Redes Sociais)

**Contexto:**
- Corretores criam vídeos para Instagram/TikTok/YouTube
- Vídeos **aumentam conversão em 80%** (dados Redfin)
- Necessário permitir upload e exibição na plataforma

**Modelo de Dados Atualizado:**

```go
// Adicionar ao Listing model (PROMPT 01)
type Listing struct {
    // ... campos existentes

    Photos []Photo `firestore:"photos" json:"photos"`

    // 🆕 Suporte a vídeos
    Videos []Video `firestore:"videos" json:"videos"`
}

type Video struct {
    ID          string    `firestore:"id" json:"id"`
    URL         string    `firestore:"url" json:"url"` // GCS URL
    ThumbnailURL string   `firestore:"thumbnail_url" json:"thumbnail_url"` // Frame do meio
    Duration    int       `firestore:"duration" json:"duration"` // segundos
    Source      string    `firestore:"source,omitempty" json:"source,omitempty"` // "upload", "youtube", "instagram"
    SourceURL   string    `firestore:"source_url,omitempty" json:"source_url,omitempty"` // URL original (se externo)
    Order       int       `firestore:"order" json:"order"`
    CreatedAt   time.Time `firestore:"created_at" json:"created_at"`
}
```

**Opções de Vídeo:**

#### Opção 1: Upload Direto (MVP)

```go
// Cloud Storage + Cloud Run (ffmpeg)
func ProcessVideo(ctx context.Context, file *storage.ObjectAttrs) error {
    // 1. Validar tamanho (max 500MB) e formato (mp4, mov)
    if file.Size > 500*1024*1024 {
        return errors.New("video too large")
    }

    // 2. Gerar thumbnail (frame do meio)
    thumbnail := GenerateThumbnail(file.URL, file.Duration/2)
    SaveThumbnail(thumbnail, "thumb_" + file.Name)

    // 3. Comprimir vídeo (se necessário)
    // H.264, 1080p max, bitrate otimizado
    if NeedsCompression(file) {
        compressed := CompressVideo(file.URL)
        ReplaceOriginal(compressed)
    }

    // 4. Extrair duração
    duration := GetVideoDuration(file.URL)

    // 5. Salvar metadata no Firestore
    SaveVideoMetadata(Video{
        URL:          file.URL,
        ThumbnailURL: thumbnail,
        Duration:     duration,
        Source:       "upload",
    })
}
```

#### Opção 2: Integração com YouTube/Instagram (MVP++)

```tsx
// Frontend Admin - Adicionar vídeo
<VideoUploader>
  <Tabs>
    <Tab label="Upload">
      <input type="file" accept="video/mp4,video/quicktime" />
    </Tab>

    <Tab label="YouTube">
      <Input
        placeholder="Cole o link do YouTube..."
        onChange={handleYouTubeLink}
      />
      {/* Embed via iframe */}
    </Tab>

    <Tab label="Instagram">
      <Input
        placeholder="Cole o link do Reels/IGTV..."
        onChange={handleInstagramLink}
      />
      {/* Embed via oEmbed API */}
    </Tab>
  </Tabs>
</VideoUploader>
```

**Exibição no Frontend Público:**

```tsx
// Página do imóvel - Galeria com fotos + vídeos
<PropertyGallery>
  {/* Carrossel combinado */}
  <Carousel>
    {/* Fotos */}
    {photos.map(photo => (
      <Image src={photo.largeURL} />
    ))}

    {/* Vídeos */}
    {videos.map(video => (
      video.source === "upload" ? (
        <video controls poster={video.thumbnailURL}>
          <source src={video.url} type="video/mp4" />
        </video>
      ) : (
        <iframe src={video.sourceURL} /> // YouTube/Instagram
      )
    ))}
  </Carousel>
</PropertyGallery>
```

### 23.4 Custos e ROI

**Custos Google Cloud (estimativa mensal para 1.000 imóveis):**

| Serviço | Uso | Custo Mensal |
|---------|-----|--------------|
| **Cloud Storage** | 10GB fotos + 50GB vídeos | ~$1.50 |
| **Vision API** | 10.000 análises/mês | $15.00 |
| **Cloud Functions** | 50.000 execuções | $0.50 |
| **Cloudinary/ImgBB** (opcional) | 5.000 enhancements | $25.00 |
| **TOTAL MVP (sem enhancement)** | - | **$17/mês** |
| **TOTAL MVP++ (com enhancement)** | - | **$42/mês** |

**ROI Esperado:**
- **Conversão de leads**: +40-60% (fotos de qualidade)
- **Tempo de venda**: -20% (imóveis mais atraentes)
- **Satisfação do corretor**: alta (menos trabalho manual)
- **Diferencial competitivo**: único no mercado brasileiro

### 23.5 Implementação Faseada

**Fase 1 (MVP - Incluir AGORA):**
- ✅ Suporte a múltiplas fotos (já existe)
- ✅ Suporte a vídeos (adicionar model Video)
- ✅ Upload direto de vídeos (GCS)
- ✅ Thumbnail automático de vídeos (ffmpeg)
- ✅ Análise básica de fotos (Vision API - labels)
- ✅ Ordenação manual (drag-and-drop admin)

**Fase 2 (MVP+ - 2-4 semanas após MVP):**
- 🔲 Ordenação inteligente sugerida (IA)
- 🔲 Análise de qualidade técnica (brightness, sharpness)
- 🔲 Integração YouTube/Instagram (embed)
- 🔲 Feedback visual de qualidade no admin ("Foto escura - melhorar iluminação")

**Fase 3 (MVP++ - 1-3 meses após MVP):**
- 🔲 Enhancement automático (brightness/contrast/straighten)
- 🔲 Remoção de objetos indesejados (pessoas, lixo)
- 🔲 Sugestão de foto de capa (melhor foto por IA)
- 🔲 Análise de composição (horizonte torto, cortes ruins)
- 🔲 AutoML treinado para imóveis brasileiros

### 23.6 Atualização de Prompts

**PROMPT 01 (Foundation MVP):**
- ✅ Adicionar `Videos []Video` ao Listing model
- ✅ Adicionar struct `Video` completo

**PROMPT 02 (Import + Deduplication):**
- ✅ Pipeline de processamento de fotos (Vision API)
- ✅ Pipeline de processamento de vídeos (ffmpeg)
- ✅ Cloud Functions para análise automática

**PROMPT 04b (Frontend Admin):**
- ✅ Upload de múltiplos vídeos
- ✅ Drag-and-drop para ordenar fotos/vídeos
- ✅ Preview de vídeos
- ✅ Indicador de qualidade de fotos (Fase 2)

**PROMPT 04 (Frontend Public):**
- ✅ Carrossel combinado (fotos + vídeos)
- ✅ Player de vídeo (HTML5 + fallback)
- ✅ Lazy loading de vídeos

### 23.7 Exemplo Completo - Upload de Vídeo

**Backend (Go):**

```go
// POST /api/v1/tenants/:tenantId/listings/:listingId/videos
func (h *ListingHandler) UploadVideo(c *gin.Context) {
    file, _ := c.FormFile("video")

    // 1. Validar
    if file.Size > 500*1024*1024 {
        c.JSON(400, gin.H{"error": "Video too large (max 500MB)"})
        return
    }

    // 2. Upload para GCS
    videoURL := h.storage.Upload(file, "videos/")

    // 3. Processar em background (Cloud Function)
    h.pubsub.Publish("video-processing", videoURL)

    // 4. Criar Video record (thumbnail será adicionado depois)
    video := &Video{
        ID:        uuid.New().String(),
        URL:       videoURL,
        Source:    "upload",
        Order:     len(listing.Videos), // Último
        CreatedAt: time.Now(),
    }

    // 5. Adicionar ao Listing
    listing.Videos = append(listing.Videos, video)
    h.listingRepo.Update(c, listing)

    c.JSON(200, video)
}
```

**Cloud Function (Processamento):**

```go
// Cloud Function: process-video
func ProcessVideo(ctx context.Context, m pubsub.Message) error {
    videoURL := string(m.Data)

    // 1. Download temporário
    tmpFile := DownloadToTemp(videoURL)

    // 2. Gerar thumbnail (frame do meio)
    cmd := exec.Command("ffmpeg",
        "-i", tmpFile,
        "-ss", "00:00:05", // 5 segundos (ou metade do vídeo)
        "-vframes", "1",
        "-vf", "scale=800:-1",
        "thumb.jpg",
    )
    cmd.Run()

    // 3. Upload thumbnail
    thumbURL := UploadToGCS("thumb.jpg", "thumbnails/")

    // 4. Extrair duração
    duration := GetDuration(tmpFile) // ffprobe

    // 5. Atualizar Firestore
    UpdateVideoMetadata(videoURL, Video{
        ThumbnailURL: thumbURL,
        Duration:     duration,
    })

    return nil
}
```

**Frontend Admin:**

```tsx
<VideoUpload>
  <input
    type="file"
    accept="video/mp4,video/quicktime"
    onChange={async (e) => {
      const file = e.target.files[0]

      // Validar tamanho
      if (file.size > 500 * 1024 * 1024) {
        toast.error("Vídeo muito grande (máx 500MB)")
        return
      }

      // Upload com progresso
      const { data } = await api.post(
        `/listings/${listingId}/videos`,
        { video: file },
        {
          onUploadProgress: (e) => {
            setProgress(Math.round((e.loaded * 100) / e.total))
          }
        }
      )

      toast.success("Vídeo enviado! Processamento em andamento...")
    }}
  />

  {progress > 0 && (
    <Progress value={progress} />
  )}
</VideoUpload>
```

### 23.8 Diferencial Competitivo

**Nenhuma plataforma brasileira faz isso bem:**
- **Zap Imóveis**: aceita fotos ruins sem aviso
- **OLX**: sem qualquer análise de qualidade
- **QuintoAndar**: exige fotos profissionais (barreira de entrada)

**Nossa plataforma:**
- ✅ Aceita fotos amadoras (baixa barreira)
- ✅ **Melhora automaticamente** (IA)
- ✅ **Sugere ordenação** (tour lógico)
- ✅ **Suporta vídeos** (redes sociais)
- ✅ **Feedback educativo** ao corretor ("essa foto está escura")

**Resultado:**
- **Corretores amam**: menos trabalho, melhores resultados
- **Clientes amam**: imóveis mais bonitos, tour organizado
- **Plataforma cresce**: diferencial claro vs concorrentes

## 24. Whitelabel (Branding Personalizado por Tenant)

### 24.1 Conceito de Whitelabel

**Whitelabel** permite que cada tenant (imobiliária) tenha sua **própria identidade visual** na plataforma, incluindo:
- Logo personalizado
- Cores da marca (paleta completa)
- Nome da empresa
- Domínio customizado (opcional)
- Informações de contato

**Objetivo**: Cada imobiliária sente que possui **sua própria plataforma**, não uma plataforma compartilhada.

### 24.2 Status Atual (MVP)

**Arquitetura Multi-Tenant: ✅ 100%**
- Isolamento completo de dados por `tenant_id`
- Firestore Security Rules impedem vazamento cross-tenant
- `Tenant.settings` é flexível (`map[string]interface{}`)

**Branding Dinâmico: ⚠️ 40%**
- ❌ Frontend usa variáveis de ambiente globais (`NEXT_PUBLIC_APP_NAME`)
- ❌ Cores hardcoded em `tailwind.config.ts`
- ❌ Logo sempre renderiza placeholder genérico
- ⚠️ Meta tags usam `tenant.name` (parcial)

**Conclusão**: Fundação sólida, mas **branding ainda não é dinâmico**.

### 24.3 Campos de Branding no Tenant.settings

**Campos obrigatórios para whitelabel completo**:

```go
type TenantSettings struct {
    // Contato
    WhatsAppDefault string `json:"whatsapp_default"`
    ContactEmail    string `json:"contact_email"`
    ContactPhone    string `json:"contact_phone"`

    // Branding Visual
    BusinessName    string `json:"business_name"`      // "Imobiliária Primavera"
    Tagline         string `json:"tagline,omitempty"`  // "Seu lar dos sonhos"
    LogoURL         string `json:"logo_url"`           // GCS URL
    FaviconURL      string `json:"favicon_url"`        // GCS URL

    // Paleta de Cores (hex)
    PrimaryColor    string `json:"primary_color"`      // "#0066FF"
    SecondaryColor  string `json:"secondary_color"`    // "#E8EAED"
    AccentColor     string `json:"accent_color"`       // "#22C55E"

    // Domínio Customizado (Opcional - MVP+2)
    CustomDomain    string `json:"custom_domain,omitempty"` // "minhaimobiliaria.com.br"

    // SEO
    MetaDescription string `json:"meta_description,omitempty"`
    MetaKeywords    string `json:"meta_keywords,omitempty"`
}
```

### 24.4 Frontend Dinâmico

**Problema Atual**:
```typescript
// ❌ ATUAL: Branding estático
export const branding = {
  name: process.env.NEXT_PUBLIC_APP_NAME || "ImóvelHub",
  logo: process.env.NEXT_PUBLIC_LOGO_URL || "/logo.svg",
  primaryColor: "#0066FF" // Hard-coded
}
```

**Solução Whitelabel**:
```typescript
// ✅ WHITELABEL: Branding dinâmico por tenant
export function useBranding() {
  const { tenantId } = useAuth()

  const { data: tenant } = useQuery({
    queryKey: ['tenant', tenantId],
    queryFn: () => api.get(`/tenants/${tenantId}`)
  })

  return {
    name: tenant?.settings?.business_name || "ImóvelHub",
    logo: tenant?.settings?.logo_url || "/logo-placeholder.svg",
    primaryColor: tenant?.settings?.primary_color || "#0066FF",
    secondaryColor: tenant?.settings?.secondary_color || "#E8EAED",
    accentColor: tenant?.settings?.accent_color || "#22C55E",
    contactEmail: tenant?.settings?.contact_email,
    contactPhone: tenant?.settings?.contact_phone,
  }
}
```

### 24.5 CSS Variables Dinâmicas

**Implementação em layout.tsx**:

```tsx
// app/layout.tsx
export default async function RootLayout({ children }) {
  const tenant = await fetchTenant() // SSR

  const style = {
    '--color-primary': tenant.settings?.primary_color || '#0066FF',
    '--color-secondary': tenant.settings?.secondary_color || '#E8EAED',
    '--color-accent': tenant.settings?.accent_color || '#22C55E',
  }

  return (
    <html lang="pt-BR" style={style}>
      <head>
        <link rel="icon" href={tenant.settings?.favicon_url || "/favicon.svg"} />
      </head>
      <body className="font-sans">
        {children}
      </body>
    </html>
  )
}
```

**Tailwind configurado para CSS Variables**:

```typescript
// tailwind.config.ts
export default {
  theme: {
    extend: {
      colors: {
        primary: 'var(--color-primary)',     // ✅ Dinâmico
        secondary: 'var(--color-secondary)', // ✅ Dinâmico
        accent: 'var(--color-accent)',       // ✅ Dinâmico
      },
    },
  },
}
```

### 24.6 Upload de Logo (Backend)

**Endpoint para upload de logo**:

```go
// POST /api/v1/tenants/{tenantId}/logo
func (h *TenantHandler) UploadLogo(c *gin.Context) {
    tenantID := c.Param("tenantId")

    // 1. Validar permissão (apenas admin do tenant)
    if !h.isAdmin(c, tenantID) {
        c.JSON(403, gin.H{"error": "Forbidden"})
        return
    }

    // 2. Upload para GCS
    file, _ := c.FormFile("logo")
    gcsPath := fmt.Sprintf("tenants/%s/branding/logo.png", tenantID)
    logoURL, err := h.storage.Upload(c, gcsPath, file)

    // 3. Atualizar Tenant.settings.logo_url
    err = h.db.Collection("tenants").Doc(tenantID).Update(c, []firestore.Update{
        {Path: "settings.logo_url", Value: logoURL},
        {Path: "updated_at", Value: time.Now()},
    })

    c.JSON(200, gin.H{"logo_url": logoURL})
}
```

### 24.7 UI de Configuração de Branding (Admin)

**Página: /app/configuracoes/branding**

```tsx
'use client'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

const brandingSchema = z.object({
  business_name: z.string().min(3, 'Mínimo 3 caracteres'),
  tagline: z.string().optional(),
  primary_color: z.string().regex(/^#[0-9A-F]{6}$/i, 'Cor inválida'),
  secondary_color: z.string().regex(/^#[0-9A-F]{6}$/i, 'Cor inválida'),
  accent_color: z.string().regex(/^#[0-9A-F]{6}$/i, 'Cor inválida'),
  contact_email: z.string().email(),
  contact_phone: z.string().min(10),
})

export default function BrandingSettingsPage() {
  const { tenantId } = useAuth()
  const queryClient = useQueryClient()

  const { data: tenant } = useQuery({
    queryKey: ['tenant', tenantId],
    queryFn: () => api.get(`/tenants/${tenantId}`)
  })

  const form = useForm({
    resolver: zodResolver(brandingSchema),
    defaultValues: tenant?.settings || {}
  })

  const updateBranding = useMutation({
    mutationFn: (data) => api.patch(`/tenants/${tenantId}`, { settings: data }),
    onSuccess: () => {
      queryClient.invalidateQueries(['tenant', tenantId])
      toast.success('Branding atualizado!')
    }
  })

  const uploadLogo = useMutation({
    mutationFn: (file: File) => {
      const formData = new FormData()
      formData.append('logo', file)
      return api.post(`/tenants/${tenantId}/logo`, formData)
    },
    onSuccess: (data) => {
      form.setValue('logo_url', data.logo_url)
      toast.success('Logo atualizado!')
    }
  })

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-3xl font-bold mb-6">Configuração de Marca</h1>

      <form onSubmit={form.handleSubmit(updateBranding.mutate)} className="space-y-6">
        {/* Logo Upload */}
        <div>
          <Label>Logo da Empresa</Label>
          <div className="mt-2 flex items-center gap-4">
            {tenant?.settings?.logo_url && (
              <img src={tenant.settings.logo_url} alt="Logo" className="h-16 w-16 object-contain" />
            )}
            <Input
              type="file"
              accept="image/*"
              onChange={(e) => {
                const file = e.target.files?.[0]
                if (file) uploadLogo.mutate(file)
              }}
            />
          </div>
          <p className="text-sm text-gray-500 mt-1">
            PNG ou SVG recomendado. Tamanho ideal: 512x512px
          </p>
        </div>

        {/* Nome da Empresa */}
        <div>
          <Label htmlFor="business_name">Nome da Imobiliária</Label>
          <Input id="business_name" {...form.register('business_name')} />
          {form.formState.errors.business_name && (
            <p className="text-sm text-red-600">{form.formState.errors.business_name.message}</p>
          )}
        </div>

        {/* Tagline */}
        <div>
          <Label htmlFor="tagline">Slogan (opcional)</Label>
          <Input id="tagline" {...form.register('tagline')} placeholder="Seu lar dos sonhos" />
        </div>

        {/* Cores */}
        <div className="grid grid-cols-3 gap-4">
          <div>
            <Label htmlFor="primary_color">Cor Primária</Label>
            <div className="flex gap-2">
              <Input
                type="color"
                id="primary_color"
                {...form.register('primary_color')}
                className="w-16 h-10"
              />
              <Input {...form.register('primary_color')} placeholder="#0066FF" />
            </div>
          </div>

          <div>
            <Label htmlFor="secondary_color">Cor Secundária</Label>
            <div className="flex gap-2">
              <Input
                type="color"
                id="secondary_color"
                {...form.register('secondary_color')}
                className="w-16 h-10"
              />
              <Input {...form.register('secondary_color')} placeholder="#E8EAED" />
            </div>
          </div>

          <div>
            <Label htmlFor="accent_color">Cor de Destaque</Label>
            <div className="flex gap-2">
              <Input
                type="color"
                id="accent_color"
                {...form.register('accent_color')}
                className="w-16 h-10"
              />
              <Input {...form.register('accent_color')} placeholder="#22C55E" />
            </div>
          </div>
        </div>

        {/* Contato */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <Label htmlFor="contact_email">Email de Contato</Label>
            <Input type="email" id="contact_email" {...form.register('contact_email')} />
          </div>

          <div>
            <Label htmlFor="contact_phone">Telefone de Contato</Label>
            <Input id="contact_phone" {...form.register('contact_phone')} placeholder="+55 11 99999-9999" />
          </div>
        </div>

        {/* Preview */}
        <div className="p-4 border rounded-lg bg-gray-50">
          <p className="text-sm font-medium mb-2">Preview:</p>
          <div className="flex items-center gap-3 p-4 bg-white rounded" style={{
            backgroundColor: 'white',
            borderLeft: `4px solid ${form.watch('primary_color') || '#0066FF'}`
          }}>
            {tenant?.settings?.logo_url && (
              <img src={tenant.settings.logo_url} alt="Logo" className="h-10" />
            )}
            <div>
              <p className="font-bold">{form.watch('business_name') || 'Sua Imobiliária'}</p>
              <p className="text-sm text-gray-600">{form.watch('tagline') || 'Seu slogan aqui'}</p>
            </div>
          </div>
        </div>

        <Button type="submit" size="lg" className="w-full" disabled={updateBranding.isPending}>
          {updateBranding.isPending ? 'Salvando...' : 'Salvar Alterações'}
        </Button>
      </form>
    </div>
  )
}
```

### 24.8 Domínios Customizados (MVP+2 - Opcional)

**Conceito**: Cada tenant pode ter seu próprio domínio (ex: `imobiliariaprimavera.com.br`).

**Backend - Middleware**:
```go
func TenantDomainMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        host := c.Request.Host

        // 1. Cache: verificar se domínio está em cache
        if tenantID, ok := domainCache.Get(host); ok {
            c.Set("tenant_id", tenantID)
            c.Next()
            return
        }

        // 2. Query: buscar tenant por custom_domain
        var tenant Tenant
        err := db.Collection("tenants").
            Where("settings.custom_domain", "==", host).
            Limit(1).
            Documents(c).
            Next().
            DataTo(&tenant)

        if err == nil {
            // 3. Cache: armazenar por 5 minutos
            domainCache.Set(host, tenant.ID, 5*time.Minute)
            c.Set("tenant_id", tenant.ID)
        }

        c.Next()
    }
}
```

**Frontend - Next.js Middleware**:
```typescript
// middleware.ts
import { NextRequest, NextResponse } from 'next/server'

export async function middleware(request: NextRequest) {
  const host = request.headers.get('host') || ''

  // 1. Verificar se é domínio customizado
  if (!host.includes('imovelhub.com')) {
    // 2. Fetch tenant by domain (com cache)
    const tenant = await fetchTenantByDomain(host)

    if (tenant) {
      // 3. Inject tenant_id no header
      const headers = new Headers(request.headers)
      headers.set('x-tenant-id', tenant.id)

      return NextResponse.rewrite(request.url, { headers })
    }
  }

  return NextResponse.next()
}
```

**Configuração DNS**:
```
# Cliente configura em seu provedor de DNS:
CNAME  www.imobiliariaprimavera.com.br  →  cname.vercel-dns.com
```

**Vercel**:
```bash
# Adicionar domínio via Vercel CLI
vercel domains add imobiliariaprimavera.com.br --project=imovel-hub
```

### 24.9 Estimativa de Esforço

**MVP+1 (Whitelabel Básico) - 13 horas**:
- ✅ Expandir Tenant.settings com campos de branding (3h)
- ✅ Refatorar frontend para `useBranding()` hook (5h)
- ✅ CSS Variables dinâmicas (4h)
- ✅ UI de configuração /app/configuracoes/branding (8h)
- ✅ Endpoint de upload de logo (3h)

**MVP+2 (Domínios Customizados) - 10 horas**:
- TenantDomainMiddleware (backend) - 4h
- Next.js middleware (frontend) - 3h
- Configuração Vercel multi-domain - 2h
- Documentação DNS - 1h

**Total**: 23 horas (~3 dias de desenvolvimento)

### 24.10 Benefícios de Whitelabel

**Para os Tenants (Imobiliárias)**:
- ✅ Identidade visual própria (não parece "plataforma compartilhada")
- ✅ Fortalecimento da marca
- ✅ Domínio customizado (opcional) aumenta credibilidade
- ✅ SEO independente por domínio

**Para a Plataforma**:
- ✅ Precificação premium (+30-50% por whitelabel)
- ✅ Redução de churn (cliente se sente "dono")
- ✅ Diferencial competitivo forte
- ✅ Facilita onboarding (5 minutos para configurar marca)

**ROI Estimado**:
- Investimento: R$ 2.300 (23h × R$ 100/h)
- Retorno: 10 tenants × R$ 500/mês (whitelabel premium) = R$ 5.000/mês
- ROI: 2.2x no primeiro mês, 26x no primeiro ano

### 24.11 Checklist de Implementação

**Backend**:
- [ ] Expandir `Tenant.settings` com 10+ campos de branding
- [ ] Criar endpoint `POST /api/v1/tenants/{id}/logo`
- [ ] Criar endpoint `PATCH /api/v1/tenants/{id}` (já existe, validar campos)
- [ ] Middleware `TenantDomainMiddleware` (MVP+2)
- [ ] Query otimizada `FindTenantByDomain()` (MVP+2)

**Frontend**:
- [ ] Criar hook `useBranding()` dinâmico
- [ ] Refatorar todos os componentes que usam `lib/branding.ts` estático
- [ ] Implementar CSS Variables em `app/layout.tsx`
- [ ] Atualizar `tailwind.config.ts` para usar variáveis
- [ ] Criar página `/app/configuracoes/branding`
- [ ] Componente `ColorPicker`
- [ ] Upload de logo com preview
- [ ] Next.js middleware para domínios customizados (MVP+2)

**Infraestrutura**:
- [ ] Configurar Vercel para multi-domain (MVP+2)
- [ ] Documentar processo de configuração DNS
- [ ] CDN para logos (GCS já suporta)

**QA**:
- [ ] Testes end-to-end de branding dinâmico
- [ ] Validar CSS Variables em diferentes browsers
- [ ] Testar upload de logo (PNG, SVG, JPEG)
- [ ] Validar domínio customizado (MVP+2)

## 25. Conclusão
MVP sólido, governável, multi-tenant, **LGPD-compliant**, com **design moderno**, **otimização de mídia por IA**, **SEO 100%**, e **preparado para whitelabel** no ecossistema Google Cloud.
