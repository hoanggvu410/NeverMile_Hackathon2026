"use client";

import { motion, useReducedMotion } from "framer-motion";

/**
 * Vivid purple/blue aurora bloom behind the app, inspired by the reference
 * "nebula over the horizon" look. Pure CSS blurred radial blobs, fixed and
 * non-interactive. Sits above the body gradient, below the starfield content.
 */
export function Aurora() {
  const reduce = useReducedMotion();
  return (
    <div
      aria-hidden
      className="pointer-events-none fixed inset-0 z-0 overflow-hidden"
    >
      {/* main bloom, top-center */}
      <motion.div
        className="absolute -top-[18vh] left-1/2 h-[60vh] w-[80vw] -translate-x-1/2 rounded-full blur-[120px]"
        style={{
          background:
            "radial-gradient(closest-side, oklch(0.6 0.26 295 / 0.55), oklch(0.5 0.2 280 / 0.22), transparent 72%)",
        }}
        animate={reduce ? undefined : { opacity: [0.75, 1, 0.75], scale: [1, 1.06, 1] }}
        transition={{ duration: 14, repeat: Infinity, ease: "easeInOut" }}
      />
      {/* cool blue companion, upper right */}
      <motion.div
        className="absolute -top-[8vh] right-[6vw] h-[42vh] w-[40vw] rounded-full blur-[110px]"
        style={{
          background:
            "radial-gradient(closest-side, oklch(0.62 0.18 250 / 0.4), transparent 70%)",
        }}
        animate={reduce ? undefined : { opacity: [0.5, 0.85, 0.5] }}
        transition={{ duration: 18, repeat: Infinity, ease: "easeInOut", delay: 2 }}
      />
      {/* warm magenta low-left for color depth */}
      <motion.div
        className="absolute bottom-[2vh] left-[2vw] h-[40vh] w-[42vw] rounded-full blur-[120px]"
        style={{
          background:
            "radial-gradient(closest-side, oklch(0.55 0.2 330 / 0.28), transparent 72%)",
        }}
        animate={reduce ? undefined : { opacity: [0.4, 0.7, 0.4] }}
        transition={{ duration: 20, repeat: Infinity, ease: "easeInOut", delay: 4 }}
      />
    </div>
  );
}
