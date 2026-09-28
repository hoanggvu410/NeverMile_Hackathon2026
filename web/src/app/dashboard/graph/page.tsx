"use client";

import { useState } from "react";
import dynamic from "next/dynamic";
import { useCommits, useGraph } from "@/hooks/useGitWhy";
import { EmptyState } from "@/components/ui/EmptyState";
import { IconGraph } from "@/components/ui/icons";
import { CommitGraph } from "@/components/graph/CommitGraph";
import { useLanguage } from "@/lib/language";

const ClaimGraph = dynamic(
  () => import("@/components/graph/ClaimGraph").then((m) => m.ClaimGraph),
  { ssr: false, loading: () => <GraphLoading /> }
);

function GraphLoading() {
  return (
    <div className="flex h-full items-center justify-center">
      <span className="h-6 w-6 animate-spin rounded-full border-2 border-accent border-t-transparent" />
    </div>
  );
}

export default function GraphPage() {
  const [view, setView] = useState<"commits" | "claims">("commits");
  const { t } = useLanguage();
  const { nodes, edges } = useGraph();
  const commits = useCommits();
  const loading = view === "commits" ? commits.isLoading : nodes.isLoading || edges.isLoading;
  const data = nodes.data ?? [];
  const commitData = commits.data ?? [];

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center justify-between px-6 py-4 lg:px-8">
        <div>
          <h1 className="text-[22px] font-light text-ink">
            {view === "commits" ? "Git commit graph" : t("Đồ thị claim")}
          </h1>
          <p className="mt-0.5 text-[12px] text-ink-muted">
            {view === "commits"
              ? `${commitData.length} commits · ${commitData.reduce((total, commit) => total + commit.parents.filter((parent) => commitData.some((item) => item.sha === parent)).length, 0)} parent links`
              : `${data.length} claims · ${edges.data?.length ?? 0} ${t("cạnh")}`}
          </p>
        </div>
        <div className="flex rounded-[10px] border border-border/70 bg-white/[0.03] p-1">
          <button
            onClick={() => setView("commits")}
            className={`rounded-lg px-3 py-1.5 text-[11px] transition ${view === "commits" ? "bg-accent/20 text-accent" : "text-ink-muted hover:text-ink"}`}
          >
            Commits
          </button>
          <button
            onClick={() => setView("claims")}
            className={`rounded-lg px-3 py-1.5 text-[11px] transition ${view === "claims" ? "bg-accent/20 text-accent" : "text-ink-muted hover:text-ink"}`}
          >
            Claims
          </button>
        </div>
      </div>
      <div className="min-h-0 flex-1">
        {loading ? (
          <GraphLoading />
        ) : view === "commits" && commitData.length > 0 ? (
          <CommitGraph commits={commitData} />
        ) : view === "claims" && data.length > 0 ? (
          <ClaimGraph nodes={data} edges={edges.data ?? []} />
        ) : view === "commits" ? (
          <EmptyState
            icon={<IconGraph className="h-6 w-6" />}
            title={t("Không tìm thấy commit")}
            description={t(commits.isError ? "Không thể đọc lịch sử Git của project đang kết nối." : "Project này chưa có commit history.")}
          />
        ) : (
          <EmptyState
            icon={<IconGraph className="h-6 w-6" />}
            title={t("Chưa có claim nào")}
            description={t("Lưu ngữ cảnh từ agent để dựng claim graph.")}
          />
        )}
      </div>
    </div>
  );
}
