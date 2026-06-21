"use client";

import { useMemo } from "react";
import ReactFlow, {
  Background,
  BackgroundVariant,
  Controls,
  MarkerType,
  type Edge,
  type Node,
} from "reactflow";
import "reactflow/dist/style.css";
import { ClaimNode, type ClaimNodeData } from "./ClaimNode";
import { EdgeLegend } from "./EdgeLegend";
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
      },
    }));
  }, [nodes]);

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
          },
          markerEnd: { type: MarkerType.ArrowClosed, color, width: 14, height: 14 },
          label: e.type,
          labelStyle: { fill: color, fontSize: 9, fontWeight: 300 },
          labelBgStyle: { fill: "rgba(10,14,30,0.85)" },
          labelBgPadding: [4, 2] as [number, number],
          labelBgBorderRadius: 4,
        };
      });
  }, [edges, nodes]);

  return (
    <div className="relative h-full w-full">
      <EdgeLegend />
      <ReactFlow
        nodes={flowNodes}
        edges={flowEdges}
        nodeTypes={nodeTypes}
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
