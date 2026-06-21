"use client";

import { useQuery } from "@tanstack/react-query";
import {
  getContext,
  getContexts,
  getDomains,
  getGraphEdges,
  getGraphNodes,
  getStatus,
  search,
} from "@/lib/api";

export function useStatus() {
  return useQuery({ queryKey: ["status"], queryFn: getStatus, refetchInterval: 8000 });
}

export function useContexts(domain?: string, topic?: string) {
  return useQuery({
    queryKey: ["contexts", domain ?? "", topic ?? ""],
    queryFn: () => getContexts(domain, topic),
  });
}

export function useContext(id: string) {
  return useQuery({
    queryKey: ["context", id],
    queryFn: () => getContext(id),
    enabled: !!id,
  });
}

export function useSearch(q: string, limit = 10) {
  return useQuery({
    queryKey: ["search", q, limit],
    queryFn: () => search(q, limit),
    enabled: q.trim().length > 0,
    staleTime: 30_000,
  });
}

export function useGraph() {
  const nodes = useQuery({ queryKey: ["graph-nodes"], queryFn: getGraphNodes });
  const edges = useQuery({ queryKey: ["graph-edges"], queryFn: getGraphEdges });
  return { nodes, edges };
}

export function useDomains() {
  return useQuery({ queryKey: ["domains"], queryFn: getDomains });
}
