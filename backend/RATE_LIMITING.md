# Rate Limiting

Este documento descreve a implementação de rate limiting no backend da API.

## Visão Geral

O rate limiting foi implementado para proteger a API contra:
- **Abuso de recursos**: Evita que um único cliente consuma recursos excessivos
- **DDoS**: Mitiga ataques de negação de serviço distribuído
- **Spam de Leads**: Previne criação em massa de leads falsos
- **Scraping**: Dificulta extração em massa de dados de imóveis

## Implementação

A implementação utiliza o algoritmo **Token Bucket** via `golang.org/x/time/rate`.

### Token Bucket

O algoritmo funciona assim:
1. Cada cliente (IP) tem um "balde" com tokens
2. Cada requisição consome 1 token
3. Tokens são repostos a uma taxa constante (requests per second)
4. Se não há tokens disponíveis, a requisição é rejeitada (429 Too Many Requests)

### Arquivo: `internal/middleware/rate_limiter.go`

O middleware implementa:
- **RateLimiterConfig**: Configuração personalizável
- **visitor**: Estrutura que armazena o limiter de cada IP
- **RateLimiter**: Gerencia visitantes e faz cleanup periódico
- **Limit()**: Middleware Gin que verifica rate limit por IP

## Configurações

### Strict Rate Limit (Rotas Públicas)

Aplicado a:
- `GET /api/v1/:tenant_id/properties`
- `GET /api/v1/:tenant_id/properties/:id`
- `GET /api/v1/:tenant_id/properties/slug/:slug`
- `POST /api/v1/:tenant_id/leads` ⚠️ **CRÍTICO - previne spam**
- `GET /api/v1/:tenant_id/property-images/*`

**Configuração:**
```go
RequestsPerSecond: 2.0  // 2 requests por segundo
Burst: 5                // Burst de até 5 requests
CleanupInterval: 5min   // Cleanup a cada 5 minutos
```

**Comportamento:**
- Cliente pode fazer 2 requests/segundo de forma sustentada
- Cliente pode fazer burst de até 5 requests rapidamente
- Após burst, precisa aguardar reposição de tokens (0.5s por token)

**Exemplo:**
```
t=0s:  5 requests em sequência ✅ (usa burst completo)
t=1s:  1 request ✅ (2 tokens repostos em 1s)
t=1s:  1 request ✅
t=1s:  1 request ❌ 429 Too Many Requests
t=2s:  1 request ✅ (2 tokens repostos em 1s)
```

### Default Rate Limit (Rotas Autenticadas)

Aplicado a:
- `POST /api/v1/admin/:tenant_id/brokers`
- `PUT /api/v1/admin/:tenant_id/properties/:id`
- Todas as rotas protegidas no `/admin/*`

**Configuração:**
```go
RequestsPerSecond: 10.0 // 10 requests por segundo
Burst: 20               // Burst de até 20 requests
CleanupInterval: 5min   // Cleanup a cada 5 minutos
```

**Comportamento:**
- Cliente pode fazer 10 requests/segundo de forma sustentada
- Cliente pode fazer burst de até 20 requests rapidamente
- Mais permissivo pois usuário está autenticado e identificado

## Uso no Código

### Aplicar Rate Limiting a uma Rota

```go
// Strict rate limiting (público)
publicRoutes := router.Group("/api/public")
publicRoutes.Use(middleware.StrictRateLimit())
{
    publicRoutes.POST("/leads", handler.CreateLead)
}

// Default rate limiting (autenticado)
adminRoutes := router.Group("/api/admin")
adminRoutes.Use(middleware.RateLimit())
{
    adminRoutes.POST("/properties", handler.CreateProperty)
}

// Custom rate limiting
customLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{
    RequestsPerSecond: 5.0,
    Burst:             10,
    CleanupInterval:   time.Minute,
})
customRoutes := router.Group("/api/custom")
customRoutes.Use(customLimiter.Limit())
```

## Resposta de Rate Limit Excedido

Quando o limite é excedido, a API retorna:

**Status Code:** `429 Too Many Requests`

**Response Body:**
```json
{
  "success": false,
  "error": "Rate limit exceeded. Please try again later."
}
```

## Tratamento no Frontend

### TypeScript/Axios

```typescript
import axios from 'axios';

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
});

// Interceptor para retry com backoff exponencial
api.interceptors.response.use(
  response => response,
  async error => {
    if (error.response?.status === 429) {
      // Rate limit excedido
      const retryAfter = error.response.headers['retry-after'] || 1;

      // Aguardar antes de retry
      await new Promise(resolve => setTimeout(resolve, retryAfter * 1000));

      // Retry request
      return api.request(error.config);
    }

    return Promise.reject(error);
  }
);
```

### React Component (com toast)

```tsx
const handleWhatsAppClick = async () => {
  try {
    await api.createLead({ ... });
  } catch (error) {
    if (error.response?.status === 429) {
      toast.error('Muitas requisições. Aguarde um momento e tente novamente.');
    } else {
      toast.error('Erro ao criar lead');
    }
  }
};
```

## Limitações por IP

O rate limiting é feito por **IP do cliente**, obtido via `c.ClientIP()` no Gin.

**Importante:**
- Se a aplicação estiver atrás de proxy/load balancer (Nginx, CloudFlare, etc.), configure o Gin para ler o IP real do header `X-Forwarded-For`:

```go
router.SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12"})
```

- **Produção com CloudFlare**: CloudFlare já faz rate limiting próprio. Considere ajustar limites.
- **NAT Corporativo**: Múltiplos usuários atrás do mesmo IP público compartilham o mesmo limite.

## Monitoramento

### Logs

O middleware não gera logs por padrão para não poluir. Para adicionar logs:

```go
func (rl *RateLimiter) Limit() gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()
        limiter := rl.getVisitor(ip)

        if !limiter.Allow() {
            log.Printf("Rate limit exceeded for IP: %s, Path: %s", ip, c.Request.URL.Path)
            c.JSON(http.StatusTooManyRequests, gin.H{
                "success": false,
                "error":   "Rate limit exceeded. Please try again later.",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### Métricas

Para monitorar rate limiting em produção, considere adicionar métricas Prometheus:

```go
var (
    rateLimitCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_rate_limit_exceeded_total",
            Help: "Total de requests bloqueados por rate limit",
        },
        []string{"path", "ip"},
    )
)

// No middleware:
if !limiter.Allow() {
    rateLimitCounter.WithLabelValues(c.Request.URL.Path, ip).Inc()
    // ...
}
```

## Ajuste de Limites

### Como Ajustar

Edite as constantes em `internal/middleware/rate_limiter.go`:

```go
func StrictRateLimiterConfig() RateLimiterConfig {
    return RateLimiterConfig{
        RequestsPerSecond: 5.0,  // ← Aumentar para 5/s
        Burst:             10,    // ← Aumentar burst
        CleanupInterval:   5 * time.Minute,
    }
}
```

### Quando Ajustar

**Aumentar limites:**
- Frontend legítimo sendo bloqueado
- Usuários reportam erro 429 durante uso normal
- Tráfego legítimo crescente

**Diminuir limites:**
- Ataques DDoS frequentes
- Scraping automatizado detectado
- Custos de infraestrutura altos

### Testes de Carga

Teste os limites antes de deploy:

```bash
# Apache Bench - 100 requests, 10 concorrentes
ab -n 100 -c 10 http://localhost:8080/api/v1/tenant123/properties

# Resultado esperado: alguns 429 após exceder limite
```

## Alternativas e Melhorias Futuras

### Redis-Based Rate Limiting

Para múltiplos servidores (horizontal scaling), use Redis:

```go
import "github.com/go-redis/redis_rate/v10"

limiter := redis_rate.NewLimiter(redisClient)
res, err := limiter.Allow(ctx, "project:123", redis_rate.PerSecond(10))
if err != nil {
    panic(err)
}
if res.Allowed == 0 {
    return fmt.Errorf("rate limit exceeded")
}
```

**Vantagens:**
- Rate limiting compartilhado entre múltiplas instâncias
- Persistente (não reseta ao reiniciar servidor)
- Suporta algoritmos avançados (sliding window)

### Rate Limiting por Tenant

Implementar limites diferentes por tenant:

```go
func (rl *RateLimiter) getTenantLimiter(tenantID string) *rate.Limiter {
    // Tenants premium: 100/s
    // Tenants free: 10/s
}
```

### Rate Limiting por Usuário Autenticado

Além de IP, limitar por `user_id`:

```go
func (rl *RateLimiter) getUserLimiter(userID string) *rate.Limiter {
    // Mais justo que por IP em NAT corporativo
}
```

## Troubleshooting

### "Rate limit exceeded" em desenvolvimento

**Solução:** Desabilite rate limiting em desenvolvimento:

```go
// cmd/server/main.go
if cfg.Environment == "development" {
    // Não aplicar rate limiting
} else {
    public.Use(middleware.StrictRateLimit())
}
```

### Frontend fazendo muitas requests

**Sintomas:** 429 frequente no console do navegador

**Causas comuns:**
- Múltiplas chamadas `useEffect` sem dependencies
- Polling muito agressivo
- Loop infinito de re-renders

**Solução:** Adicionar debounce/throttle no frontend:

```typescript
import { debounce } from 'lodash';

const debouncedFetch = debounce(async () => {
  await api.getProperties();
}, 500); // 500ms debounce
```

### Rate limit não funcionando atrás de proxy

**Sintoma:** Todos os requests parecem vir do mesmo IP

**Solução:** Configure Gin para ler IP real:

```go
router.SetTrustedProxies([]string{"10.0.0.0/8"})
```

## Segurança

### Não Adicione Headers de Rate Limit

**NÃO faça:**
```go
c.Header("X-RateLimit-Limit", "10")
c.Header("X-RateLimit-Remaining", "7")
```

**Por quê?**
- Expõe informação sobre limites para atacantes
- Facilita bypass programático
- Aumenta surface area de ataque

### Combine com Outras Defesas

Rate limiting não substitui:
- **Firewall** (CloudFlare, AWS WAF)
- **CAPTCHA** em formulários públicos
- **Autenticação** para operações sensíveis
- **Input validation** rigoroso
- **CSRF tokens** em formulários

## Referências

- [golang.org/x/time/rate Package](https://pkg.go.dev/golang.org/x/time/rate)
- [Token Bucket Algorithm](https://en.wikipedia.org/wiki/Token_bucket)
- [OWASP API Security - Rate Limiting](https://owasp.org/www-project-api-security/)
