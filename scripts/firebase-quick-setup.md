# 🚀 Setup Rápido Firebase - Lista de Verificação

## Status Atual

- ✅ Projeto Firebase criado: `ecosistema-imob-dev`
- ⏳ Aguardando configuração dos serviços

---

## 📋 Checklist Rápido (15 minutos)

### ☐ 1. Habilitar Firestore (3 min)

1. Abra: https://console.firebase.google.com/project/ecosistema-imob-dev/firestore
2. Clique em **"Criar banco de dados"**
3. Selecione: **"Iniciar no modo de produção"**
4. Localização: **"southamerica-east1"** (São Paulo) ou **"us-east1"**
5. Clique em **"Ativar"**

**Como verificar:** Você verá a mensagem "Cloud Firestore" com abas: Dados, Regras, Índices

---

### ☐ 2. Habilitar Authentication (2 min)

1. Abra: https://console.firebase.google.com/project/ecosistema-imob-dev/authentication
2. Clique em **"Começar"**
3. Clique em **"E-mail/senha"**
   - Ative o primeiro toggle (E-mail/senha)
   - Salve
4. Clique em **"Google"**
   - Ative
   - E-mail de suporte: seu email
   - Salve

**Como verificar:** Na aba "Sign-in method", você verá Email/senha e Google como "Ativado"

---

### ☐ 3. Habilitar Cloud Storage (2 min)

1. Abra: https://console.firebase.google.com/project/ecosistema-imob-dev/storage
2. Clique em **"Começar"**
3. Selecione: **"Iniciar no modo de produção"**
4. Mesma localização do Firestore
5. Clique em **"Concluído"**

**Como verificar:** Você verá um bucket vazio em "gs://ecosistema-imob-dev.appspot.com"

---

### ☐ 4. Baixar Credenciais Admin SDK (2 min)

1. Abra: https://console.firebase.google.com/project/ecosistema-imob-dev/settings/serviceaccounts/adminsdk
2. Clique em **"Gerar nova chave privada"**
3. Confirme clicando em **"Gerar chave"**
4. Arquivo JSON será baixado
5. **RENOMEIE** para: `firebase-adminsdk.json`
6. **MOVA** para: `c:\Users\danie\OneDrive\Documentos\Altatech Systems\ecosystem\ecosistema-imob\backend\config\firebase-adminsdk.json`

**Como verificar no terminal:**
```bash
ls backend/config/firebase-adminsdk.json
# Deve mostrar o arquivo
```

---

### ☐ 5. Configurar Web App (3 min)

1. Abra: https://console.firebase.google.com/project/ecosistema-imob-dev/settings/general
2. Role até **"Seus apps"**
3. Clique no ícone **</>** (Web)
4. Nome do app: `ecosistema-imob-public`
5. ✅ Marque "Também configurar o Firebase Hosting"
6. Clique em **"Registrar app"**
7. **COPIE** o objeto firebaseConfig (guarde em um bloco de notas):

```javascript
const firebaseConfig = {
  apiKey: "AIza...",
  authDomain: "ecosistema-imob-dev.firebaseapp.com",
  projectId: "ecosistema-imob-dev",
  storageBucket: "ecosistema-imob-dev.firebasestorage.app",
  messagingSenderId: "...",
  appId: "..."
};
```

**Importante:** Salve essas informações - vamos usar ao configurar o frontend

---

### ☐ 6. Criar arquivo .env do backend (1 min)

Execute no terminal:

```bash
# Copiar .env.example para .env
cp backend/.env.example backend/.env

# Editar backend/.env e substituir:
# - FIREBASE_PROJECT_ID=ecosistema-imob-dev (já está correto)
# - GCS_BUCKET_NAME=ecosistema-imob-dev.firebasestorage.app (usar o storageBucket do passo 5)
```

Ou crie manualmente o arquivo `backend/.env` com:

```env
GOOGLE_APPLICATION_CREDENTIALS=./config/firebase-adminsdk.json
FIREBASE_PROJECT_ID=ecosistema-imob-dev
PORT=8080
GIN_MODE=debug
ENVIRONMENT=development
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,http://localhost:3002
GCS_BUCKET_NAME=ecosistema-imob-dev.firebasestorage.app
LOG_LEVEL=info
```

---

### ☐ 7. Instalar Firebase CLI e Deploy Índices (5 min)

Execute no terminal (PowerShell):

```bash
# Instalar Firebase CLI (se ainda não tiver)
npm install -g firebase-tools

# Verificar instalação
firebase --version

# Login
firebase login

# Ir para a pasta do projeto
cd "c:\Users\danie\OneDrive\Documentos\Altatech Systems\ecosystem\ecosistema-imob"

# Inicializar (apenas Firestore)
firebase init firestore

# Durante o init:
# ? Select a default Firebase project: ecosistema-imob-dev
# ? What file should be used for Firestore Rules?: firestore.rules (Enter)
# ? What file should be used for Firestore indexes?: firestore.indexes.json (Enter)

# Deploy dos 56 índices
firebase deploy --only firestore:indexes
```

**Aguarde 3-5 minutos** para todos os índices serem criados.

**Como verificar:**
1. Abra: https://console.firebase.google.com/project/ecosistema-imob-dev/firestore/indexes
2. Você verá 56 índices compostos (alguns podem estar "Criando...")

---

## ✅ Verificação Final

Execute no terminal:

```bash
# Verificar credenciais
ls backend/config/firebase-adminsdk.json

# Verificar .env
cat backend/.env | grep FIREBASE_PROJECT_ID

# Verificar Firebase CLI
firebase --version

# Verificar projeto selecionado
firebase projects:list
```

Se tudo estiver OK, você verá:
- ✅ Arquivo firebase-adminsdk.json existe
- ✅ FIREBASE_PROJECT_ID=ecosistema-imob-dev
- ✅ Firebase CLI instalado (versão 13.x ou superior)
- ✅ Projeto ecosistema-imob-dev listado

---

## 🎯 Após Completar Tudo

Você estará pronto para:

1. ✅ Compilar e rodar o backend Go
2. ✅ Conectar ao Firestore
3. ✅ Implementar Repositories/Services (Prompt 02)
4. ✅ Implementar Frontend Next.js (Prompt 09)

---

## 🆘 Ajuda

**Precisa de ajuda?**

Consulte o guia completo: [FIREBASE_SETUP_GUIDE.md](../FIREBASE_SETUP_GUIDE.md)

**Problema comum:**

- **Erro "Permission denied"**: Execute `firebase logout` e depois `firebase login` novamente
- **Índices não aparecem**: Aguarde 5-10 minutos (Firestore leva tempo para criar índices compostos)
- **Credenciais inválidas**: Baixe novamente do console Firebase

---

**Tempo Total Estimado:** 15-20 minutos
**Última Atualização:** 2025-12-21
