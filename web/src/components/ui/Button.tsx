"use client";

import { cn } from "@/lib/utils";
import type { ButtonHTMLAttributes, ReactNode } from "react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "accent" | "glass" | "ghost";
  children: ReactNode;
}

export function Button({
  variant = "glass",
  className,
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      {...props}
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-[10px] px-4 py-2 text-[13px] font-light tracking-wide transition duration-150 ease-out-soft active:scale-[0.97] disabled:opacity-50",
        variant === "accent" &&
          "bg-accent text-white shadow-[0_8px_24px_-8px_var(--accent-glow)] hover:brightness-110 hover:-translate-y-px",
        variant === "glass" &&
          "glass-inner text-ink hover:border-accent/50 hover:text-white",
        variant === "ghost" &&
          "text-ink-muted hover:bg-white/5 hover:text-ink",
        className
      )}
    >
      {children}
    </button>
  );
}
