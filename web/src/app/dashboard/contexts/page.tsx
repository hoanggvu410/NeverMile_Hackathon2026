"use client";

import { useState } from "react";
import { useContexts } from "@/hooks/useGitWhy";
import { FilterBar } from "@/components/contexts/FilterBar";
import { ContextCard } from "@/components/contexts/ContextCard";
import { CardSkeleton } from "@/components/ui/Skeleton";
import { EmptyState } from "@/components/ui/EmptyState";
import { IconContexts } from "@/components/ui/icons";

export default function ContextsPage() {
  const [domain, setDomain] = useState("");
  const { data: contexts, isLoading } = useContexts(domain);

  return (
    <main className="h-full overflow-y-auto px-6 py-6 lg:px-8">
      <div className="mb-6">
        <h1 className="text-[26px] font-light text-ink">Ngữ cảnh</h1>
        <p className="mt-1 text-[13px] text-ink-muted">
          Mọi quyết định đã lưu, theo domain &amp; topic.
        </p>
      </div>

      <div className="mb-6">
        <FilterBar active={domain} onChange={setDomain} />
      </div>

      {isLoading ? (
        <div className="grid grid-cols-1 gap-3.5 lg:grid-cols-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <CardSkeleton key={i} />
          ))}
        </div>
      ) : contexts && contexts.length > 0 ? (
        <div className="grid grid-cols-1 gap-3.5 pb-6 lg:grid-cols-2">
          {contexts.map((ctx, i) => (
            <ContextCard key={ctx.id} ctx={ctx} index={i} />
          ))}
        </div>
      ) : (
        <EmptyState
          icon={<IconContexts className="h-6 w-6" />}
          title="Chưa có ngữ cảnh trong domain này"
          description="Lưu ngữ cảnh từ agent để thấy nó xuất hiện ở đây."
        />
      )}
    </main>
  );
}
