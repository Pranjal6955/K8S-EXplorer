"use client"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Separator } from "@/components/ui/separator"
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
} from "@/components/ui/sheet"

// Icons
const icons = {
    x: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M18 6L6 18M6 6l12 12" />
        </svg>
    ),
    externalLink: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6M15 3h6v6M10 14L21 3" />
        </svg>
    ),
    copy: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <rect x="9" y="9" width="13" height="13" rx="2" />
            <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
        </svg>
    ),
    arrowRight: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M5 12h14M12 5l7 7-7 7" />
        </svg>
    ),
    arrowLeft: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M19 12H5M12 19l-7-7 7-7" />
        </svg>
    ),
}

const nodeColors: Record<string, { bg: string; text: string }> = {
    deployment: { bg: "bg-blue-500", text: "text-white" },
    pod: { bg: "bg-green-500", text: "text-white" },
    service: { bg: "bg-purple-500", text: "text-white" },
    pvc: { bg: "bg-yellow-500", text: "text-black" },
    configmap: { bg: "bg-cyan-500", text: "text-white" },
    secret: { bg: "bg-red-500", text: "text-white" },
    ingress: { bg: "bg-orange-500", text: "text-white" },
}

const statusColors: Record<string, string> = {
    running: "bg-green-500",
    pending: "bg-yellow-500",
    failed: "bg-red-500",
    succeeded: "bg-blue-500",
    unknown: "bg-gray-500",
}

interface NodeDetailsProps {
    node: {
        id: string
        type: string
        name: string
        namespace?: string
        status?: string
        labels?: Record<string, string>
        annotations?: Record<string, string>
        spec?: Record<string, unknown>
        metadata?: Record<string, unknown>
    } | null
    incomingEdges?: Array<{ id: string; source: string; type: string; sourceName?: string; sourceType?: string }>
    outgoingEdges?: Array<{ id: string; target: string; type: string; targetName?: string; targetType?: string }>
    isOpen?: boolean
    onClose?: () => void
    onNavigateToNode?: (nodeId: string) => void
}

