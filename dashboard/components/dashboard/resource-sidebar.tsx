"use client"

import { useState } from "react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Separator } from "@/components/ui/separator"
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "@/components/ui/tooltip"

// Icons using SVG
const icons = {
    pod: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <rect x="3" y="3" width="18" height="18" rx="2" />
            <circle cx="12" cy="12" r="3" />
        </svg>
    ),
    deployment: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <rect x="2" y="3" width="20" height="6" rx="1" />
            <rect x="2" y="12" width="20" height="9" rx="1" />
            <line x1="6" y1="6" x2="6" y2="6.01" />
            <line x1="10" y1="6" x2="10" y2="6.01" />
        </svg>
    ),
    service: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="8" x2="12" y2="16" />
            <line x1="8" y1="12" x2="16" y2="12" />
        </svg>
    ),
    pvc: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M12 2L4 8v8l8 6 8-6V8l-8-6z" />
            <line x1="12" y1="22" x2="12" y2="11" />
            <path d="M20 8l-8 5-8-5" />
        </svg>
    ),
    configmap: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" />
            <polyline points="14,2 14,8 20,8" />
            <line x1="8" y1="13" x2="16" y2="13" />
            <line x1="8" y1="17" x2="16" y2="17" />
        </svg>
    ),
    secret: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <rect x="3" y="11" width="18" height="11" rx="2" />
            <path d="M7 11V7a5 5 0 0110 0v4" />
            <circle cx="12" cy="16" r="1" />
        </svg>
    ),
    ingress: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M4 4h16v16H4z" />
            <path d="M9 4v16" />
            <path d="M4 12h16" />
        </svg>
    ),
    namespace: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z" />
        </svg>
    ),
    chevronDown: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polyline points="6,9 12,15 18,9" />
        </svg>
    ),
    chevronRight: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polyline points="9,6 15,12 9,18" />
        </svg>
    ),
    sync: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M23 4v6h-6" />
            <path d="M1 20v-6h6" />
            <path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15" />
        </svg>
    ),
}

// Resource type configuration
const resourceTypes = [
    { id: "deployment", label: "Deployments", icon: icons.deployment, color: "bg-blue-500" },
    { id: "pod", label: "Pods", icon: icons.pod, color: "bg-green-500" },
    { id: "service", label: "Services", icon: icons.service, color: "bg-purple-500" },
    { id: "pvc", label: "PVCs", icon: icons.pvc, color: "bg-yellow-500" },
    { id: "configmap", label: "ConfigMaps", icon: icons.configmap, color: "bg-cyan-500" },
    { id: "secret", label: "Secrets", icon: icons.secret, color: "bg-red-500" },
    { id: "ingress", label: "Ingresses", icon: icons.ingress, color: "bg-orange-500" },
]

interface Resource {
    id: string
    name: string
    type: string
    namespace: string
    status?: string
}

interface ResourceSidebarProps {
    resources?: Resource[]
    selectedNamespace?: string
    namespaces?: string[]
    onNamespaceChange?: (namespace: string) => void
    onResourceSelect?: (resource: Resource) => void
    onSync?: () => void
    isLoading?: boolean
    collapsed?: boolean
}

