"use client";

import Link from "next/link";
import { motion } from "framer-motion";
import type { ContextSummary } from "@/types";
import { Badge } from "@/components/ui/Badge";
import { domainColor, timeAgo, truncate } from "@/lib/utils";
import {
  IconArrowRight,
  IconClock,
  IconContexts,
} from "@/components/ui/icons";

interface ContextCardProps {
  ctx: ContextSummary;
  index: number;
  featured?: boolean;
}

export function ContextCard({ ctx, index, featured }: ContextCardProps) {
  const color = domainColor(ctx.domain);
  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{
        duration: 0.3,
        delay: index * 0.05,
        ease: [0.23, 1, 0.32, 1],
      }}
    >
      <Link href={`/dashboard/contexts/${ctx.id}`} className="group block">
        <div
          className={`glass relative overflow-hidden rounded-[14px] p-5 transition duration-150 ease-out-soft hover:-translate-y-0.5 hover:shadow-card ${
            featured ? "shadow-glow" : ""
          }`}
          style={
            featured
              ? { borderColor: "var(--accent-glow)" }
              : undefined
          }
        >
          {/* Featured glow is locked to the brand accent (not the per-domain
              hash color) so the hero card stays on-palette across domains. */}
          {featured && (
            <>
              <div
                className="pointer-events-none absolute inset-0 opacity-[0.16]"
                style={{
                  background:
                    "radial-gradient(520px 220px at 12% -10%, var(--accent), transparent 68%)",
                }}
              />
              <div
                className="pointer-events-none absolute -inset-px rounded-[14px]"
                style={{
                  boxShadow:
                    "inset 0 0 0 1px var(--accent-glow), 0 0 50px -12px var(--accent-glow)",
                }}
              />
            </>
          )}

          <div className="relative flex items-start gap-4">
            <div
              className="mt-0.5 flex h-10 w-10 shrink-0 items-center justify-center rounded-[10px]"
              style={{ background: `${color}22`, color }}
            >
              <IconContexts className="h-5 w-5" />
            </div>

            <div className="min-w-0 flex-1">
              <div className="mb-1.5 flex flex-wrap items-center gap-1.5">
                <Badge color={color} dot>
                  {ctx.domain}
                </Badge>
                <span className="text-[11px] text-ink-muted">/ {ctx.topic}</span>
              </div>

              <h3 className="text-[15px] font-light leading-snug text-ink">
                {truncate(ctx.title || ctx.prompt, 90)}
              </h3>
              <p className="mt-1 text-[12.5px] leading-relaxed text-ink-muted">
                {truncate(ctx.prompt, 130)}
              </p>

              <div className="mt-3 flex items-center gap-4 text-[11px] text-ink-muted">
                <span className="flex items-center gap-1.5">
                  <IconClock className="h-3.5 w-3.5" />
                  {timeAgo(ctx.date)}
                </span>
                <span className="font-mono text-[10.5px] opacity-70">
                  {ctx.id}
                </span>
              </div>
            </div>

            <span className="flex items-center gap-1 self-center text-[12px] text-ink-muted opacity-0 transition duration-150 group-hover:text-accent group-hover:opacity-100">
              Xem <IconArrowRight className="h-3.5 w-3.5" />
            </span>
          </div>
        </div>
      </Link>
    </motion.div>
  );
}
