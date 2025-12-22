# Testes - Ecosistema Imob Backend

Este diretório contém os testes automatizados para o backend do Ecosistema Imob.

## Estrutura

```
tests/
├── integration/          # Testes de integração (API + Firestore)
│   └── integration_test.go
└── README.md            # Este arquivo
```

## Tipos de Testes

### Testes Unitários
Localizados junto ao código fonte em `internal/*/`

- **validators_test.go**: Testes de validadores brasileiros (CPF, CNPJ, CRECI, etc.)
- Executar: `go test ./internal/utils -v`

### Testes de Integração
Localizados em `tests/integration/`

- Testam fluxos completos da aplicação
- Requerem Firebase configurado
- Criam e limpam dados de teste automaticamente

## Executando os Testes

### Pré-requisitos

1. Firebase configurado com credenciais em `backend/config/firebase-adminsdk.json`
2. Variáveis de ambiente configuradas (`.env` file)

### Testes Unitários

```bash
# Todos os testes unitários
go test ./internal/... -v

# Apenas validadores
go test ./internal/utils -v

# Com coverage
go test ./internal/... -cover
```

### Testes de Integração

```bash
# Executar testes de integração
RUN_INTEGRATION_TESTS=1 go test ./tests/integration/... -v

# Com timeout maior (para operações de Firestore)
RUN_INTEGRATION_TESTS=1 go test ./tests/integration/... -v -timeout 5m
```

### Todos os Testes

```bash
# Executar todos os testes (unitários + integração)
RUN_INTEGRATION_TESTS=1 go test ./... -v

# Com coverage report
RUN_INTEGRATION_TESTS=1 go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Cobertura de Testes

### Testes Unitários

- ✅ Validadores (CPF, CNPJ, CRECI, Email, Phone) - 100%
- ✅ Normalização de dados - 100%

### Testes de Integração

- ✅ **Tenant Creation**: Criação de tenant com slug auto-gerado
- ✅ **Property Lifecycle**: CRUD completo de imóveis
  - Criação de property
  - Busca por ID
  - Atualização de status
  - Listagem de properties
- ✅ **Lead Creation and Routing**: Criação de lead com roteamento para corretor

## Fluxos Testados

### 1. Tenant Management
- Criação de tenant
- Validação de CNPJ
- Geração automática de slug
- Normalização de email e telefone

### 2. Property Management
- Criação de imóvel vinculado a proprietário
- Busca de imóvel por ID
- Atualização de status (available → sold)
- Listagem com filtros
- Geração de fingerprint para deduplicação

### 3. Lead Management
- Criação de lead via API
- Validação de consentimento (LGPD)
- Roteamento automático para corretor
- Vinculação lead → property → broker

## Dados de Teste

Os testes criam dados temporários no Firestore que são automaticamente limpos após a execução.

### Tenant de Teste
```json
{
  "name": "Test Imobiliária",
  "document": "11.222.333/0001-81",
  "email": "test@imob.com",
  "phone": "(11) 98765-4321"
}
```

### Property de Teste
```json
{
  "transaction_type": "sale",
  "property_type": "apartment",
  "sale_price": 500000.00,
  "bedrooms": 3,
  "bathrooms": 2,
  "area_sqm": 85.5,
  "city": "São Paulo",
  "state": "SP"
}
```

## Troubleshooting

### Erro: "Firebase credentials not found"
```bash
# Verifique se o arquivo existe
ls backend/config/firebase-adminsdk.json

# Se não existir, baixe do Firebase Console:
# 1. Acesse Firebase Console
# 2. Project Settings → Service Accounts
# 3. Generate New Private Key
# 4. Salve como firebase-adminsdk.json em backend/config/
```

### Erro: "Permission denied" no Firestore
```bash
# Verifique as regras de segurança do Firestore
# Em desenvolvimento, você pode temporariamente usar:
# allow read, write: if true;

# Ou execute firebase deploy --only firestore:rules
```

### Testes lentos
```bash
# Aumente o timeout
go test ./tests/integration/... -v -timeout 10m

# Ou execute em paralelo
go test ./tests/integration/... -v -parallel 4
```

## CI/CD

Para integração contínua, adicione ao seu pipeline:

```yaml
# GitHub Actions exemplo
- name: Run Tests
  run: |
    export RUN_INTEGRATION_TESTS=1
    go test ./... -v -cover
  env:
    FIREBASE_CREDENTIALS_PATH: ./config/firebase-adminsdk.json
```

## Próximos Passos

Testes a adicionar:

- [ ] Testes de co-corretagem (múltiplos brokers por property)
- [ ] Testes de canonical listings
- [ ] Testes de deduplicação de properties
- [ ] Testes de LGPD (anonimização, revogação de consentimento)
- [ ] Testes de upload de imagens (Storage)
- [ ] Testes de performance (stress testing)
- [ ] Testes de segurança (SQL injection, XSS, etc.)

## Boas Práticas

1. **Isolamento**: Cada teste cria seus próprios dados e limpa após execução
2. **Idempotência**: Testes podem rodar múltiplas vezes sem side effects
3. **Nomenclatura**: `Test<Feature><Scenario>` (ex: TestPropertyCreation)
4. **Assertions**: Use `assert` para checks não-críticos e `require` para pré-condições
5. **Cleanup**: Sempre use `defer` para garantir limpeza de dados

## Referências

- [Go Testing](https://golang.org/pkg/testing/)
- [Testify](https://github.com/stretchr/testify)
- [Firebase Admin SDK Go](https://firebase.google.com/docs/admin/setup)
