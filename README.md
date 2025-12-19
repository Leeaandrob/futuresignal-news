# FutureSignals

**The Bloomberg of Prediction Market Signals**

Monitor prediction markets (Polymarket) and generate editorial narratives automatically using Qwen AI.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        FUTURESIGNALS                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐    ┌──────────────┐    ┌───────────────┐     │
│  │  Polymarket  │───▶│   Signal     │───▶│     Qwen      │     │
│  │   Poller     │    │   Detector   │    │   Narrator    │     │
│  │   (Go)       │    │   (Go)       │    │   (Go)        │     │
│  └──────────────┘    └──────────────┘    └───────────────┘     │
│                                                 │               │
│                                                 ▼               │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    JSON Output                            │  │
│  │  { signal, narrative, headline, tags, sentiment... }     │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                 │               │
│                                                 ▼               │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                 Frontend (Astro + shadcn)                 │  │
│  │                      (TODO)                               │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.23+
- DashScope API Key (Alibaba Qwen Cloud)

### Setup

```bash
# Clone and setup
cd futuresignals/backend
cp .env.example .env

# Edit .env with your DashScope API key
vim .env

# Install dependencies
make deps

# Run
make run
```

### Docker

```bash
# From project root
cp .env.example .env
vim .env  # Add your DASHSCOPE_API_KEY

docker compose up -d
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DASHSCOPE_API_KEY` | (required) | Qwen Cloud API key |
| `DASHSCOPE_ENDPOINT` | `https://dashscope-intl.aliyuncs.com/compatible-mode/v1` | API endpoint |
| `QWEN_MODEL` | `qwen-plus` | Model for narratives |
| `MIN_PROBABILITY_CHANGE` | `0.07` | Min change to trigger signal (7%) |
| `MIN_VOLUME_24H` | `50000` | Min 24h volume in USD |
| `POLL_INTERVAL` | `5m` | Market polling interval |
| `OUTPUT_DIR` | `./output` | Signal output directory |
| `DEBUG` | `false` | Enable debug logging |

## Signal Detection Heuristics

A signal is detected when:

```
SIGNAL DETECTED IF:
├─ Δ probability ≥ ±7% since last check
├─ Volume 24h > $50,000
├─ Market is active (not closed)
└─ OR: New market with > $100k volume in first 48h
```

## Output Format

Each detected signal produces a JSON file:

```json
{
  "signal": {
    "id": "abc123-20251219150000",
    "market_id": "abc123",
    "market_title": "Will Fed cut rates in March 2025?",
    "category": "politics",
    "previous_prob": 0.45,
    "current_prob": 0.57,
    "prob_change": 0.12,
    "volume_24h": 2400000,
    "signal_type": "surge",
    "significance": "high",
    "detected_at": "2025-12-19T15:00:00Z"
  },
  "narrative": {
    "headline": "Markets shift dramatically on Fed decision odds",
    "subheadline": "Probability jumped 12 points in 6 hours following unemployment data",
    "what_changed": "Markets began repricing Fed rate cut odds sharply higher...",
    "why_it_matters": "This represents the largest single-session shift since September...",
    "market_context": "The move follows this morning's unemployment data release...",
    "what_to_watch": "FOMC minutes release at 2pm ET could move markets further...",
    "tags": ["fed", "rates", "economy", "markets"],
    "sentiment": "bullish",
    "significance": "high"
  },
  "generated_at": "2025-12-19T15:00:05Z"
}
```

## Project Structure

```
futuresignals/
├── backend/
│   ├── cmd/
│   │   └── signald/          # Main daemon
│   ├── internal/
│   │   ├── polymarket/       # Polymarket API client
│   │   ├── qwen/             # Qwen/DashScope client
│   │   ├── detector/         # Signal detection logic
│   │   └── config/           # Configuration
│   ├── Makefile
│   ├── Dockerfile
│   └── go.mod
├── frontend/                  # (TODO) Astro + shadcn
├── data/
│   └── signals/              # Output directory
├── docker-compose.yml
└── README.md
```

## API Rate Limits (Polymarket)

| API | Limit |
|-----|-------|
| Gamma (markets/events) | 750/10s |
| Data (trades) | 200/10s |
| `/markets` | 125/10s |
| `/events` | 100/10s |

## Roadmap

- [x] Backend: Signal detector (Go)
- [x] Backend: Qwen integration (DashScope)
- [x] Backend: Polymarket API client
- [ ] Frontend: Astro + shadcn setup
- [ ] Frontend: Signal cards component
- [ ] Frontend: Daily briefing page
- [ ] Distribution: RSS feed
- [ ] Distribution: Social publishing

## License

MIT
