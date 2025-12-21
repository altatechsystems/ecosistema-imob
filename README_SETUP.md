# 🚀 Setup do Ambiente - Ecosistema Imob

> **Status:** Iniciando implementação MVP
> **Data:** 2025-12-21

---

## ✅ Estrutura do Projeto Criada

```
ecosistema-imob/
├── backend/
│   ├── cmd/api/              # Ponto de entrada da aplicação
│   ├── internal/
│   │   ├── handlers/         # HTTP handlers (controllers)
│   │   ├── models/           # Structs Go dos modelos
│   │   ├── repository/       # Acesso a dados (Firestore)
│   │   ├── middleware/       # Auth, CORS, logging
│   │   └── utils/            # Validações, helpers
│   └── config/               # Configuração (Firebase, etc)
├── frontend-public/          # Portal público Next.js
├── frontend-admin-sales/     # Dashboard vendas Next.js
├── frontend-admin-rentals/   # Dashboard locação Next.js (MVP+4)
├── scripts/                  # Scripts utilitários
├── docs/                     # Documentação
├── prompts/                  # Prompts de implementação
└── firestore.indexes.json    # Índices Firestore
```

---

## 📋 Próximos Passos

### 1. Configurar Firebase (Manual)

Antes de continuar, você precisa:

1. **Criar Projeto Firebase:**
   - Acesse: https://console.firebase.google.com
   - Criar novo projeto: "ecosistema-imob-dev"
   - Habilitar Google Analytics (opcional)

2. **Habilitar Serviços:**
   - ✅ Firestore Database (modo produção)
   - ✅ Authentication (Email/Password + Google)
   - ✅ Cloud Storage
   - ✅ Hosting (opcional)

3. **Obter Credenciais:**
   - No console Firebase > Project Settings > Service Accounts
   - Gerar nova chave privada (JSON)
   - Salvar como: `backend/config/firebase-adminsdk.json`
   - **NUNCA commitar este arquivo!** (já está no .gitignore)

4. **Configurar Web App:**
   - No console Firebase > Project Settings > Your apps
   - Adicionar app Web
   - Copiar Firebase Config (apiKey, authDomain, etc)
   - Salvar para uso nos frontends

### 2. Deploy Índices Firestore

```bash
# Instalar Firebase CLI (se ainda não tiver)
npm install -g firebase-tools

# Login
firebase login

# Inicializar projeto
firebase init firestore

# Deploy índices
firebase deploy --only firestore:indexes
```

### 3. Configurar Go Backend

```bash
cd backend

# Inicializar módulo Go
go mod init github.com/altatech/ecosistema-imob

# Instalar dependências principais
go get github.com/gin-gonic/gin
go get firebase.google.com/go/v4
go get cloud.google.com/go/firestore
go get google.golang.org/api/option
```

### 4. Configurar Frontend Public (Next.js)

```bash
cd frontend-public

# Criar projeto Next.js 14
npx create-next-app@latest . --typescript --tailwind --app --src-dir --import-alias "@/*"

# Instalar dependências Firebase
npm install firebase

# Instalar dependências SEO
npm install next-sitemap
```

---

## 🔑 Variáveis de Ambiente

### Backend (.env)

Criar arquivo `backend/.env`:

```env
# Firebase
GOOGLE_APPLICATION_CREDENTIALS=./config/firebase-adminsdk.json
FIREBASE_PROJECT_ID=ecosistema-imob-dev

# Server
PORT=8080
GIN_MODE=debug

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,http://localhost:3002
```

### Frontend Public (.env.local)

Criar arquivo `frontend-public/.env.local`:

```env
# Firebase Config (obter do console Firebase)
NEXT_PUBLIC_FIREBASE_API_KEY=your_api_key
NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=your_auth_domain
NEXT_PUBLIC_FIREBASE_PROJECT_ID=your_project_id
NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET=your_storage_bucket
NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID=your_sender_id
NEXT_PUBLIC_FIREBASE_APP_ID=your_app_id

# API
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## 🧪 Verificar Setup

Após configurar tudo:

```bash
# Backend
cd backend
go run cmd/api/main.go
# Deve iniciar em: http://localhost:8080

# Frontend
cd frontend-public
npm run dev
# Deve iniciar em: http://localhost:3000
```

---

## 📚 Documentação de Referência

- **Arquitetura:** [AI_DEV_DIRECTIVE.md](AI_DEV_DIRECTIVE.md)
- **Prompts:** [prompts/01_foundation_mvp.txt](prompts/01_foundation_mvp.txt)
- **Plano de Execução:** [docs/PLANO_EXECUCAO.md](docs/PLANO_EXECUCAO.md)

---

**Próximo Passo:** Executar [prompts/01_foundation_mvp.txt](prompts/01_foundation_mvp.txt) para criar os modelos de dados
