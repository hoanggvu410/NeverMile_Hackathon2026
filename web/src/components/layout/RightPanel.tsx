"use client";

import Link from "next/link";
import { motion } from "framer-motion";
import { useContexts, useGraph, useStatus } from "@/hooks/useGitWhy";
import { Badge } from "@/components/ui/Badge";
import { domainColor, timeAgo, truncate } from "@/lib/utils";
import { IconArrowRight, IconSpark } from "@/components/ui/icons";

function Stat({
  label,
  value,
  color,
  filled,
}: {
  label: string;
  value: number | string;
  color: string;
  filled?: boolean;
}) {
  return (
    <div
      className="relative overflow-hidden rounded-[12px] px-3 py-3"
      style={
        filled
          ? {
              background: `linear-gradient(150deg, ${color}38, ${color}0d 70%)`,
              border: `1px solid ${color}40`,
            }
          : {
              background: "oklch(0.22 0.02 265 / 0.55)",
              border: "1px solid var(--border)",
            }
      }
    >
      <div
        className="text-[22px] font-light leading-tight"
        style={{ color: filled ? "#fff" : "var(--ink)" }}
      >
        {value}
      </div>
      <div className="mt-0.5 flex items-center gap-1.5 text-[11px] tracking-wide text-ink-muted">
        <span className="h-1.5 w-1.5 rounded-full" style={{ background: color }} />
        {label}
      </div>
    </div>
  );
}

export function RightPanel() {
  const { data: status } = useStatus();
  const { data: contexts } = useContexts();
  const { nodes, edges } = useGraph();

  const latest = contexts?.[0];
  const claimCount = nodes.data?.length ?? 0;
  const edgeCount = edges.data?.length ?? 0;
  const ctxCount = status?.context_count ?? contexts?.length ?? 0;

  const health =
    ctxCount > 0
      ? Math.min(100, Math.round((claimCount / (ctxCount * 5)) * 100))
      : 0;

  return (
    <aside className="hidden w-[300px] shrink-0 flex-col gap-4 overflow-y-auto px-4 py-5 xl:flex">
      {/* Latest context */}
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.25, ease: [0.23, 1, 0.32, 1] }}
        className="glass-strong rounded-[14px] p-4"
      >
        <div className="mb-3 flex items-center justify-between">
          <span className="text-[13px] font-light text-ink">Ngữ cảnh mới nhất</span>
          <span className="flex h-7 w-7 items-center justify-center rounded-full bg-accent/20 text-accent">
            <IconSpark className="h-4 w-4" />
          </span>
        </div>
        {latest ? (
          <Link href={`/dashboard/contexts/${latest.id}`} className="group block">
            <div className="mb-2 flex flex-wrap gap-1.5">
              <Badge color={domainColor(latest.domain)} dot>
                {latest.domain}
              </Badge>
            </div>
            <p className="text-[13px] leading-relaxed text-ink/90">
              {truncate(latest.title || latest.prompt, 90)}
            </p>
            <div className="mt-3 flex items-center justify-between text-[11px] text-ink-muted">
              <span>{timeAgo(latest.date)}</span>
              <span className="flex items-center gap-1 text-accent opacity-0 transition-opacity group-hover:opacity-100">
                Xem <IconArrowRight className="h-3 w-3" />
              </span>
            </div>
          </Link>
        ) : (
          <p className="text-[12px] text-ink-muted">Chưa có ngữ cảnh nào.</p>
        )}
      </motion.div>

      {/* Stats */}
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.25, delay: 0.05, ease: [0.23, 1, 0.32, 1] }}
        className="glass-strong rounded-[14px] p-4"
      >
        <div className="mb-3 text-[13px] font-light text-ink">Tổng quan</div>
        <div className="grid grid-cols-2 gap-2.5">
          <Stat label="Claims" value={claimCount} color="#8b6bff" filled />
          <Stat label="Cạnh đồ thị" value={edgeCount} color="#3a8dff" filled />
          <Stat label="Ngữ cảnh" value={ctxCount} color="#22c1a6" />
          <Stat label="Đang chờ" value={status?.pending_commits.length ?? 0} color="#e0a23a" />
        </div>
      </motion.div>

      {/* Graph health */}
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.25, delay: 0.1, ease: [0.23, 1, 0.32, 1] }}
        className="glass-strong rounded-[14px] p-4"
      >
        <div className="mb-1 flex items-center justify-between">
          <span className="text-[13px] font-light text-ink">Sức khỏe đồ thị</span>
          <span className="text-[13px] font-light text-accent">{health}%</span>
        </div>
        <p className="mb-3 text-[11px] text-ink-muted">
          Độ phủ claims so với kỳ vọng
        </p>
        <div className="h-2 overflow-hidden rounded-full bg-white/[0.06]">
          <motion.div
            className="h-full rounded-full bg-gradient-to-r from-accent-2 to-accent"
            initial={{ width: 0 }}
            animate={{ width: `${health}%` }}
            transition={{ duration: 0.6, ease: [0.23, 1, 0.32, 1] }}
          />
        </div>
        <div className="mt-3 flex items-center gap-2 text-[11px] text-ink-muted">
          <span
            className={`h-1.5 w-1.5 rounded-full ${
              status?.graph_ready ? "bg-success" : "bg-warning"
            }`}
          />
          {status?.graph_ready ? "Graph engine đang chạy" : "Graph offline"}
        </div>
      </motion.div>

      {/* Brand card (Sapphire-style) */}
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.25, delay: 0.15, ease: [0.23, 1, 0.32, 1] }}
        className="relative overflow-hidden rounded-[16px] p-5"
        style={{
          background:
            "linear-gradient(150deg, oklch(0.4 0.2 295 / 0.85), oklch(0.22 0.1 280 / 0.75) 60%, oklch(0.16 0.04 265 / 0.7))",
          border: "1px solid oklch(0.5 0.18 290 / 0.5)",
        }}
      >
        <div
          className="pointer-events-none absolute -right-8 -top-8 h-32 w-32 rounded-full blur-2xl"
          style={{ background: "oklch(0.65 0.24 295 / 0.45)" }}
        />
        <div className="relative flex items-center gap-3">
          <span className="flex h-11 w-11 items-center justify-center rounded-[12px] bg-white/10 backdrop-blur">
            <IconSpark className="h-6 w-6 text-white" />
          </span>
          <div>
            <div className="text-[15px] font-light leading-tight text-white">GitWhy</div>
            <div className="text-[11.5px] text-white/65">Bộ nhớ quyết định</div>
          </div>
        </div>
        <p className="relative mt-3 text-[11.5px] leading-relaxed text-white/70">
          Mọi claim, cạnh và đánh đổi agent ghi lại, sẵn sàng truy xuất trước
          quyết định kế tiếp.
        </p>
      </motion.div>
    </aside>
  );
}
