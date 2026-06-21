"use client";

import Link from "next/link";
import { motion } from "framer-motion";
import type { SearchResult } from "@/types";
import { Badge } from "@/components/ui/Badge";
import { domainColor, edgeColor, scoreColor, truncate } from "@/lib/utils";
import { IconArrowRight } from "@/components/ui/icons";

export function SearchResultCard({
  result,
  index,
}: {
  result: SearchResult;
  index: number;
}) {
  const color = domainColor(result.domain);
  const pct = Math.round((result.score ?? 0) * 100);
  const text = result.claim || result.prompt || result.title;

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.28, delay: index * 0.04, ease: [0.23, 1, 0.32, 1] }}
    >
      <Link href={`/dashboard/contexts/${result.id}`} className="group block">
        <div className="glass rounded-[14px] p-5 transition duration-150 hover:-translate-y-0.5 hover:shadow-card">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 flex-1">
              <div className="mb-2 flex flex-wrap items-center gap-1.5">
                <Badge color={color} dot>
                  {result.domain}
                </Badge>
                <span className="text-[11px] text-ink-muted">/ {result.topic}</span>
                {result.edge_type && (
                  <Badge color={edgeColor(result.edge_type)}>
                    {result.edge_type}
                  </Badge>
                )}
                {result.vector_kind && (
                  <span className="text-[10.5px] uppercase tracking-wider text-ink-muted/70">
                    {result.vector_kind}
                  </span>
                )}
              </div>
              <p className="text-[14px] leading-relaxed text-ink/90">
                {truncate(text, 200)}
              </p>
              {result.title && result.claim && (
                <p className="mt-1.5 text-[12px] text-ink-muted">
                  từ “{truncate(result.title, 70)}”
                </p>
              )}
            </div>

            {result.score > 0 && (
              <div className="flex shrink-0 flex-col items-center">
                <div
                  className="flex h-12 w-12 items-center justify-center rounded-full text-[13px] font-light"
                  style={{
                    color: scoreColor(result.score),
                    background: `${scoreColor(result.score)}1a`,
                    border: `1px solid ${scoreColor(result.score)}40`,
                  }}
                >
                  {pct}
                </div>
                <span className="mt-1 text-[10px] text-ink-muted">khớp</span>
              </div>
            )}
          </div>
          <div className="mt-3 flex items-center justify-end text-[12px] text-ink-muted opacity-0 transition-opacity group-hover:opacity-100">
            <span className="flex items-center gap-1 text-accent">
              Mở ngữ cảnh <IconArrowRight className="h-3.5 w-3.5" />
            </span>
          </div>
        </div>
      </Link>
    </motion.div>
  );
}
