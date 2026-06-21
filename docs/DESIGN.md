# DESIGN.md — GitWhy Web Dashboard UI Specification

> Tài liệu này mô tả toàn bộ giao diện của **GitWhy Web Dashboard** (`app.gitwhy.dev`) — nơi dev và team quản lý contexts, API keys, và team settings. Dùng Tailwind CSS v3. Mỗi màn hình có mô tả layout, component, và class cụ thể.

---

## 1. Design System

### 1.1 Màu sắc (Tailwind palette)

| Token | Tailwind class | Hex | Dùng cho |
|-------|---------------|-----|---------|
| Primary | `violet-600` | #7C3AED | Button chính, link active, accent |
| Primary hover | `violet-700` | #6D28D9 | Button hover |
| Primary light | `violet-50` | #F5F3FF | Background chip, selected row |
| Secondary | `slate-600` | #475569 | Text phụ, icon |
| Success | `green-600` | #16A34A | Synced status, published badge |
| Warning | `amber-500` | #F59E0B | Pending sync, cache miss |
| Danger | `red-600` | #DC2626 | Error, delete |
| Info | `sky-500` | #0EA5E9 | Team plan badge, graph links |
| Background | `gray-50` | #F9FAFB | App background |
| Surface | `white` | #FFFFFF | Card, sidebar, modal |
| Border | `gray-200` | #E5E7EB | Đường kẻ, divider |
| Text primary | `gray-900` | #111827 | Heading, body text |
| Text secondary | `gray-500` | #6B7280 | Label, placeholder, meta |

### 1.2 Typography

```
Font: Inter (Google Fonts) — font-size 14px global
Heading 1: text-2xl font-semibold text-gray-900
Heading 2: text-xl font-semibold text-gray-900
Heading 3: text-base font-semibold text-gray-900
Body: text-sm text-gray-700   (14px, line-height 1.6)
Label: text-xs font-medium text-gray-500 uppercase tracking-wide
Caption: text-xs text-gray-400
Code: font-mono text-sm bg-gray-100 px-1 rounded
```

**Lưu ý quan trọng (từ PRD v0.2):**
- Font size **14px** toàn app — không dùng 12px cho body text
- Line-height **1.6** cho tất cả body text — không 1.4
- Padding cards tối thiểu **p-6** — không p-4 cho main cards

### 1.3 Components tái dùng

#### Button

```html
<!-- Primary -->
<button class="inline-flex items-center gap-2 px-4 py-2 bg-violet-600 hover:bg-violet-700 text-white text-sm font-medium rounded-lg transition-colors">
  Label
</button>

<!-- Secondary (outline) -->
<button class="inline-flex items-center gap-2 px-4 py-2 border border-gray-300 hover:bg-gray-50 text-gray-700 text-sm font-medium rounded-lg transition-colors">
  Label
</button>

<!-- Danger -->
<button class="inline-flex items-center gap-2 px-4 py-2 bg-red-600 hover:bg-red-700 text-white text-sm font-medium rounded-lg transition-colors">
  Delete
</button>

<!-- Ghost -->
<button class="p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors">
  <!-- icon 16x16 -->
</button>
```

#### Badge

```html
<!-- Synced -->
<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-700">Synced</span>
<!-- Pending -->
<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-amber-100 text-amber-700">Pending</span>
<!-- Published -->
<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-violet-100 text-violet-700">Published</span>
<!-- Free -->
<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-600">Free</span>
<!-- Team -->
<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-sky-100 text-sky-700">Team</span>
<!-- Cache Hit -->
<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-600">⚡ Cache Hit</span>
```

#### Card

```html
<div class="bg-white border border-gray-200 rounded-xl p-6 shadow-sm">
  <!-- content — padding p-6 bắt buộc -->
</div>
```

---

## 2. Layout

### 2.1 App Shell

```
┌─────────────────────────────────────────────────────────────────┐
│ Sidebar (w-64, fixed left)    │  Main content (flex-1)          │
│                               │  max-w-5xl mx-auto px-6 py-8   │
│  [GitWhy logo]                │                                 │
│  ─────────────────            │                                 │
│  📋 Contexts                  │                                 │
│  🔍 Search                    │                                 │
│  👥 Team                      │                                 │
│  🔑 API Keys                  │                                 │
│  📊 Graph (v0.2)              │                                 │
│  ─────────────────            │                                 │
│  [Plan badge: Free/Team]      │                                 │
│  [User avatar + email]        │                                 │
│  [Settings]                   │                                 │
└───────────────────────────────┴─────────────────────────────────┘
```

