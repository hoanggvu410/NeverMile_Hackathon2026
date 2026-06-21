# FRONTEND-PLAN.md — GitWhy Frontend Implementation Plan

> Next.js 14 (App Router) · TanStack Query v5 · Tailwind CSS v3 · TypeScript  
> Hai web app: `app.gitwhy.dev` (dashboard) và `gitwhy.dev` (product site — Framer, không dev thêm)

---

## 1. Tech Stack — Web Dashboard (`app.gitwhy.dev`)

| Layer | Library | Version | Ghi chú |
|-------|---------|---------|--------|
| Framework | Next.js (App Router) | 14+ | Server + Client components |
| Data fetching | TanStack Query | v5 | Cache, mutations, optimistic updates |
| Auth state | Zustand | v4 | User info, API key, plan |
| Forms | React Hook Form + Zod | latest | Validation |
| HTTP client | Axios | latest | Instance với auth interceptor |
| UI primitives | Radix UI | latest | Dialog, DropdownMenu, Tooltip |
| Icons | Lucide React | latest | Consistent icon set |
| Graph visualization | React Flow | v11 | Context Graph (v0.2) |
| Code highlight | Shiki | latest | Syntax highlight trong context detail |
| Tailwind | Tailwind CSS v3 | 3.x | Config theo DESIGN.md theme |
| TypeScript | TypeScript | 5.x | Strict mode |
| Testing | Vitest + Testing Library | latest | Unit + component tests |

---

## 2. Theme Setup

```ts
// tailwind.config.ts
import type { Config } from 'tailwindcss'

const config: Config = {
  darkMode: 'class',
  content: ['./src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
      },
      fontSize: {
        // Base body = 14px (DESIGN.md spec)
        'base': ['14px', { lineHeight: '1.6' }],
        'sm': ['13px', { lineHeight: '1.6' }],
        'xs': ['12px', { lineHeight: '1.5' }],
      },
      colors: {
        primary: {
          50:  '#F5F3FF',
          100: '#EDE9FE',
          600: '#7C3AED',
          700: '#6D28D9',
        },
      },
    },
  },
  plugins: [],
}
export default config
```

```css
/* src/app/globals.css */
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&display=swap');
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  body {
    @apply text-gray-900 text-base leading-relaxed font-sans;
  }
}

@layer components {
  .btn-primary {
    @apply inline-flex items-center gap-2 px-4 py-2 bg-violet-600 hover:bg-violet-700 text-white text-sm font-medium rounded-lg transition-colors;
  }
  .btn-secondary {
    @apply inline-flex items-center gap-2 px-4 py-2 border border-gray-300 hover:bg-gray-50 text-gray-700 text-sm font-medium rounded-lg transition-colors;
  }
  .btn-ghost {
    @apply p-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition-colors;
  }
  .card {
    @apply bg-white border border-gray-200 rounded-xl p-6 shadow-sm;
  }
}
```

---

## 3. Directory Structure

```
src/
├── app/
│   ├── layout.tsx              ← Root layout (font, auth provider)
│   ├── page.tsx                ← Redirect → /dashboard
│   ├── auth/
│   │   ├── login/page.tsx      ← GitHub OAuth login
│   │   └── callback/page.tsx   ← OAuth callback handler
│   └── dashboard/
│       ├── layout.tsx          ← App shell (sidebar + main)
│       ├── page.tsx            ← Redirect → /dashboard/contexts
│       ├── contexts/
│       │   ├── page.tsx        ← Context list
│       │   └── [id]/page.tsx   ← Context detail
│       ├── search/page.tsx     ← Search UI
│       ├── api-keys/page.tsx   ← API key management
│       ├── team/page.tsx       ← Team settings
│       └── graph/page.tsx      ← Context graph (v0.2)
├── components/
│   ├── layout/
│   │   ├── Sidebar.tsx
│   │   └── AppShell.tsx
│   ├── contexts/
│   │   ├── ContextCard.tsx
│   │   ├── ContextDetail.tsx
│   │   ├── ContextBadge.tsx
│   │   └── PostPRModal.tsx
│   ├── search/
│   │   ├── SearchBar.tsx
│   │   └── SearchResults.tsx
│   ├── api-keys/
│   │   ├── ApiKeyList.tsx
│   │   └── CreateKeyModal.tsx
│   ├── team/
│   │   ├── MemberList.tsx
│   │   └── InviteModal.tsx
│   ├── graph/
│   │   └── ContextGraph.tsx    ← React Flow wrapper (v0.2)
│   └── ui/
│       ├── Button.tsx
│       ├── Badge.tsx
│       ├── Modal.tsx
│       ├── EmptyState.tsx
│       └── LoadingSpinner.tsx
├── lib/
│   ├── api.ts                  ← Axios instance + interceptors
│   ├── auth.ts                 ← GitHub OAuth helpers
│   └── utils.ts                ← timeago, truncate, etc.
├── store/
│   └── auth.ts                 ← Zustand store: user, apiKey, plan
├── hooks/
│   ├── useContexts.ts          ← TanStack Query hooks
│   ├── useSearch.ts
│   ├── useApiKeys.ts
│   └── useTeam.ts
└── types/
    └── index.ts                ← GitWhyContext, User, ApiKey, Team types
```

---

## 4. API Layer

