import { cn } from "@/lib/utils";

export function Skeleton({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        "animate-pulse rounded-[10px] bg-white/[0.04]",
        className
      )}
    />
  );
}

export function CardSkeleton() {
  return (
    <div className="glass rounded-[14px] p-5">
      <Skeleton className="mb-3 h-3 w-24" />
      <Skeleton className="mb-2 h-5 w-2/3" />
      <Skeleton className="mb-4 h-3 w-full" />
      <div className="flex gap-3">
        <Skeleton className="h-3 w-16" />
        <Skeleton className="h-3 w-16" />
      </div>
    </div>
  );
}
