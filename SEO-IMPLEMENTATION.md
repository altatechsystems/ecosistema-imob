# Implementação de SEO - Ecosistema Imob

## Visão Geral

Este documento descreve as otimizações de SEO implementadas no frontend público da plataforma Ecosistema Imob, seguindo as melhores práticas do Google e Schema.org.

## 1. Metadata Otimizada

### Layout Principal (`app/layout.tsx`)

**Implementações:**

- **Title Template**: Estrutura consistente de títulos em todas as páginas
  ```typescript
  title: {
    default: "Imobiliária - Encontre seu Imóvel Ideal | Casas, Apartamentos e Terrenos",
    template: "%s | Imobiliária"
  }
  ```

- **Meta Description**: Descrição otimizada com palavras-chave relevantes
- **Keywords**: Array abrangente de termos relacionados a imóveis
- **Open Graph Tags**: Otimização para compartilhamento em redes sociais
- **Twitter Cards**: Suporte para previews no Twitter
- **Robots Configuration**: Controle fino de indexação
- **Canonical URLs**: URLs canônicas para evitar conteúdo duplicado

**Variáveis de Ambiente Necessárias:**
```env
NEXT_PUBLIC_SITE_URL=https://www.seusite.com.br
NEXT_PUBLIC_GOOGLE_SITE_VERIFICATION=seu_codigo_google
NEXT_PUBLIC_YANDEX_VERIFICATION=seu_codigo_yandex
```

## 2. Arquivos de SEO

### robots.txt (`app/robots.ts`)

Define regras para crawlers de busca:

- **Allow**: `/` (todo o site público)
- **Disallow**: `/api/`, `/admin/`, `/_next/` (áreas privadas)
- **Sitemap**: Link para sitemap.xml
- **Crawl Delay**: Configuração específica por bot

### Sitemap Dinâmico (`app/sitemap.ts`)

Sitemap XML gerado automaticamente:

- **Páginas Estáticas**: Home, Imóveis, Sobre, Contato
- **Páginas Dinâmicas**: Todos os imóveis públicos disponíveis
- **Prioridades**:
  - Homepage: 1.0
  - Listagem de imóveis: 0.9
  - Imóveis em destaque: 0.8
  - Imóveis normais: 0.7
  - Páginas estáticas: 0.5
- **Frequência de Atualização**:
  - Homepage e listagem: daily
  - Imóveis: weekly
  - Páginas estáticas: monthly
- **Revalidação**: A cada 1 hora (3600 segundos)

## 3. Schema.org Structured Data

### RealEstateListing Schema (`PropertyStructuredData`)

Implementado em cada página de detalhes de imóvel:

```json
{
  "@context": "https://schema.org",
  "@type": "RealEstateListing",
  "name": "Título do Imóvel",
  "description": "Descrição completa",
  "url": "URL da página",
  "image": ["URLs das imagens"],
  "address": {
    "@type": "PostalAddress",
    "streetAddress": "Rua, número",
    "addressLocality": "Cidade",
    "addressRegion": "Estado",
    "postalCode": "CEP",
    "addressCountry": "BR"
  },
  "geo": {
    "@type": "GeoCoordinates",
    "latitude": -23.5505,
    "longitude": -46.6333
  },
  "offers": {
    "@type": "Offer",
    "price": 500000,
    "priceCurrency": "BRL",
    "availability": "https://schema.org/InStock"
  },
  "numberOfRooms": 3,
  "numberOfBathroomsTotal": 2,
  "floorSize": {
    "@type": "QuantitativeValue",
    "value": 120,
    "unitCode": "MTK"
  },
  "amenityFeature": [...]
}
```

**Benefícios:**
- Rich snippets nos resultados de busca
- Exibição de preço, localização e características
- Melhor CTR (Click-Through Rate)
- Elegível para Google Rich Results

### RealEstateAgent Schema (`OrganizationStructuredData`)