export function ResourceSidebar({
    resources = [],
    selectedNamespace = "default",
    namespaces = ["default", "kube-system", "monitoring"],
    onNamespaceChange,
    onResourceSelect,
    onSync,
    isLoading = false,
    collapsed = false,
}: ResourceSidebarProps) {
    const [expandedTypes, setExpandedTypes] = useState<string[]>(["deployment", "pod", "service"])

    const toggleType = (type: string) => {
        setExpandedTypes((prev) =>
            prev.includes(type) ? prev.filter((t) => t !== type) : [...prev, type]
        )
    }

    const getResourcesByType = (type: string) => {
        return resources.filter((r) => r.type.toLowerCase() === type)
    }

    const getStatusColor = (status?: string) => {
        switch (status) {
            case "running":
                return "bg-green-500"
            case "pending":
                return "bg-yellow-500"
            case "failed":
                return "bg-red-500"
            default:
                return "bg-gray-500"
        }
    }

    if (collapsed) {
        return (
            <div className="flex h-full w-16 flex-col border-r bg-card">
                <div className="flex h-14 items-center justify-center border-b">
                    <div className="h-8 w-8 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600" />
                </div>
                <ScrollArea className="flex-1">
                    <div className="flex flex-col items-center gap-2 p-2">
                        <TooltipProvider>
                            {resourceTypes.map((type) => (
                                <Tooltip key={type.id}>
                                    <TooltipTrigger asChild>
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className={cn(
                                                "h-10 w-10",
                                                expandedTypes.includes(type.id) && "bg-accent"
                                            )}
                                            onClick={() => toggleType(type.id)}
                                        >
                                            <div className={cn("p-1 rounded", type.color, "text-white")}>
                                                {type.icon}
                                            </div>
                                        </Button>
                                    </TooltipTrigger>
                                    <TooltipContent side="right">
                                        <p>{type.label} ({getResourcesByType(type.id).length})</p>
                                    </TooltipContent>
                                </Tooltip>
                            ))}
                        </TooltipProvider>
                    </div>
                </ScrollArea>
            </div>
        )
    }

    return (
        <div className="flex h-full w-64 flex-col border-r bg-card">
            {/* Header */}
            <div className="flex h-14 items-center justify-between border-b px-4">
                <div className="flex items-center gap-2">
                    <div className="h-8 w-8 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
                        <span className="text-xs font-bold text-white">K8S</span>
                    </div>
                    <span className="font-semibold text-sm">Resources</span>
                </div>
                <Button
                    variant="ghost"
                    size="icon"
                    onClick={onSync}
                    disabled={isLoading}
                    className={cn(isLoading && "animate-spin")}
                >
                    {icons.sync}
                </Button>
            </div>

            {/* Namespace selector */}
            <div className="border-b p-3">
                <div className="flex items-center gap-2 text-xs text-muted-foreground mb-2">
                    {icons.namespace}
                    <span>Namespace</span>
                </div>
                <select
                    value={selectedNamespace}
                    onChange={(e) => onNamespaceChange?.(e.target.value)}
                    className="w-full rounded-md border bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                >
                    <option value="">All Namespaces</option>
                    {namespaces.map((ns) => (
                        <option key={ns} value={ns}>
                            {ns}
                        </option>
                    ))}
                </select>
            </div>

            {/* Resource list */}
            <ScrollArea className="flex-1">
                <div className="p-2">
                    {resourceTypes.map((type) => {
                        const typeResources = getResourcesByType(type.id)
                        const isExpanded = expandedTypes.includes(type.id)

                        return (
                            <div key={type.id} className="mb-1">
                                <button
                                    onClick={() => toggleType(type.id)}
                                    className="flex w-full items-center justify-between rounded-md px-2 py-2 text-sm hover:bg-accent transition-colors"
                                >
                                    <div className="flex items-center gap-2">
                                        <div className={cn("p-1 rounded", type.color, "text-white")}>
                                            {type.icon}
                                        </div>
                                        <span className="font-medium">{type.label}</span>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <Badge variant="secondary" className="h-5 min-w-[1.5rem] px-1 text-xs">
                                            {typeResources.length}
                                        </Badge>
                                        {isExpanded ? icons.chevronDown : icons.chevronRight}
                                    </div>
                                </button>

                                {isExpanded && typeResources.length > 0 && (
                                    <div className="ml-4 mt-1 space-y-0.5">
                                        {typeResources.map((resource) => (
                                            <button
                                                key={resource.id}
                                                onClick={() => onResourceSelect?.(resource)}
                                                className="flex w-full items-center gap-2 rounded-md px-3 py-1.5 text-xs hover:bg-accent transition-colors text-left"
                                            >
                                                <div
                                                    className={cn(
                                                        "h-2 w-2 rounded-full",
                                                        getStatusColor(resource.status)
                                                    )}
                                                />
                                                <span className="truncate flex-1">{resource.name}</span>
                                            </button>
                                        ))}
                                    </div>
                                )}

                                {isExpanded && typeResources.length === 0 && (
                                    <div className="ml-4 mt-1 px-3 py-2 text-xs text-muted-foreground">
                                        No {type.label.toLowerCase()} found
                                    </div>
                                )}
                            </div>
                        )
                    })}
                </div>
            </ScrollArea>

            {/* Footer */}
            <Separator />
            <div className="p-3">
                <div className="flex items-center justify-between text-xs text-muted-foreground">
                    <span>Total: {resources.length} resources</span>
                    <Badge variant="outline" className="text-xs">
                        {selectedNamespace || "All"}
                    </Badge>
                </div>
            </div>
        </div>
    )
}
