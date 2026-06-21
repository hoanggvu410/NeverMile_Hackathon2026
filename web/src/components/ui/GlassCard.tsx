"use client";

import { motion } from "framer-motion";
import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

interface GlassCardProps {
  children: ReactNode;
  className?: string;
  hover?: boolean;
  glow?: boolean;
  delay?: number;
  onClick?: () => void;
}

export function GlassCard({
  children,
  className,
  hover = false,
  glow = false,
  delay = 0,
  onClick,
}: GlassCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.25, ease: [0.23, 1, 0.32, 1], delay }}
      whileHover={hover ? { y: -2 } : undefined}
      onClick={onClick}
      className={cn(
        "glass rounded-[14px] p-5",
        glow && "shadow-glow border-accent/40",
        hover && "cursor-pointer transition-shadow duration-150 hover:shadow-card",
        className
      )}
    >
      {children}
    </motion.div>
  );
}
