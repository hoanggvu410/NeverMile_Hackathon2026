# GitWhy — Product

**GitWhy lets AI agents query why decisions were made.**

---

## Problem

AI coding agents are stateless. Every time an agent starts a new session, it rediscovers context from scratch — crawling files, re-reading code, re-reasoning decisions that were already made. This burns tokens on work that was already done.

The memory exists. It lives in commit messages, PR descriptions, and chat histories. The problem is that **no agent can query it usefully.**

### The cost of stateless agents — sourced

| Stat | Source |
|---|---|
| Developers experience **12–15 major context switches daily**. At 23 minutes recovery time per switch: **>4.5 hours of lost focus every day** | DEV Community, 2026 |
| Developers spend **27% of their workday** waiting for or switching between tools — over 2 hours of pure friction daily | Medium/Munyao, 2025 |
| An unconstrained AI agent costs **$5–8 per task** on a software engineering issue | Stevens Online, 2026 |
| Agentic systems cost **5–25x more per task** than non-agentic alternatives due to retry loops and context reloading | TechAhead, 2026 |
| **73% of companies exceeded their AI budget** despite token prices dropping 67% — volume from agent loops drives cost, not unit price | Pebblous, 2026 |
| Multi-agent systems spend a **2:1 input-to-output token ratio** — most tokens go to communicating context, not generating output | arXiv Tokenomics, 2026 |

### The mental model

**Without GitWhy:** new employee, no onboarding docs, figures everything out by walking around asking questions.

**With GitWhy:** same employee, but there's a structured wiki with every decision already captured. They check the wiki first. The LLM stops being a search engine and starts being a reasoner.

---

## Who It's For

**Primary user: developers building or using AI coding agents.**

The value is speed and cost reduction, not knowledge management. The person who benefits most is someone who runs Claude Code or Cursor daily and notices their agent re-reading the same files every session.

| User | Pain | What GitWhy gives them |
|---|---|---|
| **Developer** | Agent re-figures out decisions every session | Agent queries memory before touching LLM |
| **PR Reviewer** | Only sees what changed, never why | Full reasoning chain on the PR via gitwhy-bot |
| **Team Lead** | Decisions live in individuals' heads | Shared queryable graph of all team decisions |
| **AI Agent** | Stateless — no context from prior sessions | Structured memory queryable via MCP tool call |

---

## Why Existing Solutions Fall Short

The "agent memory" category exists. Mem0, Zep, and Letta all store context and retrieve it. They are not the same problem.

| | What they store | How they retrieve | What's missing |
|---|---|---|---|
| **Mem0 / Zep** | Facts, summaries | Similarity — "what's related to X?" | Can't answer "what's downstream of X?" |
| **Git history / PR descriptions** | What changed | Human reads linearly | No agent-queryable interface. No rejected alternatives. |
| **Markdown history file** | Chronological log | Agent reads entire file | Grows unbounded. No graph. No structured fields. |
| **GitWhy** | Decisions + rejected alternatives | Graph traversal — "what caused X? what breaks if X changes?" | — |

**Mem0 remembers facts. GitWhy remembers decisions — why they were made, what was rejected, and what breaks if they change.**

The core distinction: similarity search answers "what is similar to X?" Graph traversal on typed edges answers "what is causally downstream of X?" These are different questions. Only GitWhy answers the second one.

---

## How It Works

### The mechanism

```
Developer codes with AI agent
    ↓
git commit
    ↓
post-commit hook fires gitwhy_save automatically
    ↓
Structured context written to .git/gitwhy/contexts/{id}.md
    ↓
Node inserted into context graph (SQLite), typed edges link related decisions
    ↓

Later — agent queries:
gitwhy_search("why was Kafka removed?")
    ↓
Check semantic cache → cache hit → return in <50ms, $0.00
                     → cache miss → graph traversal → decision chain returned in ~3s, ~$0.01
```

### Three components

| Component | Where | Tech |
|---|---|---|
| **MCP Server** | Local — `npm: gitwhy-mcp` | Node.js wrapper over Go binary |
| **Web Dashboard** | `app.gitwhy.dev` | Next.js 14 App Router |
| **Cloud Backend** | `api.gitwhy.dev` | Go REST API + PostgreSQL |

### 8 MCP Tools

| Tool | What it does |
|---|---|
| `gitwhy_save` | Writes structured context to `.git/gitwhy/contexts/`, auto-links HEAD commit |
| `gitwhy_get` | Retrieve a context by ID |
| `gitwhy_search` | Query the context graph — returns decision chain |
| `gitwhy_list` | Browse contexts by domain/topic |
| `gitwhy_status` | Check setup state — git repo, API key, pending syncs |
| `gitwhy_sync` | Upload local contexts to cloud (Free: 20/month) |
| `gitwhy_publish` | Share contexts with team (Team plan only) |
| `gitwhy_post_pr` | Post context summary as GitHub PR comment |

### What makes the graph different from RAG

Every context node is linked to related decisions via **typed edges**:

