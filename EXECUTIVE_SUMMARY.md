# MarketPulse - Executive Summary

## Proposta de Valor

**MarketPulse** é uma plataforma editorial que transforma sinais de mercados de previsão em narrativas jornalísticas acessíveis, posicionando-se como a "Bloomberg dos sinais" para o público geral.

---

## Problema

Mercados de previsão (Polymarket) geram sinais valiosos sobre eventos futuros, mas:
- Dados brutos são inacessíveis para não-traders
- Mídia tradicional ignora ou reage com atraso
- Não existe tradução editorial em escala

---

## Solução

Pipeline automatizado que:
1. **Detecta** mudanças significativas em tempo real
2. **Contextualiza** com notícias externas
3. **Narra** em linguagem editorial (LLM)
4. **Distribui** via site, áudio e social

---

## Mercado

| Métrica | Valor |
|---------|-------|
| Volume Polymarket 2025 | $27.9B |
| Crescimento YoY | 13.95x |
| Projeção 2035 | $95.5B |
| Janela competitiva | 6-12 meses |

---

## Modelo de Negócio

### Fase 1 (MVP)
- Google AdSense
- Google News
- Afiliados editoriais

### Fase 2+
- Newsletter premium
- API para mídia
- Terminal institucional

---

## Stack Técnico

### APIs Necessárias
| API | Função |
|-----|--------|
| Polymarket Data | Sinais + probabilidades |
| Polymarket Gamma | Contexto + categorização |
| Web Search | Contextualização factual |
| Claude/OpenAI | Geração editorial |
| ElevenLabs (opcional) | Áudio/vídeo |

### Backend: **Golang** (recomendado)

| Critério | Golang | Python |
|----------|--------|--------|
| Performance real-time | +++ | + |
| Concorrência (polling APIs) | +++ | ++ |
| Deploy K8s | +++ | ++ |
| Suas skills | +++ | ++ |
| Libs LLM | ++ | +++ |
| Prototipagem rápida | ++ | +++ |

**Decisão:** Golang para core + Python para LLM scripts (híbrido).

---

## MVP Scope (4 semanas)

### Semana 1-2: Foundation
- [ ] Signal detector (thresholds simples)
- [ ] Polymarket client
- [ ] Classificador de categorias

### Semana 3: Editorial Engine
- [ ] Integração LLM
- [ ] Template narrativo
- [ ] Web search context

### Semana 4: Distribution
- [ ] Site estático (Hugo/Next)
- [ ] RSS/Sitemap
- [ ] Social publishing

---

## Heurísticas de Sinal (v1)

```
SINAL DETECTADO SE:
├─ Δ probabilidade ≥ ±7% em 24h
├─ Volume > 2x média 7 dias
├─ Reversão de tendência (3 dias consecutivos)
└─ Mercado novo com > $50k volume em 48h
```

---

## Arquitetura MVP

```
┌─────────────────────────────────────────────────────────┐
│                    SIGNAL PIPELINE                       │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌───────────┐  │
│  │  Polymarket  │───▶│   Signal     │───▶│  Context  │  │
│  │  Poller (Go) │    │   Detector   │    │  Builder  │  │
│  └──────────────┘    └──────────────┘    └───────────┘  │
│                                                │         │
│                                                ▼         │
│  ┌──────────────┐    ┌──────────────┐    ┌───────────┐  │
│  │    Site      │◀───│   Content    │◀───│    LLM    │  │
│  │   (Static)   │    │   Generator  │    │  (Claude) │  │
│  └──────────────┘    └──────────────┘    └───────────┘  │
│         │                                                │
│         ▼                                                │
│  ┌──────────────────────────────────────────────────┐   │
│  │  Distribution: AdSense + Social + Google News    │   │
│  └──────────────────────────────────────────────────┘   │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

---

## Custos Mensais (MVP)

| Item | Custo |
|------|-------|
| Hosting (VPS/K8s) | R$50-100 |
| Claude API | R$50-150 |
| Web Search API | R$0-30 |
| Domínio + CDN | R$20 |
| **Total** | **R$120-300/mês** |

---

## Revenue Projetado

| Timeline | Conservative | Base | Optimistic |
|----------|--------------|------|------------|
| Mês 6 | R$2k | R$5k | R$10k |
| Ano 1 | R$50k | R$100k | R$200k |
| Ano 2 | R$150k | R$400k | R$800k |

---

## Riscos e Mitigações

| Risco | Mitigação |
|-------|-----------|
| Polymarket rate limit | Cache agressivo + polling otimizado |
| AdSense rejection | Linguagem 100% editorial, zero financeiro |
| Competição (Bloomberg/Reuters) | First-mover, nicho específico |
| Dependência Polymarket | Expandir para Kalshi, Metaculus |

---

## Compliance AdSense

- Linguagem explicativa, nunca prescritiva
- Sem CTA financeiro ("aposte", "compre")
- Disclaimer editorial visível
- Sem links para trading
- Categorização clara (não-cripto)

---

## Próximos Passos (Esta Semana)

1. **Validação API** (2h)
   - Testar endpoints Polymarket
   - Verificar rate limits na prática

2. **Prototype Signal Detector** (4h)
   - Script Go básico
   - Threshold simples

3. **Template Editorial** (2h)
   - Prompt LLM v1
   - Formato de output

4. **Decisão GO/NO-GO** (1h)
   - Baseado em validações

---

## Decisão Final

| Critério | Status |
|----------|--------|
| Viabilidade Técnica | ✅ Alta |
| Viabilidade Mercado | ✅ Alta |
| Timing | ✅ Excelente |
| Custo Inicial | ✅ Baixo |
| Risco | 🟡 Moderado |

**Recomendação: EXECUTAR como side project (30% tempo)**

---

*Documento gerado em: 2025-12-19*
