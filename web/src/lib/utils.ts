import { clsx, type ClassValue } from "clsx";

export function cn(...inputs: ClassValue[]) {
  return clsx(inputs);
}

export function timeAgo(iso: string): string {
  if (!iso) return "";
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "";
  const secs = Math.floor((Date.now() - then) / 1000);
  if (secs < 60) return "vừa xong";
  const mins = Math.floor(secs / 60);
  if (mins < 60) return `${mins} phút trước`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours} giờ trước`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days} ngày trước`;
  const months = Math.floor(days / 30);
  if (months < 12) return `${months} tháng trước`;
  return `${Math.floor(months / 12)} năm trước`;
}

export function formatDate(iso: string): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleDateString("vi-VN", {
    day: "numeric",
    month: "short",
    year: "numeric",
  });
}

export function truncate(s: string, n: number): string {
  if (!s) return "";
  const clean = s.replace(/\s+/g, " ").trim();
  return clean.length > n ? clean.slice(0, n).trimEnd() + "…" : clean;
}

export function shortSha(sha: string): string {
  return sha.slice(0, 7);
}

const PALETTE = [
  "#7c5cff",
  "#3a8dff",
  "#22c1a6",
  "#e0794a",
  "#d65db1",
  "#5b8def",
  "#9b6aff",
];

/** Deterministic accent color from a domain string. */
export function domainColor(domain: string): string {
  let h = 0;
  for (let i = 0; i < domain.length; i++) {
    h = (h * 31 + domain.charCodeAt(i)) >>> 0;
  }
  return PALETTE[h % PALETTE.length];
}

export const EDGE_COLORS: Record<string, string> = {
  CAUSED_BY: "#7c5cff",
  CONSTRAINS: "#e0a23a",
  IMPLEMENTS: "#34c77b",
  SUPERSEDES: "#e0556a",
  CONFLICTS_WITH: "#e0794a",
  RELATED_CANDIDATE: "#7a8095",
};

export function edgeColor(type?: string): string {
  if (!type) return EDGE_COLORS.RELATED_CANDIDATE;
  return EDGE_COLORS[type] ?? EDGE_COLORS.RELATED_CANDIDATE;
}

export function scoreColor(score: number): string {
  if (score > 0.7) return "var(--success)";
  if (score > 0.4) return "var(--warning)";
  return "var(--danger)";
}