| Edge | Meaning |
|---|---|
| `CAUSED_BY` | Decision A directly triggered decision B |
| `CONSTRAINED_BY` | Decision B had to work within bounds set by A |
| `INVALIDATES` | Decision B makes decision A obsolete |
| `CONTRADICTS` | B conflicts with A — fires an alert |
| `DEPENDS_ON` | B only holds if an assumption in A still holds |

This means an agent can ask: "if I change this decision, what else breaks?" RAG cannot answer that. Graph traversal can.

---

## 5W 1H

**Who:** AI coding agents — Claude Code, Cursor, Windsurf, Cline — and the developers who use them

**What:** A queryable memory layer that stores reasoning, decisions, and rejected alternatives from every coding session, linked to git commits

**When:** Automatically on every git commit — and retrieved instantly before any new task starts

**Where:** Locally in `.git/gitwhy/` (offline-first) — optionally synced to cloud for team sharing

**Why:** Agents are stateless — every session starts from zero. GitWhy makes prior decisions queryable so agents stop rediscovering what already happened.

**How:** Post-commit hook saves structured context → typed graph links related decisions → semantic cache returns repeated queries at $0.00

---

## Business Model

| Tier | Price | Key Limits |
|---|---|---|
| **Free** | $0 | 1 repo, 20 cloud syncs/month, unlimited local, 3 API keys |
| **Team** | $20/month | Unlimited sync + publish + PR bot + team search |
| **Enterprise** | TBD | Self-hosted, SSO, audit logs |

**GTM strategy:** Developer uses it locally first — real value, not slideware. Team buys when they want to share the graph across members. The trigger for upgrade is `gitwhy_publish`.

All local features work completely offline with no API key. Cloud features require signup.

---

## v0.2 Scope — What We're Building

v0.1 shipped the foundation: 8 MCP tools, cloud sync, PR bot, web dashboard.

v0.2 is the hackathon build. Four sprints:

| Sprint | Feature | Why it matters |
|---|---|---|
| 1 | **Auto-save hook** | Zero friction — developer does nothing after commit |
| 2 | **Semantic cache** | Repeated queries cost $0.00 — the demo's second beat |
| 3 | **Context graph** | Causal chains instead of flat search results — the core differentiator |
| 4 | **UI polish** | Dashboard reflects the product quality |

**What's out of scope for v0.2:** GitLab/Bitbucket integration, self-hosted, SSO, Jira/Slack connectors, VS Code extension.

---

## Success Metrics

| Metric | Target |
|---|---|
| Demo: "tại sao bỏ Kafka" → decision chain | < 3 seconds |
| Demo: same query repeated | < 50ms, $0.00 |
| `gitwhy_save` (local) | < 100ms |
| `gitwhy_search` (cloud) | < 3s |
| Semantic cache hit response | < 50ms |
| Dashboard FCP | < 1.5s |

---

## Use Case — End to End

### Scenario

A developer uses Claude Code to remove Kafka from the messaging pipeline. Two weeks later a teammate asks why.

---

### Day 1 — Developer commits

Developer prompts Claude Code:
```
"Remove Kafka from the messaging pipeline, replace with SQS"
```

Claude Code makes the changes and commits. The post-commit hook fires automatically:

```json
{
  "prompt": "Remove Kafka from the messaging pipeline, replace with SQS",
  "reasoning": "Kafka is overkill for current traffic. Costs $800/month. Traffic volume does not justify it.",
  "decisions": "Use SQS — managed, cheaper, team already on AWS.",
  "rejected_alternatives": "RabbitMQ: extra service to maintain. Keep Kafka: $800/month unjustified at current scale.",
  "files": ["services/queue.go", "infra/kafka.tf", "infra/sqs.tf"],
  "commits": ["a1b2c3d"],
  "domain": "infrastructure",
  "topic": "kafka-removal"
}
```

Written to `.git/gitwhy/contexts/cxt_20260601_abc123.md`. Node inserted into `graph.db`. Typed edges link it to a prior cost analysis via `CAUSED_BY`.

---

### Day 15 — Teammate queries

```
gitwhy_search("tại sao bỏ Kafka")
```

MCP embeds the query → checks `semantic.db` → cache miss → queries `graph.db`:

```
kafka-removal  →[CAUSED_BY]→  sqs-migration  →[DEPENDS_ON]→  dlq-setup
```

Returns the full decision chain — why Kafka was removed, what replaced it, how it was configured, what handles failures.

**~3 seconds. ~$0.01.**

Result cached in `semantic.db`.

---

### 5 minutes later — different agent, same question

```
gitwhy_search("why did we remove kafka")
```

Cosine similarity against cached query: **0.94** → cache hit.

**< 50ms. $0.00.**

---

### What each component did

| Component | Role |
|---|---|
| Post-commit hook | Fired `gitwhy_save` automatically — developer did nothing extra |
| `.md` whyspec file | Stored the full reasoning permanently, offline, in the repo |
| `graph.db` | Linked kafka-removal → sqs-migration → dlq-setup via typed edges |
| `semantic.db` | Answered the second query without touching the LLM |
| MCP server | Coordinated everything — the agent called one tool |

---

*Updated: 2026-06-20*
