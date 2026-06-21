import { Sidebar } from "@/components/layout/Sidebar";
import { TopBar } from "@/components/layout/TopBar";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen w-full p-3 sm:p-5 lg:p-6">
      <div className="glass-strong mx-auto flex h-[calc(100vh-2.5rem)] max-w-[1560px] overflow-hidden rounded-[24px] shadow-[0_40px_120px_-40px_rgba(0,0,0,0.8)]">
        <Sidebar />
        <div className="flex min-w-0 flex-1 flex-col border-l border-border/60">
          <TopBar />
          <div className="min-h-0 flex-1 overflow-hidden">{children}</div>
        </div>
      </div>
    </div>
  );
}