```html
<div class="flex h-screen bg-gray-50">
  <!-- Sidebar -->
  <aside class="w-64 bg-white border-r border-gray-200 flex flex-col">
    <!-- Logo -->
    <div class="px-6 py-5 border-b border-gray-200">
      <span class="text-lg font-semibold text-gray-900">GitWhy</span>
    </div>
    <!-- Nav -->
    <nav class="flex-1 px-3 py-4 space-y-1">
      <!-- Nav items -->
    </nav>
    <!-- User footer -->
    <div class="px-4 py-4 border-t border-gray-200">
      <!-- Plan badge + user info -->
    </div>
  </aside>

  <!-- Main -->
  <main class="flex-1 overflow-auto">
    <div class="max-w-5xl mx-auto px-6 py-8">
      <!-- Page content -->
    </div>
  </main>
</div>
```

---

## 3. Màn hình: Context List

**Route:** `/dashboard/contexts`

```
┌─────────────────────────────────────────────────────────────────┐
│ Contexts (47)                           [Sync pending (3)]  [Search] │
│                                                                 │
│ Filter: [All] [Published] [Pending sync]   Domain: [All ▾]     │
│ ─────────────────────────────────────────────────────────────   │
│ [Context card]                                                  │
│   🧠 "Migrate JWT từ HS256 sang RS256"          [Synced] [Published] │
│   backend/auth · jwt-migration · 2 giờ trước                   │
│   Commits: a1b2c3d · Files: 3 files                            │
│   [View] [Post PR] [Sync] [Delete]                              │
│ ─────────────────────────────────────────────────────────────   │
│ [Context card]                                                  │
│   🧠 "Remove Kafka from messaging pipeline"     [Pending]       │
│   infrastructure · kafka-removal · 1 ngày trước                │
│   ...                                                           │
└─────────────────────────────────────────────────────────────────┘
```

#### Context Card

```html
<div class="bg-white border border-gray-200 rounded-xl p-6 shadow-sm hover:shadow-md transition-shadow">
  <div class="flex items-start justify-between gap-4">
    <div class="flex-1 min-w-0">
      <div class="flex items-center gap-2 flex-wrap">
        <span class="text-sm font-semibold text-gray-900 truncate">
          🧠 {{ context.prompt | truncate(80) }}
        </span>
        <!-- Badges -->
        <span class="badge-synced">Synced</span>
        <span class="badge-published">Published</span>
      </div>
      <div class="mt-1 text-xs text-gray-500 flex items-center gap-2">
        <span class="font-mono bg-gray-100 px-1.5 py-0.5 rounded">{{ context.domain }}/{{ context.topic }}</span>
        <span>·</span>
        <span>{{ context.timestamp | timeago }}</span>
        <span>·</span>
        <span>{{ context.commits.length }} commit(s)</span>
        <span>·</span>
        <span>{{ context.files.length }} files</span>
      </div>
    </div>
    <!-- Actions -->
    <div class="flex items-center gap-2 shrink-0">
      <button class="btn-ghost">View</button>
      <button class="btn-ghost">Post PR</button>
      <button class="btn-ghost text-red-500">Delete</button>
    </div>
  </div>
</div>
```

---

## 4. Màn hình: Context Detail

**Route:** `/dashboard/contexts/{id}`

```
┌─────────────────────────────────────────────────────────────────┐
│ ← Back to Contexts                                              │
│                                                                 │
│ 🧠 Migrate JWT từ HS256 sang RS256            [Synced][Published]│
│ backend/auth · jwt-migration · 20 Jun 2026 10:30 AM            │
│ agent: claude-code · model: claude-opus-4-6                    │
│ commits: a1b2c3d · files: app/core/security.py (+2)            │
│                                                                 │
│ ── Reasoning ───────────────────────────────────────────────── │
│ Cần asymmetric key để services khác verify token mà không      │
│ cần private key...                                              │
│                                                                 │
│ ── Decisions ───────────────────────────────────────────────── │
│ • Chọn RS256: public key có thể distribute an toàn             │
│ • Key size 2048-bit: đủ secure, không quá chậm                 │
│                                                                 │
│ ── Rejected Alternatives ───────────────────────────────────── │
│ • HS256: symmetric key phải share với tất cả services          │
│ • ES256: cần thêm library, team chưa quen                      │
│                                                                 │
│ [Post to PR #42 ▾]  [Copy ID]  [Share Link]  [Delete]         │
└─────────────────────────────────────────────────────────────────┘
```

---

## 5. Màn hình: Search

**Route:** `/dashboard/search`

