"use client"

import { useState } from "react"
import { cn } from "@/lib/utils"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
    DropdownMenuCheckboxItem,
} from "@/components/ui/dropdown-menu"
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "@/components/ui/tooltip"

// Icons
const icons = {
    search: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="11" cy="11" r="8" />
            <path d="M21 21l-4.35-4.35" />
        </svg>
    ),
    filter: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" />
        </svg>
    ),
    zoomIn: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="11" cy="11" r="8" />
            <path d="M21 21l-4.35-4.35" />
            <path d="M11 8v6M8 11h6" />
        </svg>
    ),
    zoomOut: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="11" cy="11" r="8" />
            <path d="M21 21l-4.35-4.35" />
            <path d="M8 11h6" />
        </svg>
    ),
    resetView: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M15 3h6v6M14 10l6.1-6.1M9 21H3v-6M10 14l-6.1 6.1" />
        </svg>
    ),
    layout: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <rect x="3" y="3" width="18" height="18" rx="2" />
            <path d="M3 9h18M9 21V9" />
        </svg>
    ),
    sun: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="12" cy="12" r="4" />
            <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41" />
        </svg>
    ),
    moon: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z" />
        </svg>
    ),
    settings: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="12" cy="12" r="3" />
            <path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-2 2 2 2 0 01-2-2v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83 0 2 2 0 010-2.83l.06-.06a1.65 1.65 0 00.33-1.82 1.65 1.65 0 00-1.51-1H3a2 2 0 01-2-2 2 2 0 012-2h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 010-2.83 2 2 0 012.83 0l.06.06a1.65 1.65 0 001.82.33H9a1.65 1.65 0 001-1.51V3a2 2 0 012-2 2 2 0 012 2v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 0 2 2 0 010 2.83l-.06.06a1.65 1.65 0 00-.33 1.82V9a1.65 1.65 0 001.51 1H21a2 2 0 012 2 2 2 0 01-2 2h-.09a1.65 1.65 0 00-1.51 1z" />
        </svg>
    ),
    user: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2" />
            <circle cx="12" cy="7" r="4" />
        </svg>
    ),
    x: (
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M18 6L6 18M6 6l12 12" />
        </svg>
    ),
}

const nodeTypes = [
    { id: "deployment", label: "Deployments", color: "bg-blue-500" },
    { id: "pod", label: "Pods", color: "bg-green-500" },
    { id: "service", label: "Services", color: "bg-purple-500" },
    { id: "pvc", label: "PVCs", color: "bg-yellow-500" },
    { id: "configmap", label: "ConfigMaps", color: "bg-cyan-500" },
    { id: "secret", label: "Secrets", color: "bg-red-500" },
    { id: "ingress", label: "Ingresses", color: "bg-orange-500" },
]

const edgeTypes = [
    { id: "owns", label: "Owns" },
    { id: "exposes", label: "Exposes" },
    { id: "depends_on", label: "Depends On" },
    { id: "mounts", label: "Mounts" },
]

const layoutOptions = [
    { id: "force", label: "Force-directed" },
    { id: "hierarchical", label: "Hierarchical" },
    { id: "circular", label: "Circular" },
    { id: "grid", label: "Grid" },
]

interface TopbarProps {
    searchQuery?: string
    onSearchChange?: (query: string) => void
    selectedNodeTypes?: string[]
    onNodeTypesChange?: (types: string[]) => void
    selectedEdgeTypes?: string[]
    onEdgeTypesChange?: (types: string[]) => void
    selectedLabels?: string[]
    onLabelsChange?: (labels: string[]) => void
    availableLabels?: string[]
    layout?: string
    onLayoutChange?: (layout: string) => void
    onZoomIn?: () => void
    onZoomOut?: () => void
    onResetView?: () => void
    isDarkMode?: boolean
    onToggleTheme?: () => void
    nodeCount?: number
    edgeCount?: number
}

