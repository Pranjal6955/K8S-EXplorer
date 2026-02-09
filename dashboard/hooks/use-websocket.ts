"use client"

import { useEffect, useRef, useState, useCallback } from "react"

// Use environment variable or default to localhost:8080
const WEBSOCKET_URL = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/ws"
const RECONNECT_DELAY = 3000

export interface WebSocketMessage {
    type: "TOPOLOGY_UPDATE" | "TOPOLOGY_UPDATE_ALL" | "RESOURCE_EVENT"
    payload: any
}

export function useWebSocket(onMessage: (message: WebSocketMessage) => void) {
    const ws = useRef<WebSocket | null>(null)
    const [isConnected, setIsConnected] = useState(false)
    const onMessageRef = useRef(onMessage)

    // Update ref when callback changes
    useEffect(() => {
        onMessageRef.current = onMessage
    }, [onMessage])

    const connect = useCallback(() => {
        if (ws.current?.readyState === WebSocket.OPEN || ws.current?.readyState === WebSocket.CONNECTING) return

        try {
            console.log("Connecting to WebSocket:", WEBSOCKET_URL)
            const socket = new WebSocket(WEBSOCKET_URL)
            ws.current = socket

            socket.onopen = () => {
                console.log("Connected to WebSocket")
                setIsConnected(true)
            }

            socket.onmessage = (event) => {
                try {
                    const message: WebSocketMessage = JSON.parse(event.data)
                    if (onMessageRef.current) {
                        onMessageRef.current(message)
                    }
                } catch (error) {
                    console.error("Failed to parse WebSocket message:", error)
                }
            }

            socket.onclose = () => {
                console.log("WebSocket disconnected")
                setIsConnected(false)
                ws.current = null
                // Attempt to reconnect
                setTimeout(connect, RECONNECT_DELAY)
            }

            socket.onerror = (error) => {
                console.error("WebSocket error:", error)
                socket.close()
            }
        } catch (error) {
            console.error("Failed to connect:", error)
            setTimeout(connect, RECONNECT_DELAY)
        }
    }, [])

    useEffect(() => {
        connect()
        return () => {
            if (ws.current) {
                // Prevent reconnect on unmount
                ws.current.onclose = null
                ws.current.close()
            }
        }
    }, [connect])

    return { isConnected }
}