```
┌─────────────────────────────────────────────────────────────────┐
│ Search Contexts                                                 │
│                                                                 │
│ [🔍 tại sao bỏ Kafka...                              ] [Search] │
│  Domain filter: [infrastructure ▾]                             │
│                                                                 │
│ Results (3)                                    [⚡ Cache Hit]   │
│ ─────────────────────────────────────────────────────────────   │
│ 94% · infrastructure/kafka-removal · 10 Jun 2026               │
│ Remove Kafka from messaging pipeline                            │
│ "Switched to SQS for cost savings and simpler ops..."          │
│                                                                 │
│ 87% · infrastructure/sqs-migration · 15 Jun 2026               │
│ Configure SQS queues and DLQ                                    │
│ ...                                                             │
└─────────────────────────────────────────────────────────────────┘
```

---

## 6. Màn hình: API Keys

**Route:** `/dashboard/api-keys`

```
┌─────────────────────────────────────────────────────────────────┐
│ API Keys                                      [+ Create New Key] │
│                                                                 │
│ ⚠️  API keys are shown only once. Store them securely.          │
│                                                                 │
│ Name              Last used        Created      Actions         │
│ ─────────────────────────────────────────────────────────────  │
│ Claude Code Key   2 hours ago      Jun 1, 2026  [Revoke]       │
│ CI/CD Key         Never            Jun 5, 2026  [Revoke]       │
│ ─────────────────────────────────────────────────────────────  │
│ [Free] 1 of 3 keys used                                        │
└─────────────────────────────────────────────────────────────────┘
```

#### Create Key Modal

```html
<div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
  <div class="bg-white rounded-2xl p-8 w-full max-w-md shadow-xl">
    <h2 class="text-lg font-semibold text-gray-900 mb-4">Create API Key</h2>

    <label class="text-xs font-medium text-gray-500 uppercase tracking-wide">Key name</label>
    <input class="mt-1 w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-violet-500"
           placeholder="e.g. Claude Code Key" />

    <div class="mt-6 flex gap-3 justify-end">
      <button class="btn-secondary">Cancel</button>
      <button class="btn-primary">Create Key</button>
    </div>
  </div>
</div>
```

---

## 7. Màn hình: Context Graph (v0.2)

**Route:** `/dashboard/graph`

Dùng D3.js hoặc React Flow để visualize graph.

```
┌─────────────────────────────────────────────────────────────────┐
│ Context Graph                    [Domain: All ▾]  [Time: 30d ▾] │
│                                                                 │
│  ┌──────────┐     led_to     ┌──────────────────┐              │
│  │ kafka-   │  ──────────►  │ sqs-migration    │              │
│  │ removal  │  0.87 sim     │ Jun 15            │              │
│  └──────────┘               └──────────────────┘              │
│       │                              │                          │
│   led_to                         led_to                        │
│       ▼                              ▼                          │
│  ┌──────────┐               ┌──────────────────┐              │
│  │ cost-    │               │ dlq-setup        │              │
│  │ analysis │               │ Jun 16            │              │
│  └──────────┘               └──────────────────┘              │
│                                                                 │
│ Click node to view context detail                               │
└─────────────────────────────────────────────────────────────────┘
```

---

## 8. Màn hình: Team Settings

**Route:** `/dashboard/team`

```
┌─────────────────────────────────────────────────────────────────┐
│ Team Settings                                                   │
│                                                                 │
│ ── Members (4) ─────────────────────────────────────────────── │
│ quan@gitwhy.dev        owner   Joined Jun 1    –               │
│ alice@team.dev         member  Joined Jun 5    [Remove]        │
│ bob@team.dev           member  Joined Jun 10   [Remove]        │
│                                                                 │
│ [+ Invite Member]                                               │
│                                                                 │
│ ── Plan ─────────────────────────────────────────────────────── │
│ Team Plan · $20/month · Renews Jul 1, 2026                     │
│ Unlimited sync · Unlimited publish · PR bot                     │
│ [Manage Billing]                                                │
└─────────────────────────────────────────────────────────────────┘
```

---

## 9. Empty States

```html
<!-- Contexts — chưa có gì -->
<div class="text-center py-16">
  <div class="text-4xl mb-4">🧠</div>
  <h3 class="text-base font-semibold text-gray-900 mb-2">No contexts yet</h3>
  <p class="text-sm text-gray-500 max-w-sm mx-auto mb-6 leading-relaxed">
    Install gitwhy-mcp and ask your AI agent to save context after the next commit.
  </p>
  <a href="https://docs.gitwhy.dev" class="btn-primary">View Setup Guide →</a>
</div>
```

---

## 10. Responsive

| Breakpoint | Layout |
|-----------|--------|
| `< 768px` (mobile) | Sidebar collapse → hamburger menu |
| `768-1024px` (tablet) | Sidebar icon-only (w-16) |
| `> 1024px` (desktop) | Full sidebar (w-64) |

---

*Cập nhật lần cuối: 2026-06-20*
