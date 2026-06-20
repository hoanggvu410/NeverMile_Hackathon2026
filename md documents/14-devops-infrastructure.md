# 14 — DevOps & Infrastructure

---

## 1. Production Infrastructure

```
┌─────────────────────────────────────────────────────┐
│  Cloudflare (CDN + DNS)                              │
│  gitwhy.dev   →  Framer (static site)                │
│  docs.gitwhy.dev  →  Docs (static)                   │
│  dl.gitwhy.dev    →  Binary downloads (R2/S3)        │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│  Fly.io (Go API + Web Dashboard)                     │
│  api.gitwhy.dev   →  Go Cloud API (2 instances)     │
│  app.gitwhy.dev   →  Next.js Dashboard               │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│  Managed Services                                    │
│  PostgreSQL: Supabase / Neon / RDS                   │
│  Redis: Upstash / Railway Redis                      │
└─────────────────────────────────────────────────────┘
```

---

## 2. Go API — Dockerfile

```dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /gitwhy-api ./cmd/api/main.go

FROM alpine:3.19
RUN adduser -D -u 1000 appuser
COPY --from=builder /gitwhy-api /usr/local/bin/gitwhy-api
USER appuser

EXPOSE 8080
CMD ["/usr/local/bin/gitwhy-api"]
```

---

## 3. Next.js Dashboard — Dockerfile

```dockerfile
FROM node:20-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

FROM node:20-alpine AS builder
WORKDIR /app
COPY . .
COPY --from=deps /app/node_modules ./node_modules
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production

COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

EXPOSE 3000
CMD ["node", "server.js"]
```

---

## 4. Docker Compose (development)

```yaml
version: "3.9"

services:
  api:
    build:
      context: ./gitwhy-cloud
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    env_file: ./gitwhy-cloud/.env
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped

  dashboard:
    build:
      context: ./gitwhy-dashboard
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    env_file: ./gitwhy-dashboard/.env.local
    depends_on:
      - api
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: gitwhy
      POSTGRES_PASSWORD: gitwhy
      POSTGRES_DB: gitwhy_dev
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U gitwhy"]
      interval: 10s
      timeout: 5s
      retries: 5
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s

volumes:
  postgres_data:
```

---

## 5. CI/CD — GitHub Actions

### gitwhy-mcp npm package (.github/workflows/publish.yml)

```yaml
name: Publish npm package

on:
  push:
    tags:
      - 'v*'

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          registry-url: 'https://registry.npmjs.org'

      - run: npm ci
      - run: npm test

      - name: Publish to npm
        run: npm publish
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

### Go binary build and release (.github/workflows/release.yml)

```yaml
name: Release Go binaries

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    strategy:
      matrix:
        include:
          - os: ubuntu-latest
            goos: linux
            goarch: amd64
          - os: ubuntu-latest
            goos: linux
            goarch: arm64
          - os: macos-latest
            goos: darwin
            goarch: amd64
          - os: macos-latest
            goos: darwin
            goarch: arm64
          - os: windows-latest
            goos: windows
            goarch: amd64

    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          go build -ldflags "-X main.version=${{ github.ref_name }}" \
            -o gitwhy-${{ matrix.goos }}-${{ matrix.goarch }} \
            ./cmd/main.go

      - name: Upload to Release
        uses: softprops/action-gh-release@v1
        with:
          files: gitwhy-${{ matrix.goos }}-${{ matrix.goarch }}
```

### Cloud API deploy (.github/workflows/deploy-api.yml)

```yaml
name: Deploy Cloud API

on:
  push:
    branches: [main]
    paths:
      - 'gitwhy-cloud/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: superfly/flyctl-actions/setup-flyctl@master

      - name: Run tests
        run: |
          cd gitwhy-cloud
          go test ./...

      - name: Deploy to Fly.io
        run: flyctl deploy --app gitwhy-api
        env:
          FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}
```

---

## 6. Database Migrations

```bash
# Dùng golang-migrate
# Install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Tạo migration mới
migrate create -ext sql -dir db/migrations -seq add_embedding_column

# Run migrations
migrate -path db/migrations -database $DATABASE_URL up

# Rollback
migrate -path db/migrations -database $DATABASE_URL down 1
```

Ví dụ migration file:
```sql
-- db/migrations/000001_create_contexts.up.sql
CREATE TABLE contexts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    local_id VARCHAR(50) UNIQUE NOT NULL,
    user_id UUID NOT NULL,
    -- ... fields từ 07-data-models.md
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- db/migrations/000001_create_contexts.down.sql
DROP TABLE IF EXISTS contexts;
```

---

## 7. Monitoring

| Tool | Mục đích |
|------|---------|
| Fly.io metrics | CPU, RAM, request latency |
| Sentry | Error tracking (Go API + Next.js) |
| Uptime Robot | Endpoint health checks mỗi 5 phút |
| Cloudflare Analytics | CDN + DNS traffic |

### Health check endpoint

```
GET /health

Response 200:
{
  "status": "ok",
  "version": "0.1.5",
  "db": "ok",
  "redis": "ok",
  "timestamp": "2026-06-20T10:30:00Z"
}
```

---

## 8. Secrets Management

| Secret | Storage |
|--------|---------|
| Database credentials | Fly.io secrets |
| Redis URL | Fly.io secrets |
| GitHub App private key | Fly.io secrets |
| NPM token | GitHub Actions secrets |
| Sentry DSN | Fly.io secrets |

```bash
# Set Fly.io secret
flyctl secrets set DATABASE_URL="postgresql://..." --app gitwhy-api
flyctl secrets set GITHUB_APP_PRIVATE_KEY="$(cat github-app.pem)" --app gitwhy-api
```

---

*Cập nhật lần cuối: 2026-06-20*