export function Topbar({
    searchQuery = "",
    onSearchChange,
    selectedNodeTypes = nodeTypes.map((n) => n.id),
    onNodeTypesChange,
    selectedEdgeTypes = edgeTypes.map((e) => e.id),
    onEdgeTypesChange,
    selectedLabels = [],
    onLabelsChange,
    availableLabels = [],
    layout = "force",
    onLayoutChange,
    onZoomIn,
    onZoomOut,
    onResetView,
    isDarkMode = false,
    onToggleTheme,
    nodeCount = 0,
    edgeCount = 0,
}: TopbarProps) {
    const [isSearchFocused, setIsSearchFocused] = useState(false)

    const toggleNodeType = (typeId: string) => {
        if (selectedNodeTypes.includes(typeId)) {
            onNodeTypesChange?.(selectedNodeTypes.filter((t) => t !== typeId))
        } else {
            onNodeTypesChange?.([...selectedNodeTypes, typeId])
        }
    }

    const toggleEdgeType = (typeId: string) => {
        if (selectedEdgeTypes.includes(typeId)) {
            onEdgeTypesChange?.(selectedEdgeTypes.filter((t) => t !== typeId))
        } else {
            onEdgeTypesChange?.([...selectedEdgeTypes, typeId])
        }
    }

    const toggleLabel = (label: string) => {
        if (selectedLabels.includes(label)) {
            onLabelsChange?.(selectedLabels.filter((l) => l !== label))
        } else {
            onLabelsChange?.([...selectedLabels, label])
        }
    }

    return (
        <TooltipProvider>
            <div className="flex h-14 items-center justify-between border-b bg-card px-4">
                {/* Left section: Search */}
                <div className="flex items-center gap-4 flex-1 max-w-xl">
                    <div className={cn(
                        "relative flex items-center transition-all duration-200",
                        isSearchFocused ? "flex-1" : "w-64"
                    )}>
                        <div className="absolute left-3 text-muted-foreground">
                            {icons.search}
                        </div>
                        <Input
                            type="text"
                            placeholder="Search resources..."
                            value={searchQuery}
                            onChange={(e) => onSearchChange?.(e.target.value)}
                            onFocus={() => setIsSearchFocused(true)}
                            onBlur={() => setIsSearchFocused(false)}
                            className="pl-10 pr-10 bg-background"
                        />
                        {searchQuery && (
                            <button
                                onClick={() => onSearchChange?.("")}
                                className="absolute right-3 text-muted-foreground hover:text-foreground"
                            >
                                {icons.x}
                            </button>
                        )}
                    </div>
                </div>

                {/* Center section: Filters */}
                <div className="flex items-center gap-2">
                    {/* Node Type Filter */}
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button variant="outline" size="sm" className="gap-2">
                                {icons.filter}
                                <span className="hidden sm:inline">Nodes</span>
                                <Badge variant="secondary" className="h-5 px-1.5">
                                    {selectedNodeTypes.length}
                                </Badge>
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="center" className="w-48">
                            <DropdownMenuLabel>Show Node Types</DropdownMenuLabel>
                            <DropdownMenuSeparator />
                            {nodeTypes.map((type) => (
                                <DropdownMenuCheckboxItem
                                    key={type.id}
                                    checked={selectedNodeTypes.includes(type.id)}
                                    onCheckedChange={() => toggleNodeType(type.id)}
                                >
                                    <div className="flex items-center gap-2">
                                        <div className={cn("h-2 w-2 rounded-full", type.color)} />
                                        {type.label}
                                    </div>
                                </DropdownMenuCheckboxItem>
                            ))}
                            <DropdownMenuSeparator />
                            <DropdownMenuItem onClick={() => onNodeTypesChange?.(nodeTypes.map((n) => n.id))}>
                                Select All
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => onNodeTypesChange?.([])}>
                                Clear All
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>

                    {/* Edge Type Filter */}
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button variant="outline" size="sm" className="gap-2">
                                <span className="hidden sm:inline">Edges</span>
                                <Badge variant="secondary" className="h-5 px-1.5">
                                    {selectedEdgeTypes.length}
                                </Badge>
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="center" className="w-48">
                            <DropdownMenuLabel>Show Edge Types</DropdownMenuLabel>
                            <DropdownMenuSeparator />
                            {edgeTypes.map((type) => (
                                <DropdownMenuCheckboxItem
                                    key={type.id}
                                    checked={selectedEdgeTypes.includes(type.id)}
                                    onCheckedChange={() => toggleEdgeType(type.id)}
                                >
                                    {type.label}
                                </DropdownMenuCheckboxItem>
                            ))}
                        </DropdownMenuContent>
                    </DropdownMenu>

                    {/* Label Filter */}
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button variant="outline" size="sm" className="gap-2">
                                <span className="hidden sm:inline">Labels</span>
                                <Badge variant="secondary" className="h-5 px-1.5">
                                    {selectedLabels.length}
                                </Badge>
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="center" className="w-64 max-h-[300px] overflow-y-auto">
                            <DropdownMenuLabel>Filter by Labels</DropdownMenuLabel>
                            <DropdownMenuSeparator />
                            {availableLabels.length > 0 ? (
                                availableLabels.map((label) => (
                                    <DropdownMenuCheckboxItem
                                        key={label}
                                        checked={selectedLabels.includes(label)}
                                        onCheckedChange={() => toggleLabel(label)}
                                    >
                                        <span className="truncate" title={label}>{label}</span>
                                    </DropdownMenuCheckboxItem>
                                ))
                            ) : (
                                <div className="p-2 text-xs text-muted-foreground text-center">
                                    No labels found
                                </div>
                            )}
                            <DropdownMenuSeparator />
                            <DropdownMenuItem onClick={() => onLabelsChange?.([])}>
                                Clear All
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>

                    {/* Layout Selector */}
                    <Select value={layout} onValueChange={onLayoutChange}>
                        <SelectTrigger className="w-[140px] h-9">
                            <div className="flex items-center gap-2">
                                {icons.layout}
                                <SelectValue />
                            </div>
                        </SelectTrigger>
                        <SelectContent>
                            {layoutOptions.map((option) => (
                                <SelectItem key={option.id} value={option.id}>
                                    {option.label}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>

                    {/* Zoom Controls */}
                    <div className="flex items-center gap-1 border rounded-md">
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onZoomOut}>
                                    {icons.zoomOut}
                                </Button>
                            </TooltipTrigger>
                            <TooltipContent>Zoom Out</TooltipContent>
                        </Tooltip>
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onResetView}>
                                    {icons.resetView}
                                </Button>
                            </TooltipTrigger>
                            <TooltipContent>Reset View</TooltipContent>
                        </Tooltip>
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onZoomIn}>
                                    {icons.zoomIn}
                                </Button>
                            </TooltipTrigger>
                            <TooltipContent>Zoom In</TooltipContent>
                        </Tooltip>
                    </div>
                </div>

                {/* Right section: Stats and Actions */}
                <div className="flex items-center gap-4">
                    {/* Stats */}
                    <div className="hidden md:flex items-center gap-3 text-sm text-muted-foreground">
                        <div className="flex items-center gap-1">
                            <span className="font-medium text-foreground">{nodeCount}</span>
                            <span>nodes</span>
                        </div>
                        <div className="h-4 w-px bg-border" />
                        <div className="flex items-center gap-1">
                            <span className="font-medium text-foreground">{edgeCount}</span>
                            <span>edges</span>
                        </div>
                    </div>

                    {/* Theme Toggle */}
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button variant="ghost" size="icon" onClick={onToggleTheme}>
                                {isDarkMode ? icons.sun : icons.moon}
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent>
                            {isDarkMode ? "Light Mode" : "Dark Mode"}
                        </TooltipContent>
                    </Tooltip>

                    {/* Settings */}
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button variant="ghost" size="icon">
                                {icons.settings}
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent>Settings</TooltipContent>
                    </Tooltip>

                    {/* User Menu */}
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" className="rounded-full">
                                <div className="h-8 w-8 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
                                    <span className="text-xs font-bold text-white">U</span>
                                </div>
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                            <DropdownMenuLabel>My Account</DropdownMenuLabel>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem>Profile</DropdownMenuItem>
                            <DropdownMenuItem>Settings</DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem>Log out</DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                </div>
            </div>
        </TooltipProvider>
    )
}
