"use client";

import { useState } from "react";
import { motion } from "framer-motion";
import { useDomains } from "@/hooks/useGitWhy";
import { cn } from "@/lib/utils";
import { IconGit, IconPlus } from "@/components/ui/icons";

interface FilterBarProps {
  active: string;
  onChange: (domain: string) => void;
}

export function FilterBar({ active, onChange }: FilterBarProps) {
  const { data: domains } = useDomains();
  const [showSave, setShowSave] = useState(false);
  const chips = ["all", ...(domains ?? [])];

  return (
    <div className="flex flex-wrap items-center gap-2">
      {chips.map((chip) => {
        const isActive = active === chip || (chip === "all" && active === "");
        return (
          <button
            key={chip}
            onClick={() => onChange(chip === "all" ? "" : chip)}
            className={cn(
              "rounded-full border px-3.5 py-1.5 text-[12px] font-light tracking-wide transition duration-150 ease-out-soft active:scale-[0.97]",
              isActive
                ? "border-transparent bg-accent text-white shadow-[0_6px_20px_-8px_var(--accent-glow)]"
                : "glass-inner text-ink-muted hover:text-ink"
            )}
          >
            {chip === "all" ? "Tất cả" : chip}
          </button>
        );
      })}

      <div className="relative ml-auto">
        <button
          onMouseEnter={() => setShowSave(true)}
          onMouseLeave={() => setShowSave(false)}
          className="flex items-center gap-2 rounded-full bg-accent px-4 py-1.5 text-[12px] font-light text-white shadow-[0_8px_24px_-8px_var(--accent-glow)] transition duration-150 hover:-translate-y-px hover:brightness-110 active:scale-[0.97]"
        >
          <IconPlus className="h-4 w-4" />
          Lưu Context
        </button>
        {showSave && (
          <motion.div
            initial={{ opacity: 0, y: -4 }}
            animate={{ opacity: 1, y: 0 }}
            className="glass-strong absolute right-0 top-[calc(100%+8px)] z-20 w-64 rounded-[10px] p-3 text-[12px] text-ink-muted"
          >
            <div className="mb-1.5 flex items-center gap-2 text-ink">
              <IconGit className="h-4 w-4 text-accent" /> Lưu từ agent
            </div>
            Cho agent gọi{" "}
            <code className="rounded bg-white/10 px-1 py-0.5 font-mono text-[11px] text-accent">
              gitwhy_save
            </code>{" "}
            hoặc chạy{" "}
            <code className="rounded bg-white/10 px-1 py-0.5 font-mono text-[11px] text-accent">
              git why save
            </code>
            .
          </motion.div>
        )}
      </div>
    </div>
  );
}
