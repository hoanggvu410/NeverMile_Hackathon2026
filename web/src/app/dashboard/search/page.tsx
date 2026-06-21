"use client";

import { useEffect, useState } from "react";
import { motion } from "framer-motion";
import { useSearch } from "@/hooks/useGitWhy";
import { SearchResultCard } from "@/components/search/SearchResultCard";
import { CardSkeleton } from "@/components/ui/Skeleton";
import { EmptyState } from "@/components/ui/EmptyState";
import { IconSearch } from "@/components/ui/icons";

const SUGGESTIONS = [
  "Vì sao chọn claim-level retrieval?",
  "ổn định embedding provider",
  "phương án đã loại cho tripwire",
];

export default function SearchPage() {
  const [input, setInput] = useState("");
  const [query, setQuery] = useState("");
  const { data: results, isFetching } = useSearch(query);

  useEffect(() => {
    const t = setTimeout(() => setQuery(input), 300);
    return () => clearTimeout(t);
  }, [input]);

  const hasQuery = query.trim().length > 0;
  const empty = hasQuery && !isFetching && (results?.length ?? 0) === 0;

  return (
    <main className="h-full overflow-y-auto px-6 py-8 lg:px-8">
      <div className="mx-auto max-w-[760px]">
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.35 }}
          className={hasQuery ? "mb-6" : "mb-6 mt-[8vh] text-center"}
        >
          {!hasQuery && (
            <>
              <h1 className="mb-2 text-[28px] font-light text-ink">
                Tìm kiếm quyết định
              </h1>
              <p className="mb-6 text-[14px] text-ink-muted">
                Hỏi về các quyết định trong codebase. Tìm theo ngữ nghĩa trên
                claim graph.
              </p>
            </>
          )}

          <div className="glass-strong flex h-[52px] items-center gap-3 rounded-[14px] px-4 transition duration-150 focus-within:border-accent focus-within:shadow-glow">
            <IconSearch className="h-5 w-5 text-ink-muted" />
            <input
              autoFocus
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Vì sao chọn X thay vì Y?"
              className="h-full flex-1 bg-transparent text-[15px] font-light text-ink outline-none placeholder:text-ink-muted/60"
            />
            {isFetching && (
              <span className="h-4 w-4 animate-spin rounded-full border-2 border-accent border-t-transparent" />
            )}
          </div>

          {!hasQuery && (
            <div className="mt-4 flex flex-wrap justify-center gap-2">
              {SUGGESTIONS.map((s) => (
                <button
                  key={s}
                  onClick={() => setInput(s)}
                  className="glass-inner rounded-full px-3.5 py-1.5 text-[12px] text-ink-muted transition-colors hover:text-ink"
                >
                  {s}
                </button>
              ))}
            </div>
          )}
        </motion.div>

        {isFetching && hasQuery && (
          <div className="flex flex-col gap-3.5">
            <CardSkeleton />
            <CardSkeleton />
          </div>
        )}

        {!isFetching && results && results.length > 0 && (
          <div className="flex flex-col gap-3.5 pb-8">
            <div className="mb-1 text-[12px] text-ink-muted">
              {results.length} kết quả
            </div>
            {results.map((r, i) => (
              <SearchResultCard key={`${r.claim_id ?? r.id}-${i}`} result={r} index={i} />
            ))}
          </div>
        )}

        {empty && (
          <EmptyState
            icon={<IconSearch className="h-6 w-6" />}
            title="Không tìm thấy kết quả"
            description="Thử diễn đạt khác, hoặc mở rộng câu hỏi."
          />
        )}
      </div>
    </main>
  );
}