Implementado na homepage:

```json
{
  "@context": "https://schema.org",
  "@type": "RealEstateAgent",
  "name": "Imobiliária",
  "url": "https://www.seusite.com.br",
  "logo": "URL do logo",
  "description": "Descrição da imobiliária",
  "contactPoint": {
    "@type": "ContactPoint",
    "telephone": "+5511999999999",
    "contactType": "customer service",
    "areaServed": "BR",
    "availableLanguage": ["Portuguese"]
  },
  "sameAs": [
    "https://www.facebook.com/imobiliaria",
    "https://www.instagram.com/imobiliaria"
  ]
}
```

**Benefícios:**
- Knowledge Graph do Google
- Informações de contato destacadas
- Links para redes sociais
- Validação como negócio legítimo

### BreadcrumbList Schema (`BreadcrumbStructuredData`)

Implementado em páginas de detalhes:

```json
{
  "@context": "https://schema.org",
  "@type": "BreadcrumbList",
  "itemListElement": [
    {
      "@type": "ListItem",
      "position": 1,
      "name": "Início",
      "item": "https://www.seusite.com.br/"
    },
    {
      "@type": "ListItem",
      "position": 2,
      "name": "Imóveis",
      "item": "https://www.seusite.com.br/imoveis"
    },
    {
      "@type": "ListItem",
      "position": 3,
      "name": "Apartamento em São Paulo",
      "item": "https://www.seusite.com.br/imoveis/apartamento-sp-123"
    }
  ]
}
```

**Benefícios:**
- Breadcrumbs visíveis nos resultados de busca
- Melhor navegação hierárquica
- Contexto para motores de busca

## 4. Open Graph Images

### Geração Dinâmica (`opengraph-image.tsx`)

Imagens Open Graph geradas automaticamente usando Next.js Image Response:

- **Dimensões**: 1200x630px (padrão Facebook/LinkedIn)
- **Formato**: PNG
- **Design**: Gradiente azul com texto branco
- **Conteúdo**: Título da página + branding

**Exemplo de Uso:**
```html
<meta property="og:image" content="https://www.seusite.com.br/opengraph-image.png" />
<meta property="og:image:width" content="1200" />
<meta property="og:image:height" content="630" />
```

## 5. Checklist de Implementação

### Desenvolvimento

- [x] Metadata otimizada no layout raiz
- [x] robots.ts configurado
- [x] sitemap.ts dinâmico
- [x] Schema.org structured data
- [x] Open Graph images
- [x] Twitter Cards
- [x] Breadcrumbs estruturados
- [x] .env.example documentado

### Produção (Pendente)

- [ ] Configurar NEXT_PUBLIC_SITE_URL para domínio de produção
- [ ] Adicionar Google Search Console verification code
- [ ] Adicionar Yandex Webmaster verification code
- [ ] Configurar Google Analytics (se NEXT_PUBLIC_ENABLE_ANALYTICS=true)
- [ ] Submeter sitemap.xml ao Google Search Console
- [ ] Submeter sitemap.xml ao Bing Webmaster Tools
- [ ] Testar structured data com Google Rich Results Test
- [ ] Validar Open Graph com Facebook Sharing Debugger
- [ ] Validar Twitter Cards com Twitter Card Validator
- [ ] Configurar imagens OG personalizadas para propriedades

## 6. Ferramentas de Validação

### Antes do Deploy

1. **Google Rich Results Test**: https://search.google.com/test/rich-results
   - Valida Schema.org structured data
   - Verifica elegibilidade para rich snippets

2. **Schema.org Validator**: https://validator.schema.org/
   - Valida JSON-LD syntax
   - Detecta erros de estrutura

3. **Facebook Sharing Debugger**: https://developers.facebook.com/tools/debug/
   - Testa Open Graph tags
   - Preview de compartilhamento

