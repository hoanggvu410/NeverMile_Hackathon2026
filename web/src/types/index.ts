export interface Status {
  repository: string;
  branch: string;
  git_root: string;
  context_count: number;
  pending_commits: string[];
  graph_ready: boolean;
}

export interface ContextSummary {
  id: string;
  title: string;
  domain: string;
  topic: string;
  date: string;
  prompt: string;
}

export interface FileEntry {
  file: string;
  status: string;
  description: string;
}

export interface ContextDetail {
  id: string;
  title: string;
  saved_by: string;
  agent: string;
  repository: string;
  branch: string;
  date: string;
  domain: string;
  topic: string;
  prompt: string;
  what_was_done: string;
  reasoning: string;
  key_decisions: string;
  rejected_alternatives: string;
  risks_and_open_questions: string;
  verification: string;
  files: FileEntry[];
  commits: string[];
}

export interface SearchResult {
  id: string;
  claim_id?: string;
  domain: string;
  topic: string;
  title: string;
  prompt: string;
  date: string;
  score: number;
  edge_type?: string;
  depth?: number;
  claim?: string;
  claim_type?: string;
  vector_kind?: string;
}

export interface GraphNode {
  id: string;
  session_id: string;
  domain: string;
  topic: string;
  title: string;
  claim: string;
  claim_type: string;
  importance: number;
  edge_count: number;
}

export interface GraphEdge {
  id: string;
  source: string;
  target: string;
  type: string;
  confidence: number;
  status: string;
}
