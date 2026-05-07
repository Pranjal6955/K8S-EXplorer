"use client"

import { useState, useEffect, useMemo } from "react"
import { useQuery } from "@apollo/client"
import { GET_TOPOLOGY } from "@/lib/queries"
import { ResourceSidebar } from "./resource-sidebar"
import { Topbar } from "./topbar"
import { GraphView, GraphNode, GraphEdge } from "./graph-view"
import { NodeDetails } from "./node-details"
import { useWebSocket } from "@/hooks/use-websocket"
import { toast } from "sonner"

interface Resource {
    id: string
    name: string
    type: string
    namespace: string
    status?: string
}

interface TopologyNode {
    id: string
    name: string
    type: string
    namespace?: string
    properties?: {
        status?: string
        labels?: Record<string, string>
        annotations?: Record<string, string>
        [key: string]: unknown
    }
}

interface TopologyEdge {
    id: string
    sourceId: string
    targetId: string
    type: string
}

interface TopologyData {
    getTopology: {
        nodes: TopologyNode[]
        edges: TopologyEdge[]
    }
}

interface DashboardNode extends GraphNode {
    properties?: TopologyNode["properties"]
}

export function DashboardLayout() {
    const [isDarkMode, setIsDarkMode] = useState(false)
    const [selectedNamespace, setSelectedNamespace] = useState<string | undefined>("default")
    const [searchQuery, setSearchQuery] = useState("")
    const [selectedNodeTypes, setSelectedNodeTypes] = useState<string[]>([
        "deployment", "pod", "service", "pvc", "configmap", "secret", "ingress"
    ])
    const [selectedEdgeTypes, setSelectedEdgeTypes] = useState<string[]>([
        "owns", "exposes", "depends_on", "mounts"
    ])
    const [layout, setLayout] = useState("fcose")
    const [selectedLabels, setSelectedLabels] = useState<string[]>([])
    const [selectedNode, setSelectedNode] = useState<DashboardNode | null>(null)

    // GraphQL Query for Topology
    const { data, loading, error, refetch } = useQuery<TopologyData>(GET_TOPOLOGY, {
        variables: {
            filter: {
                namespace: selectedNamespace || null,
            }
        }
    })

    // WebSocket integration
    const { isConnected } = useWebSocket((message) => {
        if (message.type === "TOPOLOGY_UPDATE" || message.type === "TOPOLOGY_UPDATE_ALL") {
            console.log("Received topology update, refreshing...")
            toast.info("Topology updated")
            refetch()
        } else if (message.type === "RESOURCE_EVENT") {
            // For now, also refetch on resource events to keep graph in sync
            // In the future, we can optimize this to update the cache directly
            console.log("Received resource event:", message.payload.type, message.payload.kind, message.payload.name)
            refetch()
        }
    })

    // Toggle theme
    useEffect(() => {
        document.documentElement.classList.toggle("dark", isDarkMode)
    }, [isDarkMode])

    // Data processing
    const nodes = useMemo<DashboardNode[]>(() => {
        if (!data?.getTopology?.nodes) return []
        return data.getTopology.nodes.map((n) => ({
            id: n.id,
            name: n.name,
            type: n.type,
            namespace: n.namespace || "default",
            status: n.properties?.status || "unknown",
            properties: n.properties
        }))
    }, [data])

    const edges = useMemo(() => {
        if (!data?.getTopology?.edges) return []
        return data.getTopology.edges.map((e) => ({
            id: e.id,
            source: e.sourceId,
            target: e.targetId,
            type: e.type
        }))
    }, [data])

    const namespaces = useMemo(() => {
        // Collect unique namespaces from nodes
        const nsSet = new Set<string>()
        nodes.forEach((n) => { if (n.namespace) nsSet.add(n.namespace) })
        return Array.from(nsSet)
    }, [nodes])

    const availableLabels = useMemo(() => {
        const labelsSet = new Set<string>()
        nodes.forEach((n) => {
            if (n.properties?.labels) {
                Object.entries(n.properties.labels).forEach(([key, value]) => {
                    labelsSet.add(`${key}=${value}`)
                })
            }
        })
        return Array.from(labelsSet).sort()
    }, [nodes])

    // Convert nodes to Resource format for sidebar
    const resources: Resource[] = useMemo(() => {
        return nodes
            .filter((n) => !selectedNamespace || n.namespace === selectedNamespace)
            .map((n) => ({
                id: n.id,
                name: n.name,
                type: n.type,
                namespace: n.namespace || "default",
                status: n.status,
            }))
    }, [nodes, selectedNamespace])

    // Get edges for selected node
    const selectedNodeEdges = useMemo(() => {
        if (!selectedNode) return { incoming: [], outgoing: [] }

        const incoming = edges
            .filter((e: GraphEdge) => e.target === selectedNode.id)
            .map((e: GraphEdge) => {
                const sourceNode = nodes.find((n: GraphNode) => n.id === e.source)
                return {
                    ...e,
                    sourceName: sourceNode?.name,
                    sourceType: sourceNode?.type,
                }
            })

        const outgoing = edges
            .filter((e: GraphEdge) => e.source === selectedNode.id)
            .map((e: GraphEdge) => {
                const targetNode = nodes.find((n: GraphNode) => n.id === e.target)
                return {
                    ...e,
                    targetName: targetNode?.name,
                    targetType: targetNode?.type,
                }
            })

        return { incoming, outgoing }
    }, [selectedNode, nodes, edges])

    // Handle resource selection from sidebar
    const handleResourceSelect = (resource: Resource) => {
        const node = nodes.find((n) => n.id === resource.id)
        if (node) {
            setSelectedNode(node)
        }
    }

    // Handle node navigation from details panel
    const handleNavigateToNode = (nodeId: string) => {
        const node = nodes.find((n) => n.id === nodeId)
        if (node) {
            setSelectedNode(node)
        }
    }

    return (
        <div className="flex h-screen w-full overflow-hidden bg-background">
            {/* Sidebar */}
            <ResourceSidebar
                resources={resources}
                selectedNamespace={selectedNamespace}
                namespaces={namespaces.length > 0 ? namespaces : ["default", "kube-system"]}
                onNamespaceChange={(ns) => setSelectedNamespace(ns || undefined)}
                onResourceSelect={handleResourceSelect}
                onSync={() => refetch()}
                isLoading={loading}
            />

            {/* Main content */}
            <div className="flex flex-1 flex-col overflow-hidden">
                {/* Topbar */}
                <Topbar
                    searchQuery={searchQuery}
                    onSearchChange={setSearchQuery}
                    selectedNodeTypes={selectedNodeTypes}
                    onNodeTypesChange={setSelectedNodeTypes}
                    selectedEdgeTypes={selectedEdgeTypes}
                    onEdgeTypesChange={setSelectedEdgeTypes}
                    availableLabels={availableLabels}
                    selectedLabels={selectedLabels}
                    onLabelsChange={setSelectedLabels}
                    layout={layout}
                    onLayoutChange={setLayout}
                    isDarkMode={isDarkMode}
                    onToggleTheme={() => setIsDarkMode(!isDarkMode)}

                    nodeCount={nodes.length}
                    edgeCount={edges.length}
                />

                {/* Connection status */}
                <div className="absolute top-2 right-2 z-50">
                    <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium border ${isConnected
                        ? "bg-green-50 text-green-700 border-green-200 dark:bg-green-900/20 dark:text-green-400 dark:border-green-800"
                        : "bg-red-50 text-red-700 border-red-200 dark:bg-red-900/20 dark:text-red-400 dark:border-red-800"
                        }`}>
                        <span className={`w-1.5 h-1.5 rounded-full ${isConnected ? "bg-green-600 dark:bg-green-400" : "bg-red-600 dark:bg-red-400"}`} />
                        {isConnected ? "Live" : "Disconnected"}
                    </span>
                </div>

                {/* Graph view */}
                <main className="flex-1 overflow-hidden relative">
                    {loading ? (
                        <div className="flex h-full w-full items-center justify-center">
                            <div className="flex flex-col items-center gap-2">
                                <svg
                                    className="h-8 w-8 animate-spin text-primary"
                                    xmlns="http://www.w3.org/2000/svg"
                                    fill="none"
                                    viewBox="0 0 24 24"
                                >
                                    <circle
                                        className="opacity-25"
                                        cx="12"
                                        cy="12"
                                        r="10"
                                        stroke="currentColor"
                                        strokeWidth="4"
                                    ></circle>
                                    <path
                                        className="opacity-75"
                                        fill="currentColor"
                                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                                    ></path>
                                </svg>
                                <span className="text-sm text-muted-foreground">Loading topology...</span>
                            </div>
                        </div>
                    ) : error ? (
                        <div className="flex h-full items-center justify-center p-6 text-center">
                            <div>
                                <p className="text-destructive font-medium mb-2">Error loading topology</p>
                                <p className="text-sm text-muted-foreground">{error.message}</p>
                                <button
                                    onClick={() => refetch()}
                                    className="mt-4 px-4 py-2 bg-primary text-primary-foreground rounded-md text-sm hover:bg-primary/90 transition-colors"
                                >
                                    Retry
                                </button>
                            </div>
                        </div>
                    ) : (
                        <GraphView
                            nodes={nodes}
                            edges={edges}
                            selectedNodeId={selectedNode?.id || null}
                            onNodeSelect={setSelectedNode}
                            layout={layout}
                            visibleNodeTypes={selectedNodeTypes}
                            visibleEdgeTypes={selectedEdgeTypes}
                            searchQuery={searchQuery}
                            selectedLabels={selectedLabels}
                        />
                    )}
                </main>
            </div>

            {/* Node details panel */}
            <NodeDetails
                node={selectedNode ? {
                    ...selectedNode,
                    labels: selectedNode.properties?.labels || {},
                    annotations: selectedNode.properties?.annotations || {},
                } : null}
                incomingEdges={selectedNodeEdges.incoming}
                outgoingEdges={selectedNodeEdges.outgoing}
                isOpen={selectedNode !== null}
                onClose={() => setSelectedNode(null)}
                onNavigateToNode={handleNavigateToNode}
            />
        </div>
    )
}
