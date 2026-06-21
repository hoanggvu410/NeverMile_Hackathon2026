import axios from "axios";
import type {
  ContextDetail,
  ContextSummary,
  GraphEdge,
  GraphNode,
  SearchResult,
  Status,
} from "@/types";

const BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:7420";

export const api = axios.create({
  baseURL: BASE_URL,
  timeout: 15000,
});

export async function getStatus(): Promise<Status> {
  const { data } = await api.get<Status>("/api/status");
  return data;
}

export async function getContexts(
  domain?: string,
  topic?: string
): Promise<ContextSummary[]> {
  const { data } = await api.get<ContextSummary[]>("/api/contexts", {
    params: { domain: domain || undefined, topic: topic || undefined },
  });
  return data ?? [];
}

export async function getContext(id: string): Promise<ContextDetail> {
  const { data } = await api.get<ContextDetail>(`/api/contexts/${id}`);
  return data;
}

export async function search(
  q: string,
  limit = 10
): Promise<SearchResult[]> {
  const { data } = await api.get<SearchResult[]>("/api/search", {
    params: { q, limit },
  });
  return data ?? [];
}

export async function getGraphNodes(): Promise<GraphNode[]> {
  const { data } = await api.get<GraphNode[]>("/api/graph/nodes");
  return data ?? [];
}

export async function getGraphEdges(): Promise<GraphEdge[]> {
  const { data } = await api.get<GraphEdge[]>("/api/graph/edges");
  return data ?? [];
}

export async function getDomains(): Promise<string[]> {
  const { data } = await api.get<string[]>("/api/domains");
  return data ?? [];
}
