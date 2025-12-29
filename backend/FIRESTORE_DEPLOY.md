# Deploy de Índices e Regras do Firestore

Este documento explica como fazer deploy dos índices compostos e regras de segurança do Firestore.

## Pré-requisitos

1. Firebase CLI instalado:
```bash
npm install -g firebase-tools
```

2. Autenticação no Firebase:
```bash
firebase login
```

## Configuração do Projeto

1. Inicializar Firebase no diretório (se ainda não foi feito):
```bash
firebase init firestore
```

Quando solicitado:
- Selecione o projeto: `ecosistema-imob-dev`
- Firestore Rules file: `firestore.rules`
- Firestore Indexes file: `firestore.indexes.json`

## Deploy

### Deploy de Índices Compostos

```bash
firebase deploy --only firestore:indexes --project ecosistema-imob-dev
```

Este comando irá:
- Ler o arquivo `firestore.indexes.json`
- Criar os índices compostos no Firestore
- Pode levar alguns minutos para os índices serem construídos

**IMPORTANTE:** Após o deploy, aguarde a construção dos índices no Console do Firebase antes de executar queries que dependam deles.

### Deploy de Regras de Segurança

```bash
firebase deploy --only firestore:rules --project ecosistema-imob-dev
```

Este comando irá:
- Ler o arquivo `firestore.rules`
- Atualizar as regras de segurança no Firestore
- Aplicação imediata (sem delay)

### Deploy Completo (Índices + Regras)

```bash
firebase deploy --only firestore --project ecosistema-imob-dev
```

## Verificação

### Verificar Índices

1. Acesse o Firebase Console: https://console.firebase.google.com
2. Selecione o projeto `ecosistema-imob-dev`
3. Navegue para **Firestore Database** > **Indexes**
4. Verifique se todos os índices estão com status "Enabled" (podem levar alguns minutos)

### Verificar Regras

1. Acesse o Firebase Console
2. Navegue para **Firestore Database** > **Rules**
3. Verifique se as regras foram atualizadas com timestamp recente

### Teste de Regras (Simulador)

No Firebase Console, você pode usar o **Rules Playground** para testar as regras:

1. Vá para **Firestore Database** > **Rules**
2. Clique em **Playground**
3. Teste diferentes cenários:
   - Usuário não autenticado lendo properties (deve permitir)
   - Usuário não autenticado criando lead (deve permitir)
   - Usuário não autenticado criando broker (deve negar)
   - Usuário autenticado do tenant X acessando dados do tenant Y (deve negar)

## Índices Criados

O arquivo `firestore.indexes.json` contém índices compostos para as seguintes queries:

### Properties
- `tenant_id + status + visibility + created_at` (listagem básica)
- `tenant_id + status + visibility + featured + created_at` (destaques)
- `tenant_id + transaction_type + status + visibility + created_at` (filtro por tipo de transação)
- `tenant_id + property_type + status + visibility + created_at` (filtro por tipo de imóvel)
- `tenant_id + city + status + visibility + created_at` (filtro por cidade)
- `tenant_id + neighborhood + status + visibility + created_at` (filtro por bairro)

### Listings
- `tenant_id + property_id + created_at` (listings de um property)
- `tenant_id + status + created_at` (listings por status)

### Leads
- `tenant_id + property_id + created_at` (leads de um property)
- `tenant_id + status + created_at` (leads por status)
- `tenant_id + channel + created_at` (leads por canal)

### Activity Logs
- `tenant_id + entity_type + entity_id + timestamp` (timeline de uma entidade)
- `tenant_id + action + timestamp` (logs por tipo de ação)

### Brokers & Owners
- `tenant_id + is_active + created_at` (brokers ativos/inativos)
- `tenant_id + is_anonymized + created_at` (owners anonimizados)

## Troubleshooting

### Erro: "The query requires an index"

Se você receber este erro ao executar uma query:

1. Copie o link fornecido no erro (geralmente começa com https://console.firebase.google.com/...)
2. Abra o link no navegador
3. Clique em "Create Index"
4. Aguarde a construção do índice
5. Adicione o índice ao `firestore.indexes.json` para não perder no próximo deploy

### Erro: "Permission denied"

Verifique as regras de segurança:
1. O usuário está autenticado?
2. O token JWT contém `tenant_id` correto?
3. A operação está permitida nas rules?

### Índices Não Aparecem

Se após o deploy os índices não aparecem:
1. Verifique se o projeto está correto: `firebase use`
2. Verifique se o database é `(default)` ou `imob-dev`
3. Se usar database não-default, especifique no deploy:
```bash
firebase deploy --only firestore:indexes --project ecosistema-imob-dev
```

## Manutenção

### Adicionar Novo Índice

1. Edite `firestore.indexes.json`
2. Adicione o novo índice na array `indexes`
3. Execute: `firebase deploy --only firestore:indexes`
4. Aguarde construção no Console

### Remover Índice Obsoleto

1. Remova do `firestore.indexes.json`
2. Execute deploy
3. No Console, delete manualmente o índice antigo (opcional, não há cobrança por índices não usados)

### Atualizar Regras

1. Edite `firestore.rules`
2. Execute: `firebase deploy --only firestore:rules`
3. Teste no Playground

## Custos

- **Índices Compostos:** Sem custo adicional (apenas storage das entradas de índice)
- **Regras de Segurança:** Sem custo
- **Queries com Índices:** Mesmo custo de leitura de documentos

## Referências

- [Firestore Index Documentation](https://firebase.google.com/docs/firestore/query-data/indexing)
- [Firestore Security Rules](https://firebase.google.com/docs/firestore/security/get-started)
- [Firebase CLI Reference](https://firebase.google.com/docs/cli)
