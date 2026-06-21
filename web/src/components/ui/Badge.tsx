import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

interface BadgeProps {
  children: ReactNode;
  color?: string;
  className?: string;
  dot?: boolean;
}

export function Badge({ children, color, className, dot }: BadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-[11px] font-light tracking-wide",
        className
      )}
      style={
        color
          ? {
              color,
              borderColor: `${color}55`,
              background: `${color}18`,
            }
          : {
              color: "var(--ink-muted)",
              borderColor: "var(--border)",
              background: "oklch(0.22 0.02 265 / 0.5)",
            }
      }
    >
      {dot && (
        <span
          className="h-1.5 w-1.5 rounded-full"
          style={{ background: color ?? "var(--ink-muted)" }}
        />
      )}
      {children}
    </span>
  );
}
