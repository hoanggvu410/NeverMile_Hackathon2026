"use client";

import { motion } from "framer-motion";
import { IconCheck } from "@/components/ui/icons";
import type { Status } from "@/types";

interface Step {
  label: string;
  hint: string;
  done: boolean;
}

export function SetupChecklist({ status }: { status?: Status }) {
  const ctxCount = status?.context_count ?? 0;
  const steps: Step[] = [
    {
      label: "Đã cài MCP server",
      hint: "gitwhy2-mcp",
      done: true,
    },
    {
      label: "Đã cài post-commit hook",
      hint: "git why hook install",
      done: (status?.pending_commits.length ?? 0) >= 0 && ctxCount > 0,
    },
    {
      label: "Đã lưu ngữ cảnh đầu tiên",
      hint: "gitwhy_save",
      done: ctxCount > 0,
    },
    {
      label: "Chạy tìm kiếm",
      hint: 'git why search "..."',
      done: ctxCount > 1,
    },
  ];

  return (
    <div className="glass-strong rounded-[14px] p-6">
      <h3 className="text-[16px] font-light text-ink">Bắt đầu với GitWhy</h3>
      <p className="mt-1 text-[13px] text-ink-muted">
        Vài bước để bắt đầu ghi lại lý do đằng sau code của bạn.
      </p>
      <div className="mt-5 flex flex-col gap-2.5">
        {steps.map((step, i) => (
          <motion.div
            key={step.label}
            initial={{ opacity: 0, x: -6 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: i * 0.06, duration: 0.25 }}
            className="glass-inner flex items-center gap-3 rounded-[10px] px-4 py-3"
          >
            <span
              className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full ${
                step.done
                  ? "bg-success/20 text-success"
                  : "border border-border text-ink-muted"
              }`}
            >
              {step.done ? (
                <IconCheck className="h-3.5 w-3.5" />
              ) : (
                <span className="text-[11px]">{i + 1}</span>
              )}
            </span>
            <div className="min-w-0 flex-1">
              <div className="text-[13px] font-light text-ink">{step.label}</div>
            </div>
            <code className="rounded bg-white/[0.06] px-2 py-1 font-mono text-[11px] text-accent">
              {step.hint}
            </code>
          </motion.div>
        ))}
      </div>
    </div>
  );
}
