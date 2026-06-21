# GitWhy Dashboard

Dark glassmorphic web dashboard for GitWhy. Reads real data from the local repo
via a no-auth Go HTTP server — no cloud, no login.

## Run it (two terminals)

**1. Local API server** (wraps `internal/context` store + `internal/graph`):

```bash
# from repo root
go build -o serve.exe ./cmd/serve/
./serve.exe                 # serves http://localhost:7420
./serve.exe -repo /path/to/other/repo   # point at a different git repo
```

**2. Web dashboard:**

```bash
cd web
npm install
npm run dev                 # http://localhost:3000  → redirects to /dashboard
```

Override the API origin with `NEXT_PUBLIC_API_URL` if the server runs elsewhere.

## API surface (`cmd/serve/main.go`)

| Method | Path | Returns |
|---|---|---|
| GET | `/api/status` | repo, branch, context count, pending commits |
| GET | `/api/contexts?domain=&topic=` | context summaries |
| GET | `/api/contexts/:id` | full whyspec |
| GET | `/api/search?q=&limit=` | claim-level graph search results |
| GET | `/api/graph/nodes` | claim nodes (domain, topic, claim text, edge count) |
| GET | `/api/graph/edges` | typed edges |
| GET | `/api/domains` | unique domain list |

## Screens

- **/dashboard** — hero, domain filter chips, context feed, right panel (latest / overview / graph health). Empty repos show a setup checklist.
- **/dashboard/contexts** — full context grid.
- **/dashboard/contexts/[id]** — structured whyspec (sections, files table, commits, raw toggle, copy ID).
- **/dashboard/search** — debounced semantic search over the claim graph.
- **/dashboard/graph** — React Flow claim graph, nodes colored by domain, typed animated edges.

## Stack

Next.js 14 (App Router) · TanStack Query v5 · Tailwind CSS v3 · TypeScript strict ·
React Flow v11 · Framer Motion. Canvas starfield + SVG mountain backdrop, OKLCH design
tokens, `prefers-reduced-motion` respected.
