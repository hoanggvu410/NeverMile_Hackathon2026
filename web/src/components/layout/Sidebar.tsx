"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { motion } from "framer-motion";
import { cn } from "@/lib/utils";
import {
  IconArrowRight,
  IconContexts,
  IconDashboard,
  IconGraph,
  IconLogout,
  IconSearch,
  IconSettings,
} from "@/components/ui/icons";
import type { ComponentType, SVGProps } from "react";

interface NavItem {
  href: string;
  label: string;
  icon: ComponentType<SVGProps<SVGSVGElement>>;
}

const NAV: NavItem[] = [
  { href: "/dashboard", label: "Tổng quan", icon: IconDashboard },
  { href: "/dashboard/contexts", label: "Ngữ cảnh", icon: IconContexts },
  { href: "/dashboard/search", label: "Tìm kiếm", icon: IconSearch },
  { href: "/dashboard/graph", label: "Đồ thị", icon: IconGraph },
];

function NavLink({ item, active }: { item: NavItem; active: boolean }) {
  const Icon = item.icon;
  return (
    <Link href={item.href} className="group relative block">
      {active && (
        <motion.span
          layoutId="nav-active"
          className="absolute inset-0 rounded-[10px] bg-accent shadow-[0_8px_24px_-8px_var(--accent-glow)]"
          transition={{ type: "spring", stiffness: 380, damping: 32 }}
        />
      )}
      <span
        className={cn(
          "relative flex items-center gap-3 rounded-[10px] px-3 py-2.5 text-[13px] transition duration-150 ease-out-soft",
          active
            ? "text-white"
            : "text-ink-muted hover:translate-x-0.5 hover:text-ink"
        )}
      >
        <Icon className="h-[18px] w-[18px] shrink-0" />
        <span className="font-light tracking-wide">{item.label}</span>
        {!active && (
          <IconArrowRight className="ml-auto h-3.5 w-3.5 -translate-x-1 opacity-0 transition duration-150 group-hover:translate-x-0 group-hover:opacity-60" />
        )}
      </span>
    </Link>
  );
}

export function Sidebar() {
  const pathname = usePathname();
  const isActive = (href: string) =>
    href === "/dashboard"
      ? pathname === "/dashboard"
      : pathname.startsWith(href);

  return (
    <aside className="flex h-full w-[220px] shrink-0 flex-col px-3.5 py-5">
      {/* Logo */}
      <Link href="/dashboard" className="mb-8 flex items-center gap-2.5 px-2">
        <span className="flex h-9 w-9 items-center justify-center rounded-[10px] bg-gradient-to-br from-accent to-accent-2 shadow-[0_8px_24px_-8px_var(--accent-glow)]">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
            <path
              d="M5 5l14 14M19 5L5 19"
              stroke="white"
              strokeWidth="2.2"
              strokeLinecap="round"
            />
          </svg>
        </span>
        <span className="text-[17px] font-light tracking-tight text-ink">
          Git<span className="text-accent">Why</span>
        </span>
      </Link>

      <nav className="flex flex-col gap-1">
        {NAV.map((item) => (
          <NavLink key={item.href} item={item} active={isActive(item.href)} />
        ))}
      </nav>

      <div className="my-5 h-px bg-border" />

      <div className="mt-auto flex flex-col gap-1">
        <button className="group flex items-center gap-3 rounded-[10px] px-3 py-2.5 text-[13px] text-ink-muted transition duration-150 hover:translate-x-0.5 hover:text-ink">
          <IconSettings className="h-[18px] w-[18px]" />
          <span className="font-light tracking-wide">Cài đặt</span>
        </button>
        <button className="group flex items-center gap-3 rounded-[10px] px-3 py-2.5 text-[13px] text-ink-muted transition duration-150 hover:translate-x-0.5 hover:text-ink">
          <IconLogout className="h-[18px] w-[18px]" />
          <span className="font-light tracking-wide">Đăng xuất</span>
        </button>
      </div>
    </aside>
  );
}
