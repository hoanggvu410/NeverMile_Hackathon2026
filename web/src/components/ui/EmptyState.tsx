import type { ReactNode } from "react";

interface EmptyStateProps {
  icon?: ReactNode;
  title: string;
  description?: string;
  children?: ReactNode;
}

export function EmptyState({ icon, title, description, children }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center rounded-[14px] px-6 py-16 text-center">
      {icon && (
        <div className="mb-4 flex h-14 w-14 items-center justify-center rounded-full glass-inner text-ink-muted">
          {icon}
        </div>
      )}
      <h3 className="text-lg font-light text-ink">{title}</h3>
      {description && (
        <p className="mt-1.5 max-w-sm text-[13px] text-ink-muted">{description}</p>
      )}
      {children && <div className="mt-5">{children}</div>}
    </div>
  );
}
