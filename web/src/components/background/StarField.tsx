"use client";

import { useEffect, useRef } from "react";

interface Star {
  x: number;
  y: number;
  r: number;
  base: number;
  amp: number;
  phase: number;
  speed: number;
}

export function StarField() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const mouse = useRef({ x: 0, y: 0 });

  useEffect(() => {
    const reduce = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;

    if (!canvasRef.current) return;
    const cv: HTMLCanvasElement = canvasRef.current;
    const context = cv.getContext("2d");
    if (!context) return;
    const ctx: CanvasRenderingContext2D = context;

    let stars: Star[] = [];
    let raf = 0;
    let dpr = Math.min(window.devicePixelRatio || 1, 2);

    function resize() {
      dpr = Math.min(window.devicePixelRatio || 1, 2);
      cv.width = window.innerWidth * dpr;
      cv.height = window.innerHeight * dpr;
      cv.style.width = window.innerWidth + "px";
      cv.style.height = window.innerHeight + "px";
      const count = Math.floor((window.innerWidth * window.innerHeight) / 4200);
      stars = Array.from({ length: Math.min(count, 320) }, () => {
        const bright = Math.random() > 0.88;
        return {
          x: Math.random() * cv.width,
          y: Math.random() * cv.height * 0.88,
          r: (bright ? Math.random() * 1.4 + 1.1 : Math.random() * 1.1 + 0.4) * dpr,
          base: bright ? Math.random() * 0.3 + 0.65 : Math.random() * 0.45 + 0.3,
          amp: Math.random() * 0.45 + 0.15,
          phase: Math.random() * Math.PI * 2,
          speed: Math.random() * 0.001 + 0.0004,
        };
      });
    }

    function draw(t: number) {
      ctx.clearRect(0, 0, cv.width, cv.height);
      const ox = (mouse.current.x - 0.5) * 18 * dpr;
      const oy = (mouse.current.y - 0.5) * 12 * dpr;
      for (const s of stars) {
        const tw = reduce ? s.base : s.base + Math.sin(t * s.speed + s.phase) * s.amp;
        const op = Math.max(0, Math.min(1, tw));
        const px = s.x + ox * (s.r / dpr);
        const py = s.y + oy * (s.r / dpr);
        if (s.r > 1.6 * dpr) {
          ctx.shadowColor = "rgba(180, 200, 255, 0.9)";
          ctx.shadowBlur = 6 * dpr;
        } else {
          ctx.shadowBlur = 0;
        }
        ctx.beginPath();
        ctx.arc(px, py, s.r, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(226, 233, 255, ${op})`;
        ctx.fill();
      }
      ctx.shadowBlur = 0;
      if (!reduce) raf = requestAnimationFrame(draw);
    }

    function onMove(e: MouseEvent) {
      mouse.current.x = e.clientX / window.innerWidth;
      mouse.current.y = e.clientY / window.innerHeight;
    }

    resize();
    draw(0);
    window.addEventListener("resize", resize);
    if (!reduce) window.addEventListener("mousemove", onMove);

    return () => {
      cancelAnimationFrame(raf);
      window.removeEventListener("resize", resize);
      window.removeEventListener("mousemove", onMove);
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      aria-hidden
      className="pointer-events-none fixed inset-0 z-0"
    />
  );
}
