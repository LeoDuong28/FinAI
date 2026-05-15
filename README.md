# FinAI

AI-powered personal finance tracker built with Go, HTMX, and Python.

## Tech Stack

- **Backend:** Go 1.24, Chi router, sqlc, pgx
- **Frontend:** Templ + HTMX + Tailwind CSS
- **AI Service:** Python 3.12, FastAPI, FinBERT, Google Gemini
- **Database:** PostgreSQL 16 (Neon.tech for production)
- **Deployment:** Render.com, GitHub Actions CI/CD

## Features

- Bank account linking via Plaid (sandbox)
- Transaction tracking with AI categorization
- Budget management with spending alerts
- Recurring bill tracking and reminders
- Subscription monitoring
- Savings goals with fund tracking
- AI-powered financial insights and forecasting
- Net worth tracking with historical snapshots
- AI chat assistant (Gemini-powered)
- Bill negotiation tips
- PWA support with offline capability
- Two-factor authentication (TOTP)

## Getting Started

### Prerequisites

- Go 1.24+
- Python 3.12+
- Docker & Docker Compose
- Node.js (for Tailwind CSS)

### Local Development

```bash
# Clone and configure
cp .env.example .env
# Edit .env with your API keys (Plaid, Gemini, etc.)

# Start database and all services
make dev

# Or run services individually
docker compose up -d postgres
make air          # Go backend with hot reload
make templ-watch  # Template file watcher
make tailwind-watch  # CSS watcher
```

The app will be available at `http://localhost:8080`.

### Database Migrations

```bash
make migrate-up        # Run all pending migrations
make migrate-down      # Rollback last migration
make migrate-create NAME=add_feature  # Create new migration
```

### Code Generation

```bash
make sqlc   # Regenerate type-safe SQL code
make templ  # Regenerate Templ templates
```

### Testing

```bash
make test         # Run all tests with race detector
make test-cover   # Run tests with coverage report
make security     # Run gosec + govulncheck
make lint         # Run golangci-lint
```

### Docker

```bash
make up    # Start all services (postgres, go-backend, ai-service)
make down  # Stop all services
make build # Build production binary
```

## Project Structure

```
cmd/server/          # Application entry point
internal/
  config/            # Configuration loading and validation
  domain/            # Domain models and repository interfaces
  errors/            # Structured error types
  handler/           # HTTP handlers (11 modules)
  middleware/        # Auth, CSRF, rate limiting, security headers
  repository/postgres/  # PostgreSQL implementations (12 repos)
  router/            # Route definitions and dependency wiring
  service/           # Business logic (18 services)
  testutil/          # Test mocks and fixtures
migrations/          # PostgreSQL migrations (15 files)
templates/           # Templ templates (layouts, pages, components)
static/              # CSS, JS, PWA assets
ai-service/          # Python FastAPI microservice
  app/routers/       # Categorize, insights, negotiate, chat, subscriptions
  app/services/      # FinBERT categorizer, anomaly detector, negotiator
```

## Deployment

### Render.com (Free Tier)

1. Connect your GitHub repo to Render
2. Use the `render.yaml` blueprint for automatic service setup
3. Set `DATABASE_URL` to your Neon.tech connection string
4. Set `GEMINI_API_KEY` and `PLAID_CLIENT_ID`/`PLAID_SECRET`
5. Migrations run automatically on each deploy via pre-deploy command

### Environment Variables

See [.env.example](.env.example) for all required configuration.

## Architecture

- **Clean Architecture** with domain-driven design boundaries
- **Modular Monolith** (Go) + **AI Microservice** (Python)
- Repository pattern with sqlc-generated type-safe queries
- Circuit breaker for resilient AI service communication
- Token bucket rate limiting per route group
- AES-256-GCM encryption for sensitive credentials
- JWT authentication with cookie-based tokens
- CSRF double-submit cookie protection

## License

MIT
