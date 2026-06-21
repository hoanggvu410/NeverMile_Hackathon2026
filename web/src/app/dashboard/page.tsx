"use client";

import { useState } from "react";
import { motion } from "framer-motion";
import { useContexts, useStatus } from "@/hooks/useGitWhy";
import { FilterBar } from "@/components/contexts/FilterBar";
import { ContextCard } from "@/components/contexts/ContextCard";
import { SetupChecklist } from "@/components/contexts/SetupChecklist";
import { KnowledgeHeatmap } from "@/components/dashboard/KnowledgeHeatmap";
import { RightPanel } from "@/components/layout/RightPanel";
import { CardSkeleton } from "@/components/ui/Skeleton";

export default function DashboardPage() {
  const [domain, setDomain] = useState("");
  const { data: status } = useStatus();
  const { data: contexts, isLoading } = useContexts(domain);

  const hasContexts = (contexts?.length ?? 0) > 0;

  return (
    <div className="flex h-full">
      <main className="min-w-0 flex-1 overflow-y-auto px-6 py-6 lg:px-8">
        {/* Hero */}
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, ease: [0.23, 1, 0.32, 1] }}
          className="mb-7 max-w-[60ch]"
        >
          <h1 className="bg-gradient-to-r from-[#cdd9ff] via-[#6f8dff] to-[#8a6bff] bg-clip-text text-[36px] font-light leading-[1.08] text-transparent drop-shadow-[0_2px_24px_rgba(83,58,253,0.35)]">
            Vì sao build cái này?
          </h1>
          <p className="mt-2.5 text-[14px] leading-relaxed text-ink-muted">
            Lý do, đánh đổi và hướng đã loại bỏ sau mỗi quyết định agent đưa ra.
          </p>
        </motion.div>

        {/* Signature visual: knowledge density heatmap */}
        {hasContexts && (
          <div className="mb-6">
            <KnowledgeHeatmap />
          </div>
        )}

        {/* Filter */}
        <div className="mb-6">
          <FilterBar active={domain} onChange={setDomain} />
        </div>

        {/* Feed */}
        {isLoading ? (
          <div className="flex flex-col gap-3.5">
            <CardSkeleton />
            <CardSkeleton />
            <CardSkeleton />
          </div>
        ) : hasContexts ? (
          <div className="flex flex-col gap-3.5 pb-6">
            {contexts!.map((ctx, i) => (
              <ContextCard key={ctx.id} ctx={ctx} index={i} featured={i === 0} />
            ))}
          </div>
        ) : (
          <SetupChecklist status={status} />
        )}
      </main>

      <RightPanel />
    </div>
  );
}
