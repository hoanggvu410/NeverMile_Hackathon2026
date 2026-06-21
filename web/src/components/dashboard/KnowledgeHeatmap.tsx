"use client";

import Link from "next/link";
import { motion } from "framer-motion";
import { useGraph } from "@/hooks/useGitWhy";
import { Skeleton } from "@/components/ui/Skeleton";
import { IconArrowRight, IconGraph } from "@/components/ui/icons";
import { truncate } from "@/lib/utils";
import type { GraphNode } from "@/types";

const COLS = 12;
const HEAT = ["var(--heat-1)", "var(--heat-2)", "var(--heat-3)", "var(--heat-4)"];

// importance 1-5 -> heat band 0-3 (Low / Medium / High / Best)
function band(importance: number): number {
  if (importance <= 2) return 0;
  if (importance === 3) return 1;
  if (importance === 4) return 2;
  return 3;
}

const LEGEND = [
  { label: "Thấp", color: "var(--heat-1)" },
  { label: "Vừa", color: "var(--heat-2)" },
  { label: "Cao", color: "var(--heat-3)" },
  { label: "Tối đa", color: "var(--heat-4)" },
];

interface Row {
  domain: string;
  claims: GraphNode[];
}

export function KnowledgeHeatmap() {
  const { nodes } = useGraph();

  if (nodes.isLoading) {
    return <Skeleton className="h-[260px] w-full rounded-[16px]" />;
  }

  const data = nodes.data ?? [];
  if (data.length === 0) return null;

  const byDomain = new Map<string, GraphNode[]>();
  for (const n of data) {
    const arr = byDomain.get(n.domain) ?? [];
    arr.push(n);
    byDomain.set(n.domain, arr);
  }
  const rows: Row[] = [...byDomain.entries()]
    .map(([domain, claims]) => ({
      domain,
      claims: [...claims].sort((a, b) => b.importance - a.importance),
    }))
    .sort((a, b) => b.claims.length - a.claims.length);

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35, ease: [0.23, 1, 0.32, 1] }}
      className="glass-strong relative overflow-hidden rounded-[16px] p-5"
    >
      <div
        className="pointer-events-none absolute inset-0 opacity-[0.5]"
        style={{
          background:
            "radial-gradient(620px 200px at 85% -20%, oklch(0.6 0.24 292 / 0.25), transparent 70%)",
        }}
      />
      <div className="relative">
        <div className="mb-4 flex items-start justify-between">
          <div>
            <h2 className="flex items-center gap-2 text-[15px] font-light text-ink">
              <span className="flex h-7 w-7 items-center justify-center rounded-[8px] bg-accent/20 text-accent">
                <IconGraph className="h-4 w-4" />
              </span>
              Mật độ tri thức
            </h2>
            <p className="mt-1 text-[12px] text-ink-muted">
              Claims theo domain, đậm dần theo độ quan trọng
            </p>
          </div>
          <div className="flex items-center gap-3">
            <div className="hidden items-center gap-2.5 sm:flex">
              {LEGEND.map((l) => (
                <span key={l.label} className="flex items-center gap-1.5 text-[11px] text-ink-muted">
                  <span className="h-2.5 w-2.5 rounded-[3px]" style={{ background: l.color }} />
                  {l.label}
                </span>
              ))}
            </div>
            <Link
              href="/dashboard/graph"
              className="group flex items-center gap-1 text-[12px] text-accent"
            >
              Mở đồ thị
              <IconArrowRight className="h-3.5 w-3.5 transition-transform group-hover:translate-x-0.5" />
            </Link>
          </div>
        </div>

        <div className="flex flex-col gap-1.5">
          {rows.map((row, ri) => (
            <div key={row.domain} className="flex items-center gap-3">
              <span className="w-28 shrink-0 truncate text-right font-mono text-[10.5px] uppercase tracking-wider text-ink-muted">
                {row.domain.split("/").pop()}
              </span>
              <div
                className="grid flex-1 gap-1.5"
                style={{ gridTemplateColumns: `repeat(${COLS}, minmax(0, 1fr))` }}
              >
                {Array.from({ length: COLS }).map((_, ci) => {
                  const claim = row.claims[ci];
                  if (!claim) {
                    return <div key={ci} className="heat-empty aspect-square rounded-[5px]" />;
                  }
                  const b = band(claim.importance);
                  return (
                    <motion.div
                      key={claim.id}
                      initial={{ opacity: 0, scale: 0.9 }}
                      animate={{ opacity: 1, scale: 1 }}
                      transition={{
                        delay: ri * 0.05 + ci * 0.02,
                        duration: 0.22,
                        ease: [0.23, 1, 0.32, 1],
                      }}
                      title={`${claim.claim_type} · importance ${claim.importance}\n${truncate(
                        claim.claim,
                        100
                      )}`}
                      className="aspect-square cursor-default rounded-[5px] transition-transform duration-150 hover:scale-110"
                      style={{
                        background: HEAT[b],
                        boxShadow: b >= 2 ? `0 0 12px -2px ${HEAT[b]}` : undefined,
                      }}
                    />
                  );
                })}
              </div>
            </div>
          ))}
        </div>
      </div>
    </motion.div>
  );
}
