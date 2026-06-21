"use client";

import { usePathname } from "next/navigation";
import { useStatus } from "@/hooks/useGitWhy";
import { IconBell, IconChevronDown, IconSearch } from "@/components/ui/icons";

function pageTitle(pathname: string): string {
  if (pathname === "/dashboard") return "Tổng quan";
  if (pathname.startsWith("/dashboard/contexts")) return "Ngữ cảnh";
  if (pathname.startsWith("/dashboard/search")) return "Tìm kiếm";
  if (pathname.startsWith("/dashboard/graph")) return "Đồ thị";
  return "Tổng quan";
}

export function TopBar() {
  const pathname = usePathname();
  const { data: status } = useStatus();
  const title = pageTitle(pathname);

  return (
    <header className="flex h-14 shrink-0 items-center justify-between px-6">
      <div className="flex items-center gap-2 text-[13px] text-ink-muted">
        <span className="hover:text-ink">Trang chủ</span>
        <span className="opacity-40">/</span>
        <span className="text-ink">{title}</span>
      </div>

      <div className="flex items-center gap-2.5">
        <button className="glass-inner flex items-center gap-2 rounded-[10px] px-3 py-1.5 text-[12px] text-ink transition-colors duration-150 hover:border-accent/50">
          <span className="h-1.5 w-1.5 rounded-full bg-success shadow-[0_0_8px] shadow-success" />
          <span className="font-light">
            {status?.repository ?? "repo nội bộ"}
          </span>
          <span className="opacity-40">/</span>
          <span className="text-ink-muted">{status?.branch ?? "main"}</span>
          <IconChevronDown className="h-3.5 w-3.5 opacity-60" />
        </button>

        <button className="glass-inner flex h-9 w-9 items-center justify-center rounded-[10px] text-ink-muted transition-colors duration-150 hover:text-ink">
          <IconSearch className="h-[18px] w-[18px]" />
        </button>
        <button className="glass-inner relative flex h-9 w-9 items-center justify-center rounded-[10px] text-ink-muted transition-colors duration-150 hover:text-ink">
          <IconBell className="h-[18px] w-[18px]" />
          {status && status.pending_commits.length > 0 && (
            <span className="absolute right-2 top-2 h-1.5 w-1.5 rounded-full bg-danger" />
          )}
        </button>
        <div className="ml-1 flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-accent to-accent-2 text-[12px] font-medium text-white">
          GW
        </div>
      </div>
    </header>
  );
}