export function NodeDetails({
    node,
    incomingEdges = [],
    outgoingEdges = [],
    isOpen = false,
    onClose,
    onNavigateToNode,
}: NodeDetailsProps) {
    if (!node) return null

    const colors = nodeColors[node.type.toLowerCase()] || { bg: "bg-gray-500", text: "text-white" }

    return (
        <Sheet open={isOpen} onOpenChange={(open) => !open && onClose?.()}>
            <SheetContent className="w-[400px] sm:w-[540px] p-0">
                <SheetHeader className="p-6 border-b">
                    <div className="flex items-start gap-3">
                        <div className={cn("p-3 rounded-lg", colors.bg, colors.text)}>
                            <span className="text-lg font-bold uppercase">{node.type.slice(0, 2)}</span>
                        </div>
                        <div className="flex-1 min-w-0">
                            <SheetTitle className="text-lg truncate">{node.name}</SheetTitle>
                            <SheetDescription className="flex items-center gap-2 mt-1">
                                <Badge variant="outline" className="capitalize">
                                    {node.type}
                                </Badge>
                                {node.namespace && (
                                    <Badge variant="secondary">{node.namespace}</Badge>
                                )}
                                {node.status && (
                                    <div className="flex items-center gap-1">
                                        <div className={cn("h-2 w-2 rounded-full", statusColors[node.status] || "bg-gray-500")} />
                                        <span className="text-xs capitalize">{node.status}</span>
                                    </div>
                                )}
                            </SheetDescription>
                        </div>
                    </div>
                </SheetHeader>

                <ScrollArea className="h-[calc(100vh-140px)]">
                    <div className="p-6 space-y-6">
                        {/* Basic Info */}
                        <section>
                            <h3 className="text-sm font-semibold mb-3 text-muted-foreground uppercase tracking-wider">
                                Basic Information
                            </h3>
                            <div className="grid gap-3">
                                <InfoRow label="UID" value={node.id} copyable />
                                <InfoRow label="Type" value={node.type} />
                                <InfoRow label="Namespace" value={node.namespace || "default"} />
                                {node.status && <InfoRow label="Status" value={node.status} />}
                            </div>
                        </section>

                        <Separator />

                        {/* Relationships */}
                        {(incomingEdges.length > 0 || outgoingEdges.length > 0) && (
                            <>
                                <section>
                                    <h3 className="text-sm font-semibold mb-3 text-muted-foreground uppercase tracking-wider">
                                        Relationships
                                    </h3>

                                    {incomingEdges.length > 0 && (
                                        <div className="mb-4">
                                            <div className="text-xs text-muted-foreground mb-2 flex items-center gap-1">
                                                {icons.arrowLeft}
                                                Incoming ({incomingEdges.length})
                                            </div>
                                            <div className="space-y-2">
                                                {incomingEdges.map((edge) => (
                                                    <button
                                                        key={edge.id}
                                                        onClick={() => onNavigateToNode?.(edge.source)}
                                                        className="w-full flex items-center gap-2 p-2 rounded-md border hover:bg-accent transition-colors text-left"
                                                    >
                                                        <Badge variant="secondary" className="text-xs shrink-0">
                                                            {edge.type}
                                                        </Badge>
                                                        <span className="text-sm truncate flex-1">
                                                            {edge.sourceName || edge.source}
                                                        </span>
                                                        {edge.sourceType && (
                                                            <Badge variant="outline" className="text-xs capitalize">
                                                                {edge.sourceType}
                                                            </Badge>
                                                        )}
                                                    </button>
                                                ))}
                                            </div>
                                        </div>
                                    )}

                                    {outgoingEdges.length > 0 && (
                                        <div>
                                            <div className="text-xs text-muted-foreground mb-2 flex items-center gap-1">
                                                {icons.arrowRight}
                                                Outgoing ({outgoingEdges.length})
                                            </div>
                                            <div className="space-y-2">
                                                {outgoingEdges.map((edge) => (
                                                    <button
                                                        key={edge.id}
                                                        onClick={() => onNavigateToNode?.(edge.target)}
                                                        className="w-full flex items-center gap-2 p-2 rounded-md border hover:bg-accent transition-colors text-left"
                                                    >
                                                        <Badge variant="secondary" className="text-xs shrink-0">
                                                            {edge.type}
                                                        </Badge>
                                                        <span className="text-sm truncate flex-1">
                                                            {edge.targetName || edge.target}
                                                        </span>
                                                        {edge.targetType && (
                                                            <Badge variant="outline" className="text-xs capitalize">
                                                                {edge.targetType}
                                                            </Badge>
                                                        )}
                                                    </button>
                                                ))}
                                            </div>
                                        </div>
                                    )}
                                </section>

                                <Separator />
                            </>
                        )}

                        {/* Labels */}
                        {node.labels && Object.keys(node.labels).length > 0 && (
                            <>
                                <section>
                                    <h3 className="text-sm font-semibold mb-3 text-muted-foreground uppercase tracking-wider">
                                        Labels
                                    </h3>
                                    <div className="flex flex-wrap gap-2">
                                        {Object.entries(node.labels).map(([key, value]) => (
                                            <Badge key={key} variant="outline" className="text-xs font-mono">
                                                {key}: {value}
                                            </Badge>
                                        ))}
                                    </div>
                                </section>

                                <Separator />
                            </>
                        )}

                        {/* Annotations */}
                        {node.annotations && Object.keys(node.annotations).length > 0 && (
                            <>
                                <section>
                                    <h3 className="text-sm font-semibold mb-3 text-muted-foreground uppercase tracking-wider">
                                        Annotations
                                    </h3>
                                    <div className="space-y-2">
                                        {Object.entries(node.annotations).map(([key, value]) => (
                                            <div key={key} className="text-xs">
                                                <div className="font-mono text-muted-foreground truncate">{key}</div>
                                                <div className="font-mono mt-0.5 truncate">{value}</div>
                                            </div>
                                        ))}
                                    </div>
                                </section>

                                <Separator />
                            </>
                        )}

                        {/* Actions */}
                        <section>
                            <h3 className="text-sm font-semibold mb-3 text-muted-foreground uppercase tracking-wider">
                                Actions
                            </h3>
                            <div className="flex flex-wrap gap-2">
                                <Button variant="outline" size="sm" className="gap-2">
                                    {icons.externalLink}
                                    View in K8s
                                </Button>
                                <Button variant="outline" size="sm" className="gap-2">
                                    View YAML
                                </Button>
                                <Button variant="outline" size="sm" className="gap-2">
                                    View Logs
                                </Button>
                                <Button variant="destructive" size="sm">
                                    Delete
                                </Button>
                            </div>
                        </section>
                    </div>
                </ScrollArea>
            </SheetContent>
        </Sheet>
    )
}

// Helper component for info rows
function InfoRow({
    label,
    value,
    copyable = false,
}: {
    label: string
    value: string
    copyable?: boolean
}) {
    const handleCopy = () => {
        navigator.clipboard.writeText(value)
    }

    return (
        <div className="flex items-center justify-between gap-4">
            <span className="text-sm text-muted-foreground">{label}</span>
            <div className="flex items-center gap-2">
                <span className="text-sm font-medium truncate max-w-[200px]">{value}</span>
                {copyable && (
                    <button
                        onClick={handleCopy}
                        className="p-1 hover:bg-accent rounded transition-colors text-muted-foreground hover:text-foreground"
                    >
                        {icons.copy}
                    </button>
                )}
            </div>
        </div>
    )
}
