"use client";

import { useEffect, useMemo, useState } from "react";
import ReactFlow, {
  Background,
  BackgroundVariant,
  Controls,
  Handle,
  MarkerType,
  Position,
  useNodesState,
  type Edge,
  type Node,
} from "reactflow";
import "reactflow/dist/style.css";
import type { GitCommit } from "@/types";
import { shortSha, truncate } from "@/lib/utils";
import { useLanguage } from "@/lib/language";

interface CommitNodeData {
  commit: GitCommit;
  isHead: boolean;
  focused: boolean;
}

function CommitNodeView({ data }: { data: CommitNodeData }) {
  const { commit, isHead, focused } = data;
  const { locale } = useLanguage();
  return (
    <div className="glass-strong w-[260px] cursor-pointer rounded-[10px] border border-accent/40 px-3 py-2.5 shadow-card transition-[box-shadow,border-color] duration-200" style={{ boxShadow: focused ? "0 0 0 1px rgba(139,124,255,.9), 0 0 24px rgba(139,124,255,.42)" : undefined }}>
      <Handle type="target" position={Position.Top} className="!h-1.5 !w-1.5 !border-0 !bg-accent" />
      <div className="flex items-center justify-between gap-2">
        <code className="text-[10px] font-medium text-accent">{shortSha(commit.sha)}</code>
        {isHead && <span className="rounded-full bg-accent/15 px-2 py-0.5 text-[9px] text-accent">HEAD</span>}
        <time className="text-[9px] text-ink-muted">
          {new Date(commit.date).toLocaleDateString(locale === "en" ? "en-US" : "vi-VN", { day: "numeric", month: "short", year: "numeric" })}
        </time>
      </div>
      <p className="mt-1 text-[11px] leading-snug text-ink/90" title={commit.subject}>
        {truncate(commit.subject, 62)}
      </p>
      <div className="mt-1 truncate text-[9px] text-ink-muted">{commit.author}</div>
      <Handle type="source" position={Position.Bottom} className="!h-1.5 !w-1.5 !border-0 !bg-accent" />
    </div>
  );
}

const nodeTypes = { commit: CommitNodeView };

function commitLanes(commits: GitCommit[]) {
  const lanes = new Map<string, number>();
  const positions = new Map<string, { x: number; y: number }>();
  let nextLane = 0;

  commits.forEach((commit, row) => {
    let lane = lanes.get(commit.sha);
    if (lane === undefined) lane = nextLane++;
    lanes.delete(commit.sha);
    positions.set(commit.sha, { x: lane * 320, y: row * 116 });

    commit.parents.forEach((parent, index) => {
      if (!lanes.has(parent)) lanes.set(parent, index === 0 ? lane : nextLane++);
    });
  });
  return positions;
}

export function CommitGraph({ commits }: { commits: GitCommit[] }) {
  const { t } = useLanguage();
  const [focusedId, setFocusedId] = useState<string | null>(null);
  const relatedIds = useMemo(() => {
    if (!focusedId) return new Set<string>();
    const related = new Set([focusedId]);
    for (const commit of commits) {
      if (commit.sha === focusedId) commit.parents.forEach((parent) => related.add(parent));
      if (commit.parents.includes(focusedId)) related.add(commit.sha);
    }
    return related;
  }, [commits, focusedId]);

  const flowNodes = useMemo<Node<CommitNodeData>[]>(() => {
    const positions = commitLanes(commits);
    return commits.map((commit, index) => ({
      id: commit.sha,
      type: "commit",
      position: positions.get(commit.sha) ?? { x: 0, y: index * 116 },
      data: {
        commit,
        isHead: index === 0,
        focused: focusedId === commit.sha,
      },
      style: {
        opacity: !focusedId || relatedIds.has(commit.sha) ? 1 : 0.15,
        transition: "opacity 220ms ease, filter 220ms ease",
        zIndex: focusedId === commit.sha ? 2 : 1,
      },
    }));
  }, [commits, focusedId, relatedIds]);

  const [visibleNodes, setVisibleNodes, onNodesChange] = useNodesState<CommitNodeData>(flowNodes);
  useEffect(() => {
    setVisibleNodes((current) => {
      const previous = new Map(current.map((node) => [node.id, node]));
      return flowNodes.map((node) => {
        const existing = previous.get(node.id);
        return existing
          ? { ...node, position: existing.position, selected: existing.selected, dragging: existing.dragging }
          : node;
      });
    });
  }, [flowNodes, setVisibleNodes]);

  const flowEdges = useMemo<Edge[]>(() => {
    const known = new Set(commits.map((commit) => commit.sha));
    return commits.flatMap((commit) =>
      commit.parents
        .filter((parent) => known.has(parent))
        .map((parent) => ({
          id: `${commit.sha}-${parent}`,
          source: commit.sha,
          target: parent,
          type: "smoothstep",
          style: {
            stroke: "#8b7cff",
            strokeWidth: focusedId && (commit.sha === focusedId || parent === focusedId) ? 2.7 : 1.6,
            opacity: !focusedId || commit.sha === focusedId || parent === focusedId ? 1 : 0.1,
            transition: "opacity 220ms ease, stroke-width 220ms ease",
          },
          markerEnd: { type: MarkerType.ArrowClosed, color: "#8b7cff", width: 12, height: 12 },
        }))
    );
  }, [commits, focusedId]);

  return (
    <div className="relative h-full w-full">
      <div className="glass-strong absolute right-4 top-4 z-10 rounded-[12px] px-3.5 py-3 text-[11px] text-ink-muted">
        {t("Đường nối biểu thị quan hệ parent trong Git")}
        <div className="mt-1 text-[10px]">{t("Kéo node để sắp xếp · kéo nền để di chuyển khung")}</div>
      </div>
      <ReactFlow
        nodes={visibleNodes}
        onNodesChange={onNodesChange}
        edges={flowEdges}
        nodeTypes={nodeTypes}
        nodesDraggable
        autoPanOnNodeDrag
        onNodeClick={(_, node) => setFocusedId((current) => current === node.id ? null : node.id)}
        defaultViewport={{ x: 72, y: 84, zoom: 0.68 }}
        minZoom={0.12}
        maxZoom={1.5}
        proOptions={{ hideAttribution: true }}
        className="bg-transparent"
      >
        <Background variant={BackgroundVariant.Dots} gap={28} size={1} color="rgba(255,255,255,0.06)" />
        <Controls
          className="!rounded-[10px] !border !border-border !bg-[rgba(20,26,48,0.8)] !shadow-card [&_button]:!border-border [&_button]:!bg-transparent [&_button]:!text-ink [&_button:hover]:!bg-white/10 [&_path]:!fill-ink"
          showInteractive={false}
        />
      </ReactFlow>
    </div>
  );
}