4. **Twitter Card Validator**: https://cards-dev.twitter.com/validator
   - Valida Twitter Cards
   - Preview de tweets

5. **Google PageSpeed Insights**: https://pagespeed.web.dev/
   - Performance e SEO básico
   - Core Web Vitals

### Após Deploy

1. **Google Search Console**:
   - Submeter sitemap
   - Monitorar indexação
   - Verificar erros de rastreamento
   - Analisar queries de busca

2. **Bing Webmaster Tools**:
   - Submeter sitemap
   - Verificar coverage

3. **Analytics**:
   - Google Analytics (se habilitado)
   - Monitorar tráfego orgânico

## 7. Boas Práticas Implementadas

### Conteúdo

- ✅ Títulos únicos e descritivos para cada página
- ✅ Meta descriptions entre 150-160 caracteres
- ✅ URLs amigáveis com slugs semânticos
- ✅ Hierarquia de headings (H1 → H6) correta
- ✅ Alt text em todas as imagens
- ✅ Links internos contextuais

### Técnico

- ✅ Sitemap XML atualizado automaticamente
- ✅ robots.txt otimizado
- ✅ Canonical URLs configuradas
- ✅ Schema.org markup validado
- ✅ Mobile-first responsive design
- ✅ Lazy loading de imagens
- ✅ Compressão de imagens (JPEG 75-90%)

### Performance

- ✅ Next.js Image optimization
- ✅ Code splitting automático
- ✅ Prefetching de links
- ✅ Revalidação incremental do sitemap

## 8. Próximos Passos (Recomendações)

### Curto Prazo

1. **Conteúdo Otimizado**:
   - Adicionar descrições únicas para cada imóvel
   - Criar títulos otimizados para SEO
   - Melhorar alt text das imagens com keywords

2. **Blog/Conteúdo**:
   - Criar seção de blog para conteúdo evergreen
   - Artigos sobre mercado imobiliário
   - Guias de compra/aluguel

3. **Local SEO**:
   - Adicionar LocalBusiness schema
   - Google My Business integration
   - Mapas e localização

### Médio Prazo

1. **Análise e Otimização**:
   - Implementar Google Analytics 4
   - Configurar conversões e goals
   - A/B testing de meta descriptions

2. **Link Building**:
   - Estratégia de backlinks
   - Parcerias com portais imobiliários
   - Guest posts em blogs relevantes

3. **Conteúdo Avançado**:
   - Vídeos de propriedades
   - Tours virtuais 360°
   - Calculadoras de financiamento

### Longo Prazo

1. **Internacional**:
   - Suporte multi-idioma (hreflang)
   - Geo-targeting avançado

2. **Features Avançadas**:
   - AMP (Accelerated Mobile Pages)
   - PWA (Progressive Web App)
   - Voice search optimization

## 9. Monitoramento e KPIs

### Métricas Essenciais

- **Orgânico**: Tráfego de busca orgânica
- **Rankings**: Posições de keywords principais
- **CTR**: Click-through rate nos SERPs
- **Conversões**: Leads gerados via busca
- **Core Web Vitals**: LCP, FID, CLS
- **Indexação**: Páginas indexadas vs total

### Ferramentas

- Google Search Console
- Google Analytics 4
- SEMrush / Ahrefs (opcional)
- Hotjar (heatmaps e behavior)

## 10. Suporte e Documentação

### Recursos Úteis

- [Next.js Metadata Documentation](https://nextjs.org/docs/app/building-your-application/optimizing/metadata)
- [Schema.org Real Estate](https://schema.org/RealEstateListing)
- [Google Search Central](https://developers.google.com/search)
- [Open Graph Protocol](https://ogp.me/)

### Contato

Para dúvidas sobre a implementação de SEO, consulte a documentação do projeto ou entre em contato com o time de desenvolvimento.

---

**Última atualização**: 2025-12-29
**Versão**: 1.0.0
**Status**: Implementado no frontend público
