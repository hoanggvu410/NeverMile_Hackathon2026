"use client";

import dynamic from "next/dynamic";
import { useGraph } from "@/hooks/useGitWhy";
import { EmptyState } from "@/components/ui/EmptyState";
import { IconGraph } from "@/components/ui/icons";

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
  const { nodes, edges } = useGraph();
  const loading = nodes.isLoading || edges.isLoading;
  const data = nodes.data ?? [];

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center justify-between px-6 py-4 lg:px-8">
        <div>
          <h1 className="text-[22px] font-light text-ink">Đồ thị claim</h1>
          <p className="mt-0.5 text-[12px] text-ink-muted">
            {data.length} claims · {edges.data?.length ?? 0} cạnh
          </p>
        </div>
      </div>
      <div className="min-h-0 flex-1">
        {loading ? (
          <GraphLoading />
        ) : data.length === 0 ? (
          <EmptyState
            icon={<IconGraph className="h-6 w-6" />}
            title="Chưa có claim nào"
            description="Lưu ngữ cảnh từ agent để dựng claim graph."
          />
        ) : (
          <ClaimGraph nodes={data} edges={edges.data ?? []} />
        )}
      </div>
    </div>
  );
}
