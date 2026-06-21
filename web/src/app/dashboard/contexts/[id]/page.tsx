"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { motion } from "framer-motion";
import { useContext } from "@/hooks/useGitWhy";
import { Prose } from "@/components/contexts/Prose";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Skeleton } from "@/components/ui/Skeleton";
import {
  IconArrowRight,
  IconCheck,
  IconCommit,
  IconCopy,
  IconFile,
} from "@/components/ui/icons";
import { domainColor, formatDate, shortSha } from "@/lib/utils";

const SECTIONS: { key: string; label: string }[] = [
  { key: "what_was_done", label: "Đã làm gì" },
  { key: "reasoning", label: "Lý do" },
  { key: "key_decisions", label: "Quyết định chính" },
  { key: "rejected_alternatives", label: "Phương án đã loại" },
  { key: "risks_and_open_questions", label: "Rủi ro & câu hỏi mở" },
  { key: "verification", label: "Kiểm chứng" },
];

export default function ContextDetailPage() {
  const params = useParams();
  const router = useRouter();
  const id = String(params.id);
  const { data: ctx, isLoading } = useContext(id);
  const [copied, setCopied] = useState(false);
  const [showRaw, setShowRaw] = useState(false);

  const copyId = () => {
    navigator.clipboard.writeText(id);
    setCopied(true);
    setTimeout(() => setCopied(false), 1600);
  };

  if (isLoading) {
    return (
      <main className="h-full overflow-y-auto px-8 py-7">
        <Skeleton className="mb-4 h-4 w-32" />
        <Skeleton className="mb-6 h-8 w-2/3" />
        <Skeleton className="mb-3 h-24 w-full" />
        <Skeleton className="h-24 w-full" />
      </main>
    );
  }

  if (!ctx) {
    return (
      <main className="flex h-full items-center justify-center px-8">
        <div className="text-center">
          <p className="text-ink-muted">Không tìm thấy ngữ cảnh.</p>
          <Button className="mt-4" onClick={() => router.push("/dashboard/contexts")}>
            Về danh sách ngữ cảnh
          </Button>
        </div>
      </main>
    );
  }

  const color = domainColor(ctx.domain);

  return (
    <main className="h-full overflow-y-auto px-6 py-6 lg:px-10">
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3, ease: [0.23, 1, 0.32, 1] }}
        className="mx-auto max-w-[860px]"
      >
        <button
          onClick={() => router.back()}
          className="mb-4 flex items-center gap-1.5 text-[12px] text-ink-muted transition-colors hover:text-ink"
        >
          <IconArrowRight className="h-3.5 w-3.5 rotate-180" /> Quay lại
        </button>

        {/* Header */}
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="min-w-0 flex-1">
            <div className="mb-2.5 flex flex-wrap items-center gap-2">
              <Badge color={color} dot>
                {ctx.domain}
              </Badge>
              <span className="text-[12px] text-ink-muted">/ {ctx.topic}</span>
              <span className="text-[12px] text-ink-muted">·</span>
              <span className="text-[12px] text-ink-muted">
                {formatDate(ctx.date)}
              </span>
              {ctx.branch && (
                <Badge>
                  <span className="font-mono text-[10.5px]">{ctx.branch}</span>
                </Badge>
              )}
            </div>
            <h1 className="text-[24px] font-light leading-snug text-ink">
              {ctx.title}
            </h1>
            <div className="mt-2 flex items-center gap-3 text-[12px] text-ink-muted">
              {ctx.saved_by && <span>lưu bởi {ctx.saved_by}</span>}
              {ctx.agent && (
                <span className="flex items-center gap-1.5">
                  <span className="h-1 w-1 rounded-full bg-accent" />
                  {ctx.agent}
                </span>
              )}
            </div>
          </div>

          <div className="flex shrink-0 items-center gap-2">
            <Button variant="glass" onClick={copyId}>
              {copied ? (
                <IconCheck className="h-4 w-4 text-success" />
              ) : (
                <IconCopy className="h-4 w-4" />
              )}
              {copied ? "Đã chép" : "Chép ID"}
            </Button>
            <Button
              variant={showRaw ? "accent" : "glass"}
              onClick={() => setShowRaw((v) => !v)}
            >
              Raw
            </Button>
          </div>
        </div>

        {/* Prompt callout */}
        <div className="glass-inner mt-6 rounded-[12px] border-l-2 border-l-accent px-5 py-4">
          <div className="mb-1 text-[11px] uppercase tracking-wider text-accent/80">
            Yêu cầu
          </div>
          <p className="text-[13.5px] leading-relaxed text-ink/85">
            {ctx.prompt}
          </p>
        </div>

        {showRaw ? (
          <pre className="glass-inner mt-6 overflow-x-auto rounded-[12px] p-5 font-mono text-[12px] leading-relaxed text-ink/80">
            {rawMarkdown(ctx)}
          </pre>
        ) : (
          <div className="mt-6 flex flex-col gap-4">
            {SECTIONS.map((s) => {
              const value = (ctx as unknown as Record<string, string>)[s.key];
              if (!value || !value.trim()) return null;
              return (
                <section
                  key={s.key}
                  className="glass rounded-[14px] p-5"
                >
                  <h2
                    className="mb-3 text-[13px] font-medium uppercase tracking-wider"
                    style={{ color }}
                  >
                    {s.label}
                  </h2>
                  <Prose text={value} />
                </section>
              );
            })}

            {/* Files */}
            {ctx.files.length > 0 && (
              <section className="glass rounded-[14px] p-5">
                <h2 className="mb-3 flex items-center gap-2 text-[13px] font-medium uppercase tracking-wider text-ink-muted">
                  <IconFile className="h-4 w-4" /> Tệp ({ctx.files.length})
                </h2>
                <div className="overflow-hidden rounded-[10px] border border-border">
                  {ctx.files.map((f, i) => (
                    <div
                      key={i}
                      className="flex items-center gap-3 border-b border-border/50 px-4 py-2.5 last:border-0"
                    >
                      <code className="flex-1 truncate font-mono text-[12px] text-ink/85">
                        {f.file}
                      </code>
                      <Badge
                        color={
                          f.status === "added"
                            ? "var(--success)"
                            : f.status === "deleted"
                            ? "var(--danger)"
                            : "var(--accent-2)"
                        }
                      >
                        {f.status || "modified"}
                      </Badge>
                    </div>
                  ))}
                </div>
              </section>
            )}

            {/* Commits */}
            {ctx.commits.length > 0 && (
              <section className="glass rounded-[14px] p-5">
                <h2 className="mb-3 flex items-center gap-2 text-[13px] font-medium uppercase tracking-wider text-ink-muted">
                  <IconCommit className="h-4 w-4" /> Commits
                </h2>
                <div className="flex flex-wrap gap-2">
                  {ctx.commits.map((c) => (
                    <code
                      key={c}
                      className="glass-inner rounded-[8px] px-2.5 py-1 font-mono text-[12px] text-accent"
                    >
                      {shortSha(c)}
                    </code>
                  ))}
                </div>
              </section>
            )}
          </div>
        )}
      </motion.div>
    </main>
  );
}

function rawMarkdown(ctx: {
  title: string;
  id: string;
  domain: string;
  topic: string;
  prompt: string;
  what_was_done: string;
  reasoning: string;
  key_decisions: string;
  rejected_alternatives: string;
  risks_and_open_questions: string;
  verification: string;
}): string {
  return [
    `# Context: ${ctx.title}`,
    ``,
    `**Context ID:** ${ctx.id}`,
    `**Domain:** ${ctx.domain}`,
    `**Topic:** ${ctx.topic}`,
    ``,
    `## Prompt`,
    ``,
    ctx.prompt,
    ``,
    `## What Was Done`,
    ``,
    ctx.what_was_done,
    ``,
    `## Reasoning`,
    ``,
    ctx.reasoning,
    ``,
    `## Key Decisions`,
    ``,
    ctx.key_decisions,
    ``,
    `## Rejected Alternatives`,
    ``,
    ctx.rejected_alternatives,
    ``,
    `## Risks & Open Questions`,
    ``,
    ctx.risks_and_open_questions,
    ``,
    `## Verification`,
    ``,
    ctx.verification,
  ].join("\n");
}
