import { EDGE_COLORS } from "@/lib/utils";

export function EdgeLegend() {
  return (
    <div className="glass-strong absolute left-4 top-4 z-10 rounded-[12px] p-3.5">
      <div className="mb-2 text-[11px] uppercase tracking-wider text-ink-muted">
        Loại cạnh
      </div>
      <div className="flex flex-col gap-1.5">
        {Object.entries(EDGE_COLORS).map(([type, color]) => (
          <div key={type} className="flex items-center gap-2 text-[11px] text-ink/80">
            <span
              className="h-0.5 w-5 rounded-full"
              style={{ background: color }}
            />
            {type}
          </div>
        ))}
      </div>
    </div>
  );
}
