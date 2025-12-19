# FutureSignals - Executive Summary

## Status: LIVE

**Website:** [futuresignals.news](https://futuresignals.news)
**Version:** 1.1.0
**Last Updated:** 2024-12-19

---

## Proposta de Valor

**FutureSignals** é uma plataforma editorial que transforma sinais de mercados de previsão em narrativas jornalísticas acessíveis, posicionando-se como a "Bloomberg dos sinais" para o público geral.

---

## Problema Resolvido

Mercados de previsão (Polymarket) geram sinais valiosos sobre eventos futuros, mas:
- Dados brutos são inacessíveis para não-traders
- Mídia tradicional ignora ou reage com atraso
- Não existe tradução editorial em escala
- Falta correlação com sinais sociais de influenciadores

---

## Solução Implementada

Pipeline automatizado que:
1. **Detecta** mudanças significativas em tempo real (Polymarket API)
2. **Correlaciona** com sinais sociais de influenciadores (XTracker)
3. **Contextualiza** com notícias externas (Perplexity)
4. **Narra** em linguagem editorial Bloomberg-style (Qwen LLM)
5. **Distribui** via site SSR otimizado para SEO (Astro + Cloudflare)

---

## Mercado

| Métrica | Valor |
|---------|-------|
| Volume Polymarket 2025 | $27.9B |
| Crescimento YoY | 13.95x |
| Projeção 2035 | $95.5B |
| Janela competitiva | 6-12 meses |

---

## Stack Técnico (Produção)

### Backend (Go 1.23+)
| Componente | Tecnologia |
|------------|------------|
| Framework | Gin HTTP Router |
| Database | MongoDB Atlas |
| LLM | Qwen (DashScope) |
| Context | Perplexity API |
| Social | XTracker (Polymarket) |
| Deploy | Kubernetes (GKE) |

### Frontend (Astro 5.x)
| Componente | Tecnologia |
|------------|------------|
| Framework | Astro + React Islands |
| UI Library | shadcn/ui |
| Styling | Tailwind CSS 4 |
| Deploy | Cloudflare Pages (SSR) |
| SEO | JSON-LD, Google News schema |

---

## Arquitetura Produção

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                           FUTURESIGNALS v1.1.0                                │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ┌────────────────┐    ┌────────────────┐    ┌────────────────┐              │
│  │   Polymarket   │───▶│     Signal     │───▶│   XTracker     │              │
│  │    Poller      │    │    Detector    │    │   Correlator   │              │
│  └────────────────┘    └────────────────┘    └────────────────┘              │
│         │                     │                     │                         │
│         ▼                     ▼                     ▼                         │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                     Content Generator + LLM                              │ │
│  │       Qwen Narratives │ Perplexity Context │ Social Signals              │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                    │                                          │
│                                    ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                          MongoDB Atlas                                   │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                    │                                          │
│                                    ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                         REST API (Gin)                                   │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                    │                                          │
│                                    ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │              Frontend (Astro SSR @ Cloudflare Workers)                   │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                               │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Features Implementadas

### Backend
- [x] Signal detector com thresholds configuráveis
- [x] Polymarket API client com rate limiting
- [x] Qwen LLM integration (DashScope)
- [x] Perplexity context enrichment
- [x] XTracker social signal correlation
- [x] MongoDB persistence
- [x] REST API completa
- [x] Backfill tool para dados históricos

### Frontend
- [x] Homepage com feed personalizado
- [x] Páginas de artigos com SEO
- [x] Market cards com probability bars
- [x] Market Pulse (sentiment por categoria)
- [x] Social signals display
- [x] Twitter share + copy link
- [x] Categorias e filtros
- [x] SSR via Cloudflare Workers

### Infrastructure
- [x] Docker containerization
- [x] Kubernetes deployment
- [x] GHCR image registry
- [x] Cloudflare Pages
- [x] MongoDB Atlas

---

## Tipos de Artigo

| Tipo | Trigger | Descrição |
|------|---------|-----------|
| `breaking` | Δ prob > 10% | Notícia urgente |
| `trending` | Volume + movimento | Mercado em alta |
| `new_market` | Novo + volume | Mercado recém-criado |
| `briefing` | Horário | Resumo diário |
| `deep_dive` | Manual | Análise profunda |
| `social_signal` | XTracker | Baseado em tweets |

---

## Custos Operacionais

| Item | Custo Mensal |
|------|--------------|
| MongoDB Atlas (M0) | $0 |
| GKE Autopilot | ~$50 |
| Cloudflare Pages | $0 |
| Qwen API | ~$20-50 |
| Perplexity API | ~$10-20 |
| **Total** | **~$80-120** |

---

## Modelo de Negócio

### Fase Atual (MVP)
- Google AdSense (pendente aplicação)
- Google News indexing
- Organic SEO traffic

### Fase 2 (Q1 2025)
- Newsletter premium
- RSS feeds
- Social publishing automation

### Fase 3 (Q2 2025)
- API para mídia/traders
- Audio briefings (ElevenLabs)
- Expand: Kalshi, Metaculus

---

## Métricas (a acompanhar)

| Métrica | Target Mês 1 | Target Mês 3 |
|---------|--------------|--------------|
| Articles/day | 10-20 | 50+ |
| Page views | 1k | 10k |
| Google News | Pending | Indexed |
| AdSense | Applied | Active |

---

## Próximos Passos

### Imediato
1. Aplicar Google AdSense
2. Submeter para Google News
3. Adicionar mais influenciadores ao XTracker tracking
4. Implementar RSS feeds

### Curto Prazo
1. Newsletter com Resend
2. Social publishing (Twitter bot)
3. Audio briefings

### Médio Prazo
1. Kalshi integration
2. Premium API tier
3. Mobile app

---

## Compliance

### AdSense
- [x] Linguagem explicativa, nunca prescritiva
- [x] Sem CTA financeiro
- [x] Disclaimer editorial visível
- [x] Sem links diretos para trading
- [x] Categorização clara

### SEO
- [x] JSON-LD NewsArticle schema
- [x] OpenGraph meta tags
- [x] Twitter cards
- [x] Sitemap.xml
- [x] Robots.txt

---

## Conclusão

| Critério | Status |
|----------|--------|
| MVP Completo | ✅ |
| Live em Produção | ✅ |
| SEO Otimizado | ✅ |
| Social Signals | ✅ |
| Monetização | 🟡 Pendente |
| Escala | 🟡 A validar |

**Status:** OPERACIONAL - Aguardando aprovação AdSense e indexação Google News

---

*Atualizado em: 2024-12-19*
