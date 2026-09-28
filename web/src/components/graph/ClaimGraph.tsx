"use client";

import { useEffect, useMemo, useState } from "react";
import ReactFlow, {
  Background,
  BackgroundVariant,
  Controls,
  MarkerType,
  useNodesState,
  type Edge,
  type Node,
} from "reactflow";
import "reactflow/dist/style.css";
import { ClaimNode, type ClaimNodeData } from "./ClaimNode";
import { EdgeLegend } from "./EdgeLegend";
import { useLanguage } from "@/lib/language";
import { edgeColor } from "@/lib/utils";
import type { GraphEdge, GraphNode } from "@/types";

const nodeTypes = { claim: ClaimNode };

function layout(nodes: GraphNode[]): Map<string, { x: number; y: number }> {
  const byDomain = new Map<string, GraphNode[]>();
  for (const n of nodes) {
    const arr = byDomain.get(n.domain) ?? [];
    arr.push(n);
    byDomain.set(n.domain, arr);
  }
  const positions = new Map<string, { x: number; y: number }>();
  const domains = [...byDomain.keys()];
  const clusterR = Math.max(360, domains.length * 130);
  const cx = clusterR + 200;
  const cy = clusterR + 100;

  domains.forEach((domain, di) => {
    const angle = (di / Math.max(1, domains.length)) * Math.PI * 2;
    const clx = cx + Math.cos(angle) * clusterR;
    const cly = cy + Math.sin(angle) * clusterR;
    const members = byDomain.get(domain)!;
    const r = Math.max(120, members.length * 42);
    members.forEach((m, mi) => {
      if (members.length === 1) {
        positions.set(m.id, { x: clx, y: cly });
        return;
      }
      const a = (mi / members.length) * Math.PI * 2;
      positions.set(m.id, {
        x: clx + Math.cos(a) * r,
        y: cly + Math.sin(a) * r,
      });
    });
  });
  return positions;
}

export function ClaimGraph({
  nodes,
  edges,
}: {
  nodes: GraphNode[];
  edges: GraphEdge[];
}) {
  const { t } = useLanguage();
  const [focusedId, setFocusedId] = useState<string | null>(null);
  const relatedIds = useMemo(() => {
    if (!focusedId) return new Set<string>();
    const related = new Set([focusedId]);
    for (const edge of edges) {
      if (edge.source === focusedId) related.add(edge.target);
      if (edge.target === focusedId) related.add(edge.source);
    }
    return related;
  }, [edges, focusedId]);

  const flowNodes = useMemo<Node<ClaimNodeData>[]>(() => {
    const pos = layout(nodes);
    return nodes.map((n) => ({
      id: n.id,
      type: "claim",
      position: pos.get(n.id) ?? { x: 0, y: 0 },
      data: {
        domain: n.domain,
        topic: n.topic,
        claim: n.claim,
        claimType: n.claim_type,
        importance: n.importance,
        edgeCount: n.edge_count,
        focused: focusedId === n.id,
        related: !focusedId || relatedIds.has(n.id),
      },
      style: {
        opacity: !focusedId || relatedIds.has(n.id) ? 1 : 0.16,
        transition: "opacity 220ms ease, filter 220ms ease",
        filter: focusedId === n.id ? "drop-shadow(0 0 12px rgba(139,124,255,0.55))" : undefined,
        zIndex: focusedId === n.id ? 2 : 1,
      },
    }));
  }, [focusedId, nodes, relatedIds]);

  const [visibleNodes, setVisibleNodes, onNodesChange] = useNodesState<ClaimNodeData>(flowNodes);
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
    const ids = new Set(nodes.map((n) => n.id));
    return edges
      .filter((e) => ids.has(e.source) && ids.has(e.target))
      .map((e) => {
        const color = edgeColor(e.type);
        return {
          id: e.id,
          source: e.source,
          target: e.target,
          animated: e.status === "active",
          style: {
            stroke: color,
            strokeWidth: 1.5,
            strokeDasharray: e.status === "candidate" ? "4 4" : undefined,
            opacity: !focusedId || e.source === focusedId || e.target === focusedId ? 1 : 0.1,
            transition: "opacity 220ms ease, stroke-width 220ms ease",
            ...(focusedId && (e.source === focusedId || e.target === focusedId) ? { strokeWidth: 2.6 } : {}),
          },
          markerEnd: { type: MarkerType.ArrowClosed, color, width: 14, height: 14 },
          label: e.type,
          labelStyle: { fill: color, fontSize: 9, fontWeight: 300 },
          labelBgStyle: { fill: "rgba(10,14,30,0.85)" },
          labelBgPadding: [4, 2] as [number, number],
          labelBgBorderRadius: 4,
        };
      });
  }, [edges, focusedId, nodes]);

  return (
    <div className="relative h-full w-full">
      <EdgeLegend />
      <div className="glass-strong absolute right-4 top-4 z-10 rounded-[12px] px-3.5 py-3 text-[10px] text-ink-muted">
        {t("Kéo node để sắp xếp · kéo nền để di chuyển khung")}
      </div>
      <ReactFlow
        nodes={visibleNodes}
        onNodesChange={onNodesChange}
        edges={flowEdges}
        nodeTypes={nodeTypes}
        nodesDraggable
        autoPanOnNodeDrag
        onNodeClick={(_, node) => setFocusedId((current) => current === node.id ? null : node.id)}
        fitView
        fitViewOptions={{ padding: 0.25 }}
        minZoom={0.15}
        maxZoom={1.6}
        proOptions={{ hideAttribution: true }}
        className="bg-transparent"
      >
        <Background
          variant={BackgroundVariant.Dots}
          gap={28}
          size={1}
          color="rgba(255,255,255,0.06)"
        />
        <Controls
          className="!rounded-[10px] !border !border-border !bg-[rgba(20,26,48,0.8)] !shadow-card [&_button]:!border-border [&_button]:!bg-transparent [&_button]:!text-ink [&_button:hover]:!bg-white/10 [&_path]:!fill-ink"
          showInteractive={false}
        />
      </ReactFlow>
    </div>
  );
}
