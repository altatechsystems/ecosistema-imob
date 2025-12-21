# 🔥 Guia Completo de Configuração Firebase

> **Projeto:** ecosistema-imob-dev
> **Status:** Projeto criado ✅ | Configuração em andamento ⏳

---

## ✅ Passo 1: Habilitar Firestore Database

1. No console Firebase (onde você está agora), clique em **"Criação"** no menu lateral esquerdo
2. Clique em **"Firestore Database"**
3. Clique em **"Criar banco de dados"**
4. Escolha **"Iniciar no modo de produção"** (vamos configurar regras depois)
5. Selecione localização: **"us-east1"** (ou mais próxima do Brasil: "southamerica-east1")
6. Clique em **"Ativar"**

**Aguarde 1-2 minutos** até o Firestore ser provisionado.

---

## ✅ Passo 2: Habilitar Authentication

1. No menu lateral, clique em **"Criação"** → **"Authentication"**
2. Clique em **"Começar"**
3. Na aba **"Sign-in method"**, habilite:
   - ✅ **E-mail/senha** (clique em "Ativar" e salve)
   - ✅ **Google** (clique em "Ativar", aceite os defaults, salve)

---

## ✅ Passo 3: Habilitar Cloud Storage

1. No menu lateral, clique em **"Criação"** → **"Storage"**
2. Clique em **"Começar"**
3. Escolha **"Iniciar no modo de produção"**
4. Selecione a mesma localização do Firestore
5. Clique em **"Concluído"**

---

## ✅ Passo 4: Baixar Credenciais do Admin SDK (Backend Go)

1. No console Firebase, clique no ⚙️ (engrenagem) ao lado de "Visão geral do projeto"
2. Clique em **"Configurações do projeto"**
3. Vá para a aba **"Contas de serviço"**
4. Certifique-se de estar em **"Firebase Admin SDK"**
5. Clique em **"Gerar nova chave privada"**
6. Confirme clicando em **"Gerar chave"**
7. Um arquivo JSON será baixado (exemplo: `ecosistema-imob-dev-firebase-adminsdk-xxxxx.json`)

**IMPORTANTE:**
- Renomeie o arquivo para: `firebase-adminsdk.json`
- Mova para: `c:\Users\danie\OneDrive\Documentos\Altatech Systems\ecosystem\ecosistema-imob\backend\config\firebase-adminsdk.json`
- Você precisa criar a pasta `config` primeiro:

```bash
mkdir c:\Users\danie\OneDrive\Documentos\Altatech Systems\ecosystem\ecosistema-imob\backend\config
```

⚠️ **NUNCA COMMITE ESTE ARQUIVO NO GIT!** (já está no .gitignore)

---

## ✅ Passo 5: Configurar Web App (Frontend Next.js)

1. Na mesma página de "Configurações do projeto", clique na aba **"Geral"**
2. Role para baixo até **"Seus apps"**
3. Clique no ícone **</> (Web)**
4. Preencha:
   - Nome do app: **"ecosistema-imob-public"**
   - ✅ Marque: "Também configurar o Firebase Hosting para este app" (opcional)
5. Clique em **"Registrar app"**
6. Copie o objeto `firebaseConfig` que aparece. Será algo assim:

```javascript
const firebaseConfig = {
  apiKey: "AIzaSyXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
  authDomain: "ecosistema-imob-dev.firebaseapp.com",
  projectId: "ecosistema-imob-dev",
  storageBucket: "ecosistema-imob-dev.firebasestorage.app",
  messagingSenderId: "123456789012",
  appId: "1:123456789012:web:xxxxxxxxxxxxx"
};
```

7. **Salve essas informações** - vamos usar no próximo passo

---

## ✅ Passo 6: Criar Arquivo .env para Backend

Crie o arquivo: `c:\Users\danie\OneDrive\Documentos\Altatech Systems\ecosystem\ecosistema-imob\backend\.env`

