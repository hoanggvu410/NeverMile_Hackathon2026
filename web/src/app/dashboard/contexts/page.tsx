"use client";

import { useState } from "react";
import { useContexts } from "@/hooks/useGitWhy";
import { FilterBar } from "@/components/contexts/FilterBar";
import { ContextCard } from "@/components/contexts/ContextCard";
import { CardSkeleton } from "@/components/ui/Skeleton";
import { EmptyState } from "@/components/ui/EmptyState";
import { IconContexts } from "@/components/ui/icons";
import { useLanguage } from "@/lib/language";

export default function ContextsPage() {
  const { t } = useLanguage();
  const [domain, setDomain] = useState("");
  const { data: contexts, isLoading } = useContexts(domain);

  return (
    <main className="h-full overflow-y-auto px-6 py-6 lg:px-8">
      <div className="mb-6">
        <h1 className="text-[26px] font-light text-ink">{t("Ngữ cảnh")}</h1>
        <p className="mt-1 text-[13px] text-ink-muted">
          {t("Mọi quyết định đã lưu, theo domain & topic.")}
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
          title={t("Chưa có ngữ cảnh trong domain này")}
          description={t("Lưu ngữ cảnh từ agent để thấy nó xuất hiện ở đây.")}
        />
      )}
    </main>
  );
}
