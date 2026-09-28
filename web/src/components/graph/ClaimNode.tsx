"use client";

import { Handle, Position } from "reactflow";
import { domainColor, truncate } from "@/lib/utils";
import { useLanguage } from "@/lib/language";

export interface ClaimNodeData {
  domain: string;
  topic: string;
  claim: string;
  claimType: string;
  importance: number;
  edgeCount: number;
  focused?: boolean;
  related?: boolean;
}

export function ClaimNode({ data }: { data: ClaimNodeData }) {
  const { t } = useLanguage();
  const color = domainColor(data.domain);
  const scale = Math.min(1.4, 1 + data.edgeCount * 0.08);

  return (
    <div
      className="glass-strong rounded-[10px] px-3 py-2.5 shadow-card"
      style={{
        width: 184,
        borderColor: `${color}66`,
        borderLeft: `3px solid ${color}`,
        transform: `scale(${scale})`,
        transition: "box-shadow 220ms ease, border-color 220ms ease",
        boxShadow: data.focused ? `0 0 0 1px ${color}, 0 0 22px ${color}66` : undefined,
      }}
    >
      <Handle type="target" position={Position.Left} className="!h-1.5 !w-1.5 !border-0 !bg-white/30" />
      <div className="mb-1 flex items-center gap-1.5">
        <span
          className="h-1.5 w-1.5 rounded-full"
          style={{ background: color }}
        />
        <span className="truncate text-[10px] uppercase tracking-wider" style={{ color }}>
          {data.domain}
        </span>
      </div>
      <p className="text-[11.5px] leading-snug text-ink/90">
        {truncate(data.claim, 70)}
      </p>
      {data.edgeCount > 0 && (
        <div className="mt-1.5 text-[9.5px] text-ink-muted">
          {data.edgeCount} {t("liên kết")}
        </div>
      )}
      <Handle type="source" position={Position.Right} className="!h-1.5 !w-1.5 !border-0 !bg-white/30" />
    </div>
  );
}