```env
# Firebase
GOOGLE_APPLICATION_CREDENTIALS=./config/firebase-adminsdk.json
FIREBASE_PROJECT_ID=ecosistema-imob-dev

# Server
PORT=8080
GIN_MODE=debug
ENVIRONMENT=development

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,http://localhost:3002

# Cloud Storage (use o storageBucket do firebaseConfig acima)
GCS_BUCKET_NAME=ecosistema-imob-dev.firebasestorage.app

# Logging
LOG_LEVEL=info
```

**Substitua** `ecosistema-imob-dev.firebasestorage.app` pelo valor real do `storageBucket` que você copiou no passo 5.

---

## ✅ Passo 7: Instalar Firebase CLI e Deploy Índices

Abra o terminal (PowerShell ou CMD) e execute:

```bash
# Instalar Firebase CLI globalmente
npm install -g firebase-tools

# Login no Firebase
firebase login

# Ir para a pasta do projeto
cd "c:\Users\danie\OneDrive\Documentos\Altatech Systems\ecosystem\ecosistema-imob"

# Inicializar Firebase (apenas Firestore)
firebase init firestore

# Seleções durante o init:
# - Use an existing project → ecosistema-imob-dev
# - Firestore rules file → firestore.rules (aceite default)
# - Firestore indexes file → firestore.indexes.json (aceite default)

# Deploy dos índices
firebase deploy --only firestore:indexes
```

**Aguarde** 2-5 minutos para todos os 56 índices serem criados.

---

## ✅ Passo 8: Verificar Configuração

### Backend

```bash
cd backend

# Verificar se as credenciais existem
dir config\firebase-adminsdk.json

# Se existir, você verá:
# Mode                 LastWriteTime         Length Name
# ----                 -------------         ------ ----
# -a----        21/12/2025     XX:XX           XXXX firebase-adminsdk.json
```

### Firestore

1. Volte ao console Firebase
2. Clique em **"Firestore Database"**
3. Você deve ver o banco vazio, pronto para uso
4. Clique na aba **"Índices"**
5. Após o deploy, você verá 56 índices compostos listados

### Storage

1. No console Firebase, clique em **"Storage"**
2. Você verá o bucket vazio, pronto para upload de imagens

---

## 📋 Checklist Final

Antes de continuar a implementação, confirme:

- [ ] ✅ Firestore Database criado e ativo
- [ ] ✅ Authentication habilitado (Email/Password + Google)
- [ ] ✅ Cloud Storage habilitado
- [ ] ✅ Arquivo `backend/config/firebase-adminsdk.json` baixado e salvo
- [ ] ✅ Arquivo `backend/.env` criado com configurações corretas
- [ ] ✅ Firebase CLI instalado (`firebase --version` funciona)
- [ ] ✅ Logged in no Firebase (`firebase login` feito)
- [ ] ✅ Projeto inicializado (`firebase init firestore` feito)
- [ ] ✅ Índices deployados (`firebase deploy --only firestore:indexes` feito)
- [ ] ✅ 56 índices visíveis na aba "Índices" do Firestore

---

## 🚀 Próximos Passos

Após completar todos os itens acima, você estará pronto para:

1. **Executar Prompt 02** - Backend API MVP (repositories, services, handlers)
2. **Executar Prompt 09** - Next.js 14 SEO Setup (frontend público)
3. **Testar integração** - Backend conectando ao Firestore

---

## 🆘 Troubleshooting

### Erro: "Permission denied" no firebase deploy

```bash
# Fazer logout e login novamente
firebase logout
firebase login
```

### Erro: "Project not found"

```bash
# Listar projetos disponíveis
firebase projects:list

# Selecionar o projeto correto
firebase use ecosistema-imob-dev
```

### Erro: "firebase-adminsdk.json not found"

- Verifique se o arquivo está em: `backend/config/firebase-adminsdk.json`
- Verifique se o caminho no `.env` está correto: `./config/firebase-adminsdk.json`

---

**Última Atualização:** 2025-12-21
**Status:** Aguardando conclusão dos 8 passos acima
