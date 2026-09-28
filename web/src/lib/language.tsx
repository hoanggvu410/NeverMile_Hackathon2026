"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";

export type Locale = "vi" | "en";

const english: Record<string, string> = {
  "Trang chủ": "Home",
  "Tổng quan": "Overview",
  "Ngữ cảnh": "Contexts",
  "Tìm kiếm": "Search",
  "Đồ thị": "Graph",
  "Cài đặt": "Settings",
  "Đăng xuất": "Sign out",
  "repo nội bộ": "Local repository",
  "Vì sao build cái này?": "Why did we build this?",
  "Lý do, đánh đổi và hướng đã loại bỏ sau mỗi quyết định agent đưa ra.": "The reasoning, trade-offs, and rejected paths behind every agent decision.",
  "Mật độ tri thức": "Knowledge density",
  "Claims theo domain, đậm dần theo độ quan trọng": "Claims by domain, shaded by importance",
  "Mở đồ thị": "Open graph",
  "Thấp": "Low",
  "Vừa": "Medium",
  "Cao": "High",
  "Tối đa": "Highest",
  "Tất cả": "All",
  "Lưu Context": "Save context",
  "Lưu từ agent": "Save from an agent",
  "Cho agent gọi": "Have your agent call",
  "hoặc chạy": "or run",
  "Đã cài MCP server": "MCP server installed",
  "Đã cài post-commit hook": "Post-commit hook installed",
  "Đã lưu ngữ cảnh đầu tiên": "First context saved",
  "Chạy tìm kiếm": "Run a search",
  "Bắt đầu với GitWhy": "Get started with GitWhy",
  "Vài bước để bắt đầu ghi lại lý do đằng sau code của bạn.": "A few steps to start recording the reasoning behind your code.",
  "Ngữ cảnh mới nhất": "Latest context",
  "Xem": "View",
  "Chưa có ngữ cảnh nào.": "No contexts yet.",
  "Cạnh đồ thị": "Graph edges",
  "Đang chờ": "Pending",
  "Sức khỏe đồ thị": "Graph health",
  "Độ phủ claims so với kỳ vọng": "Claim coverage versus expected",
  "Graph engine đang chạy": "Graph engine is running",
  "Bộ nhớ quyết định": "Decision memory",
  "Mọi claim, cạnh và đánh đổi agent ghi lại, sẵn sàng truy xuất trước quyết định kế tiếp.": "Every claim, relationship, and trade-off recorded by an agent, ready to retrieve before the next decision.",
  "Mọi quyết định đã lưu, theo domain & topic.": "All saved decisions, organized by domain and topic.",
  "Chưa có ngữ cảnh trong domain này": "No contexts in this domain",
  "Lưu ngữ cảnh từ agent để thấy nó xuất hiện ở đây.": "Save a context from an agent to see it appear here.",
  "Đã làm gì": "What was done",
  "Lý do": "Reasoning",
  "Quyết định chính": "Key decisions",
  "Phương án đã loại": "Rejected alternatives",
  "Rủi ro & câu hỏi mở": "Risks & open questions",
  "Kiểm chứng": "Verification",
  "Không tìm thấy ngữ cảnh.": "Context not found.",
  "Về danh sách ngữ cảnh": "Back to contexts",
  "Quay lại": "Back",
  "lưu bởi": "saved by",
  "Đã chép": "Copied",
  "Chép ID": "Copy ID",
  "Yêu cầu": "Request",
  "Tệp": "Files",
  "Tìm kiếm quyết định": "Search decisions",
  "Hỏi về các quyết định trong codebase. Tìm theo ngữ nghĩa trên claim graph.": "Ask about decisions in your codebase. Search semantically across the claim graph.",
  "Vì sao chọn X thay vì Y?": "Why choose X over Y?",
  "Vì sao chọn claim-level retrieval?": "Why use claim-level retrieval?",
  "ổn định embedding provider": "embedding provider stability",
  "phương án đã loại cho tripwire": "rejected alternatives for the tripwire",
  "Không tìm thấy kết quả": "No results found",
  "Thử diễn đạt khác, hoặc mở rộng câu hỏi.": "Try rephrasing or broadening your query.",
  "kết quả": "results",
  "từ “": "from “",
  "khớp": "match",
  "Mở ngữ cảnh": "Open context",
  "Đồ thị claim": "Claim graph",
  "cạnh": "edges",
  "Không tìm thấy commit": "No commits found",
  "Không thể đọc lịch sử Git của project đang kết nối.": "Could not read Git history for the connected project.",
  "Project này chưa có commit history.": "This project has no commit history.",
  "Lưu ngữ cảnh từ agent để dựng claim graph.": "Save a context from an agent to build the claim graph.",
  "Chưa có claim nào": "No claims yet",
  "Loại cạnh": "Edge types",
  "liên kết": "links",
  "Đường nối biểu thị quan hệ parent trong Git": "Edges show parent relationships in Git",
  "Kéo node để sắp xếp · kéo nền để di chuyển khung": "Drag nodes to rearrange · drag the canvas to pan",
  "vừa xong": "just now",
  "phút trước": "minutes ago",
  "giờ trước": "hours ago",
  "ngày trước": "days ago",
  "tháng trước": "months ago",
  "năm trước": "years ago",
  "Switch to English": "Switch to English",
  "Switch to Vietnamese": "Switch to Vietnamese",
  "Chưa ghi nhận.": "Nothing recorded yet.",
};

interface LanguageContextValue {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: (text: string) => string;
}

const LanguageContext = createContext<LanguageContextValue>({
  locale: "vi",
  setLocale: () => undefined,
  t: (text) => text,
});

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>("vi");
  const [hydrated, setHydrated] = useState(false);

  const setLocale = useCallback((nextLocale: Locale) => {
    window.localStorage.setItem("gitwhy-language", nextLocale);
    document.documentElement.lang = nextLocale;
    setLocaleState(nextLocale);
  }, []);

  useEffect(() => {
    const saved = window.localStorage.getItem("gitwhy-language");
    if (saved === "en" || saved === "vi") setLocaleState(saved);
    setHydrated(true);
  }, []);

  useEffect(() => {
    document.documentElement.lang = locale;
    if (hydrated) window.localStorage.setItem("gitwhy-language", locale);
  }, [locale, hydrated]);

  const t = useCallback(
    (text: string) => (locale === "en" ? english[text] ?? text : text),
    [locale]
  );
  const value = useMemo(() => ({ locale, setLocale, t }), [locale, t]);

  return <LanguageContext.Provider value={value}>{children}</LanguageContext.Provider>;
}

export function useLanguage() {
  return useContext(LanguageContext);
}