```ts
// src/lib/api.ts
import axios from 'axios'
import { useAuthStore } from '@/store/auth'

export const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'https://api.gitwhy.dev',
})

// Auto-inject API key
apiClient.interceptors.request.use((config) => {
  const apiKey = useAuthStore.getState().apiKey
  if (apiKey) {
    config.headers.Authorization = `Bearer ${apiKey}`
  }
  return config
})

// Handle 401 → redirect login
apiClient.interceptors.response.use(
  (res) => res,
  (error) => {
    if (error.response?.status === 401) {
      window.location.href = '/auth/login'
    }
    return Promise.reject(error)
  }
)
```

---

## 5. TanStack Query Hooks

```ts
// src/hooks/useContexts.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiClient } from '@/lib/api'
import type { GitWhyContext } from '@/types'

export function useContexts(filters?: { domain?: string; published?: boolean }) {
  return useQuery({
    queryKey: ['contexts', filters],
    queryFn: async () => {
      const { data } = await apiClient.get<{ results: GitWhyContext[] }>('/v1/contexts', {
        params: filters,
      })
      return data.results
    },
    staleTime: 30_000,
  })
}

export function useContext(id: string) {
  return useQuery({
    queryKey: ['context', id],
    queryFn: async () => {
      const { data } = await apiClient.get<{ context: GitWhyContext }>(`/v1/contexts/${id}`)
      return data.context
    },
    enabled: !!id,
  })
}

export function useSyncContext() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (contextId: string) => {
      const { data } = await apiClient.post('/v1/contexts/sync', {
        context_ids: [contextId],
      })
      return data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['contexts'] })
    },
  })
}

export function usePostPR() {
  return useMutation({
    mutationFn: async ({
      contextId,
      repo,
      prNumber,
    }: {
      contextId: string
      repo: string
      prNumber: number
    }) => {
      const { data } = await apiClient.post('/v1/pr/comment', {
        context_local_id: contextId,
        repo,
        pr_number: prNumber,
      })
      return data
    },
  })
}
```

---

## 6. Auth Store (Zustand)

```ts
// src/store/auth.ts
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthState {
  user: { id: string; email: string; github_login?: string } | null
  apiKey: string | null
  plan: 'free' | 'team'
  setAuth: (user: AuthState['user'], apiKey: string, plan: 'free' | 'team') => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      apiKey: null,
      plan: 'free',
      setAuth: (user, apiKey, plan) => set({ user, apiKey, plan }),
      logout: () => set({ user: null, apiKey: null, plan: 'free' }),
    }),
    { name: 'gitwhy-auth' }
  )
)
```

---

## 7. Key Components

### ContextCard

```tsx
// src/components/contexts/ContextCard.tsx
'use client'

import { GitWhyContext } from '@/types'
import Badge from '@/components/ui/Badge'
import { timeago } from '@/lib/utils'

export function ContextCard({ context, onPostPR, onSync, onDelete }: {
  context: GitWhyContext
  onPostPR: (id: string) => void
  onSync: (id: string) => void
  onDelete: (id: string) => void
}) {
  return (
    <div className="card hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-sm font-semibold text-gray-900">
              🧠 {context.prompt.slice(0, 80)}{context.prompt.length > 80 ? '...' : ''}
            </span>
            {context.synced && <Badge variant="synced" />}
            {context.published && <Badge variant="published" />}
          </div>
          <div className="mt-1.5 text-xs text-gray-500 flex items-center gap-2 flex-wrap">
            <code className="bg-gray-100 px-1.5 py-0.5 rounded text-xs">
              {context.domain}/{context.topic}
            </code>
            <span>·</span>
            <span>{timeago(context.timestamp)}</span>
            <span>·</span>
            <span>{context.commits.length} commit(s)</span>
            <span>·</span>
            <span>{context.files.length} files</span>
          </div>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <a href={`/dashboard/contexts/${context.id}`} className="btn-secondary text-xs">
            View
          </a>
          <button onClick={() => onPostPR(context.id)} className="btn-ghost text-xs">
            Post PR
          </button>
          {!context.synced && (
            <button onClick={() => onSync(context.id)} className="btn-ghost text-xs">
              Sync
            </button>
          )}
          <button onClick={() => onDelete(context.id)} className="btn-ghost text-xs text-red-500">
            Delete
          </button>
        </div>
      </div>
    </div>
  )
}
```

---

## 8. Routing & Page Structure

| Route | Component | Notes |
|-------|-----------|-------|
| `/auth/login` | LoginPage | GitHub OAuth button |
| `/auth/callback` | CallbackPage | Handle OAuth token |
| `/dashboard/contexts` | ContextsPage | List + filters |
| `/dashboard/contexts/[id]` | ContextDetailPage | Full context view |
| `/dashboard/search` | SearchPage | Semantic search |
| `/dashboard/api-keys` | ApiKeysPage | Create/revoke |
| `/dashboard/team` | TeamPage | Members + billing |
| `/dashboard/graph` | GraphPage | React Flow (v0.2) |

---

## 9. Deployment (Fly.io)

```toml
# fly.toml
app = "gitwhy-dashboard"
primary_region = "sin"  # Singapore — gần SEA users

[build]
  dockerfile = "Dockerfile"

[env]
  PORT = "3000"
  NODE_ENV = "production"

[http_service]
  internal_port = 3000
  force_https = true

[[vm]]
  memory = "512mb"
  cpu_kind = "shared"
  cpus = 1
```

---

## 10. Performance Targets

| Metric | Target |
|--------|--------|
| FCP (First Contentful Paint) | < 1.5s |
| Context list (50 items) | < 200ms render |
| Search results | < 500ms (cloud) / < 100ms (cached) |
| Graph render (100 nodes) | < 1s |
| Lighthouse score | > 90 |

---

*Cập nhật lần cuối: 2026-06-20*
