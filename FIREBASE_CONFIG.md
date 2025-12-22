# 🔥 Firebase Configuration

> **IMPORTANTE:** Este arquivo contém configurações públicas do Firebase.
> **NÃO COMMITAR** se contiver informações sensíveis!

---

## 📋 Firebase Web App Config

**App Name:** ecosistema-imob-public
**App ID:** 1:83278095706:web:4ce16e73d01c0307a73f63

### JavaScript SDK Config

```javascript
const firebaseConfig = {
  apiKey: "AIzaSyDx6PRtyNnAgTxsT3coow9Ut0xM4Xi5A71",
  authDomain: "ecosistema-imob-dev.firebaseapp.com",
  projectId: "ecosistema-imob-dev",
  storageBucket: "ecosistema-imob-dev.firebasestorage.app",
  messagingSenderId: "83278095706",
  appId: "1:83278095706:web:4ce16e73d01c0307a73f63"
};
```

### Environment Variables para Next.js

Cole no arquivo `frontend-public/.env.local`:

```env
NEXT_PUBLIC_FIREBASE_API_KEY=AIzaSyDx6PRtyNnAgTxsT3coow9Ut0xM4Xi5A71
NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=ecosistema-imob-dev.firebaseapp.com
NEXT_PUBLIC_FIREBASE_PROJECT_ID=ecosistema-imob-dev
NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET=ecosistema-imob-dev.firebasestorage.app
NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID=83278095706
NEXT_PUBLIC_FIREBASE_APP_ID=1:83278095706:web:4ce16e73d01c0307a73f63
```

---

## ✅ Progresso da Configuração

- [x] Firestore Database criado
- [x] Authentication habilitado (Email/Password + Google)
- [x] Web App registrado
- [x] Firebase Config salvo
- [ ] Cloud Storage (próximo passo)
- [ ] Admin SDK credentials (próximo passo)

---

**Última Atualização:** 2025-12-21
