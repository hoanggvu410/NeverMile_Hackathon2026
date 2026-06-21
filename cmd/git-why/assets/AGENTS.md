# GitWhy Agent Contract

This file defines required agent behaviors for GitWhy2 tools.
Read this before calling any `gitwhy2_*` tool.

---

## GitWhy2 Plan Tripwire Gate (REQUIRED)

**For any non-trivial coding task**, after you have drafted a concrete plan and **before you edit any files**, call `gitwhy2_tripwire`.

This is the plan-stage interrupt gate. It is not optional. Do not substitute `gitwhy2_search` or `gitwhy2_list` for this check — those tools answer questions; this tool catches conflicts with prior decisions.

### When to call it

- You have a concrete plan: you know which files to touch and what to change.
- The user asked you to code, refactor, simplify, or remove something.
- A phrase like "check against gitwhy2", "check memory", or "check prior decisions" appears — that means call `gitwhy2_tripwire`, not `gitwhy2_search`.

### How to build the event object

```json
{
  "event_type": "agent_plan_created",
  "user_request": "<exact user prompt>",
  "agent_plan": "<your concrete plan — what you intend to change and why>",
  "files_likely_touched": ["internal/graph/graph.go", "internal/mcp/server.go"],
  "concepts": ["claim graph", "tripwire", "search fallback"],
  "proposed_changes": ["remove graph search", "replace with markdown fallback"],
  "new_dependencies": [],
  "risk_surfaces": ["graph retrieval", "tripwire interrupt logic"]
}
```

Fill every field you can. Sparse events produce weaker matches.

### How to handle the response

**If `interrupt: true`:**
1. Stop. Do not edit files yet.
2. Show the user the relevant prior claims returned in `candidates`.
3. Ask one of:
   - "A prior decision conflicts with this plan. Revise the plan?"
   - "Continue anyway and override the old decision?"
   - "Mark the old decision as superseded and proceed?"
4. Act on the user's answer before touching any file.

**If `interrupt: false`:**
Proceed with your plan. No prior decisions conflict.

**If the tool is unavailable** (graph not initialized, DB missing):
Report: "gitwhy2_tripwire was unavailable — proceeding cautiously without memory check."
Then continue.

### Tool role summary

| Tool | Purpose | When to use |
|---|---|---|
| `gitwhy2_tripwire` | Plan-stage interrupt gate | Before editing files — mandatory |
| `gitwhy2_search` | Memory Q&A and retrieval | Answering questions about past decisions |
| `gitwhy2_get` | Fetch a specific context by ID | When you have an ID from search/list |
| `gitwhy2_list` | Browse domain/topic hierarchy | Exploring what's saved |
| `gitwhy2_save` | Save new context after a session | After a non-trivial implementation |

---

## GitWhy Save Contract

This file tells AI agents how to write high-quality `gitwhy2_save` payloads.
GitWhy structures and indexes what you write. It cannot improve vague input.

## When to save

- Before ending a session where non-obvious decisions were made
- After choosing between competing approaches
- After any implementation that future agents might undo or contradict

## Field contracts

### `decisions` (required)

Write durable constraint sentences. Each decision must make sense without surrounding context.
Use constraint language: **Use / Do not / Never / Prefer / Always / Avoid**.

```
# BAD
decisions: fixed spacing, updated layout stuff

# GOOD
decisions: |
  - Use 4/8/16/24/32/48/64 spacing scale for all planner UI elements.
  - Do not introduce ad-hoc margin or padding values in planner controls.
  - Prefer gap-based layout over margin on flex containers.
```

### `rejected_alternatives` (optional, include when relevant)

Name the option AND explain why it was rejected. Omitting the reason makes the record useless.

```
# BAD
rejected_alternatives: tried memoization, considered a different DB

# GOOD
rejected_alternatives: |
  - Memoization rejected: cache invalidation on domain changes was complex and caused
    stale reads in tests.
  - Separate DB rejected: would require a migration path we do not have time for
    before the release freeze.
```

### `risks` (optional, include when there are real future hazards)

Frame as trigger → consequence. Tell the next agent what situation would break this decision.

```
# BAD
risks: might break

# GOOD
risks: |
  - If a new planner control adds margin outside the spacing scale, layout will
    diverge visually — check spacing values before any planner UI edit.
  - If the graph DB schema changes, SaveToGraph must be updated before the next
    release or edge linking will silently drop records.
```

### `reasoning` (required)

Explain the trade-off, not what was done. Reference `rejected_alternatives` if relevant.

```
# BAD
reasoning: I implemented the spacing system

# GOOD
reasoning: |
  Chose a fixed spacing scale to prevent layout drift across planner components.
  Freeform spacing was rejected because inconsistency appeared within two PRs.
```

### `what_was_done` (optional)

Implementation summary only. Not reasoning, not decisions — just what changed.

### `domain` / `topic`

Use hierarchical labels so search and browsing are useful:
- `domain`: `frontend/planner`, `backend/auth`, `infra/db`
- `topic`: `spacing-scale`, `jwt-migration`, `graph-schema`

## Anti-patterns to avoid

| Vague | Better |
|---|---|
| `fixed stuff` | `Fixed the spacing by enforcing the 8pt scale` |
| `updated things` | `Updated planner card layout to use gap instead of margin` |
| `seemed good` | `Preferred gap because it avoids margin-collapse edge cases` |
| `various changes` | List each change as a separate decision bullet |
| `none` for risks | Leave blank if genuinely no risk; write a real risk if there is one |

## Minimal viable save (if time-pressed)

Even a minimal save is better than nothing:

```json
{
  "prompt": "exact user request here",
  "reasoning": "chose X because Y, not Z",
  "decisions": "Use X. Do not Y.",
  "domain": "frontend/planner",
  "topic": "spacing-scale"
}
```
