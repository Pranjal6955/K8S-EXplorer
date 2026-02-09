"use client"

import React, { useEffect, useRef, useMemo } from "react"
import cytoscape from "cytoscape"
// @ts-expect-error fcose does not have types
import fcose from "cytoscape-fcose"
import { cn } from "@/lib/utils"

if (typeof window !== "undefined") {
    cytoscape.use(fcose)
}

const nodeStyles = {
    Deployment: { color: "#3b82f6", label: "D" },
    Pod: { color: "#22c55e", label: "P" },
    Service: { color: "#a855f7", label: "S" },
    PVC: { color: "#eab308", label: "V" },
    ConfigMap: { color: "#06b6d4", label: "C" },
    Secret: { color: "#ef4444", label: "K" },
    Ingress: { color: "#f97316", label: "I" },
    Namespace: { color: "#6366f1", label: "N" },
    Default: { color: "#6b7280", label: "R" },
}

const edgeStyles = {
    OWNS: { color: "#3b82f6", lineStyle: "solid" },
    EXPOSES: { color: "#a855f7", lineStyle: "dashed" },
    DEPENDS_ON: { color: "#f97316", lineStyle: "dotted" },
    MOUNTS: { color: "#22c55e", lineStyle: "solid" },
    DEFAULT: { color: "#94a3b8", lineStyle: "solid" },
}

export interface GraphNode {
    id: string
    type: string
    name: string
    namespace?: string
    status?: string
}

export interface GraphEdge {
    id: string
    source: string
    target: string
    type: string
}

interface GraphViewProps {
    nodes?: GraphNode[]
    edges?: GraphEdge[]
    selectedNodeId?: string | null
    onNodeSelect?: (node: GraphNode | null) => void
    layout?: string
    visibleNodeTypes?: string[]
    visibleEdgeTypes?: string[]
    searchQuery?: string
    selectedLabels?: string[]
    className?: string
}

export function GraphView({
    nodes = [],
    edges = [],
    selectedNodeId = null,
    onNodeSelect,
    layout = "force",
    visibleNodeTypes,
    visibleEdgeTypes,
    searchQuery = "",
    selectedLabels = [],
    className,
}: GraphViewProps) {
    const containerRef = useRef<HTMLDivElement>(null)
    const cyRef = useRef<cytoscape.Core | null>(null)

    // Filter nodes and edges based on visibility and search
    const filteredNodes = useMemo(() => {
        return nodes.filter((node) => {
            if (visibleNodeTypes && !visibleNodeTypes.includes(node.type.toLowerCase())) {
                return false
            }
            if (searchQuery && !node.name.toLowerCase().includes(searchQuery.toLowerCase())) {
                return false
            }
            if (selectedLabels && selectedLabels.length > 0) {
                const nodeLabels = (node as unknown as { properties?: { labels?: Record<string, string> } }).properties?.labels || {}
                const hasAllLabels = selectedLabels.every((label) => {
                    const [key, value] = label.split("=")
                    return nodeLabels[key] === value
                })
                if (!hasAllLabels) {
                    return false
                }
            }
            return true
        })
    }, [nodes, visibleNodeTypes, searchQuery, selectedLabels])

    const filteredEdges = useMemo(() => {
        return edges.filter((edge) => {
            if (visibleEdgeTypes && !visibleEdgeTypes.includes(edge.type.toLowerCase())) {
                return false
            }
            const sourceVisible = filteredNodes.some((n) => n.id === edge.source)
            const targetVisible = filteredNodes.some((n) => n.id === edge.target)
            return sourceVisible && targetVisible
        })
    }, [edges, visibleEdgeTypes, filteredNodes])

    const elements = useMemo(() => {
        const cyNodes = filteredNodes.map((node) => ({
            data: {
                ...node,
                id: node.id,
                label: node.name,
                type: node.type,
                status: node.status,
            },
        }))

        const cyEdges = filteredEdges.map((edge) => ({
            data: {
                id: edge.id,
                source: edge.source,
                target: edge.target,
                type: edge.type,
            },
        }))

        return [...cyNodes, ...cyEdges]
    }, [filteredNodes, filteredEdges])

    useEffect(() => {
        if (!containerRef.current) return

        const cy = cytoscape({
            container: containerRef.current,
            elements: elements,
            style: [
                {
                    selector: "node",
                    style: {
                        "background-color": (ele: cytoscape.NodeSingular) => {
                            const type = ele.data("type") as string
                            return (nodeStyles as Record<string, { color: string }>)[type]?.color || nodeStyles.Default.color
                        },
                        label: "data(label)",
                        "font-size": "10px",
                        "text-valign": "bottom",
                        "text-halign": "center",
                        "text-margin-y": 4,
                        width: 30,
                        height: 30,
                        color: "var(--foreground)",
                        "font-family": "var(--font-sans)",
                    },
                },
                {
                    selector: "node:selected",
                    style: {
                        "border-width": 3,
                        "border-color": "#3b82f6",
                        "border-opacity": 0.8,
                    },
                },
                {
                    selector: "edge",
                    style: {
                        width: 2,
                        "line-color": (ele: cytoscape.EdgeSingular) => {
                            const type = ele.data("type")?.toUpperCase()
                            return (edgeStyles as Record<string, { color: string }>)[type]?.color || edgeStyles.DEFAULT.color
                        },
                        "target-arrow-color": (ele: cytoscape.EdgeSingular) => {
                            const type = ele.data("type")?.toUpperCase()
                            return (edgeStyles as Record<string, { color: string }>)[type]?.color || edgeStyles.DEFAULT.color
                        },
                        "target-arrow-shape": "triangle",
                        "curve-style": "bezier",
                        "line-style": (ele: cytoscape.EdgeSingular) => {
                            const type = ele.data("type")?.toUpperCase()
                            return ((edgeStyles as Record<string, { lineStyle: string }>)[type]?.lineStyle || "solid") as "solid" | "dotted" | "dashed"
                        },
                        opacity: 0.6,
                    },
                },
            ],
            layout: {
                name: layout === "force" ? "fcose" : layout,
                animate: true,
                randomize: true,
                fit: true,
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
            } as cytoscape.LayoutOptions | any,
        })

        cy.on("tap", "node", (evt) => {
            const nodeData = evt.target.data()
            onNodeSelect?.(nodeData as GraphNode)
        })

        cy.on("tap", (evt) => {
            if (evt.target === cy) {
                onNodeSelect?.(null)
            }
        })

        cyRef.current = cy

        return () => {
            cy.destroy()
        }
    }, [elements, onNodeSelect, layout])

    useEffect(() => {
        if (cyRef.current) {
            cyRef.current.nodes().unselect()
            if (selectedNodeId) {
                cyRef.current.$id(selectedNodeId).select()
            }
        }
    }, [selectedNodeId])

    return (
        <div className={cn("relative h-full w-full bg-background/50", className)}>
            <div ref={containerRef} className="h-full w-full" />

            <div className="absolute bottom-4 left-4 p-3 bg-card/80 backdrop-blur rounded-lg border shadow-sm pointer-events-none">
                <div className="flex flex-wrap gap-x-4 gap-y-2">
                    {Object.entries(nodeStyles).map(([type, style]) => (
                        <div key={type} className="flex items-center gap-2">
                            <div
                                className="w-3 h-3 rounded-full"
                                style={{ backgroundColor: style.color }}
                            />
                            <span className="text-[10px] font-medium">{type}</span>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    )
}
