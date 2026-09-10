import { useState, useEffect, useRef, useCallback } from 'react'
import type { Device, DevicePlacement } from '../types'
import { apiFetch } from '../api'

interface DeviceStreamMessage {
  type?: string
  device?: Device
  entity_id?: string
  error?: unknown
}

/**
 * Owns the live device-state stream: an initial REST fetch plus a self-healing
 * WebSocket that keeps the device map in sync with Home Assistant, and exposes
 * the raw `send`/`refresh` primitives for consumers that manage their own
 * actions (e.g. read-only views).
 */
export function useRealtimeDeviceMap(onMessage?: (message: DeviceStreamMessage) => void) {
  const [deviceMap, setDeviceMap] = useState<Record<string, Device>>({})
  const [connected, setConnected] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const onMessageRef = useRef(onMessage)
  useEffect(() => {
    onMessageRef.current = onMessage
  }, [onMessage])

  const fetchDevices = useCallback(async () => {
    try {
      const res = await fetch('/api/devices')
      if (res.ok) {
        const payload = await res.json()
        if (payload?.status === 'success' && Array.isArray(payload.data)) {
          const map: Record<string, Device> = {}
          for (const d of payload.data as Device[]) {
            map[d.id] = d
          }
          setDeviceMap(map)
        }
      }
    } catch (err) {
      console.error('Failed to fetch devices:', err)
    }
  }, [])

  const send = useCallback((payload: unknown): boolean => {
    const ws = wsRef.current
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(payload))
      return true
    }
    return false
  }, [])

  useEffect(() => {
    fetchDevices()

    let shouldReconnect = true

    function connectWS() {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/api/ws`

      const ws = new WebSocket(wsUrl)
      wsRef.current = ws

      ws.onopen = () => {
        setConnected(true)
      }

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as DeviceStreamMessage

          // Real-time confirmed state update from Home Assistant
          if (data.type === 'state_changed' && data.device) {
            const dev = data.device
            setDeviceMap((prev) => ({
              ...prev,
              [dev.id]: dev,
            }))
            onMessageRef.current?.(data)
          } else if (data.type === 'action_success' && data.entity_id) {
            // Acknowledged action: refetch fresh state, then let the consumer settle its pending state
            setTimeout(() => {
              fetchDevices().finally(() => onMessageRef.current?.(data))
            }, 600)
          } else if (data.type === 'error' && data.entity_id) {
            onMessageRef.current?.(data)
            fetchDevices()
          }
        } catch (e) {
          console.error('Failed to parse WS message:', e)
        }
      }

      ws.onclose = () => {
        setConnected(false)
        wsRef.current = null
        if (shouldReconnect) {
          reconnectTimerRef.current = setTimeout(connectWS, 2500)
        }
      }

      ws.onerror = () => {
        ws.close()
      }
    }

    connectWS()

    return () => {
      shouldReconnect = false
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current)
      }
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [fetchDevices])

  return { deviceMap, connected, send, refreshDevices: fetchDevices }
}

/**
 * Live device state plus allow-listed actuator toggling with pending
 * bookkeeping. Shared by the floor view and the house overview.
 */
export function useRealtimeDeviceControl() {
  const [pendingDevices, setPendingDevices] = useState<Record<string, boolean>>({})
  const pendingTimersRef = useRef<Record<string, ReturnType<typeof setTimeout>>>({})

  // Clear pending state for an entity
  const clearPending = useCallback((entityId: string) => {
    if (pendingTimersRef.current[entityId]) {
      clearTimeout(pendingTimersRef.current[entityId])
      delete pendingTimersRef.current[entityId]
    }
    setPendingDevices((prev) => {
      if (!prev[entityId]) return prev
      const copy = { ...prev }
      delete copy[entityId]
      return copy
    })
  }, [])

  const handleStreamMessage = useCallback(
    (data: DeviceStreamMessage) => {
      if (data.type === 'state_changed' && data.device) {
        clearPending(data.device.id)
      } else if (data.type === 'action_success' && data.entity_id) {
        clearPending(data.entity_id)
      } else if (data.type === 'error' && data.entity_id) {
        console.warn(`Home Assistant action failed for ${data.entity_id}:`, data.error)
        clearPending(data.entity_id)
      }
    },
    [clearPending]
  )

  const { deviceMap, connected, send, refreshDevices } = useRealtimeDeviceMap(handleStreamMessage)

  useEffect(() => {
    return () => {
      for (const t of Object.values(pendingTimersRef.current)) {
        clearTimeout(t)
      }
    }
  }, [])

  // Execute action without optimistic state toggle, tracking intermediate pending state
  const toggleDevice = useCallback(
    async (entityId: string) => {
      // Prevent multiple concurrent actions on the same device
      if (pendingDevices[entityId]) {
        return
      }

      // Mark device as pending (intermediate processing state)
      setPendingDevices((prev) => ({ ...prev, [entityId]: true }))

      // Safety timeout: reset pending state after 15s if HA or network fails to respond
      if (pendingTimersRef.current[entityId]) {
        clearTimeout(pendingTimersRef.current[entityId])
      }
      pendingTimersRef.current[entityId] = setTimeout(() => {
        clearPending(entityId)
        refreshDevices()
      }, 15000)

      // 1. Try WebSocket send first
      if (send({ type: 'action', entity_id: entityId, action: 'toggle' })) {
        return
      }

      // 2. REST fallback
      try {
        const res = await apiFetch('/api/actions', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            entity_id: entityId,
            action: 'toggle',
          }),
        })
        if (!res.ok) {
          console.warn('REST action failed, resetting pending state')
          clearPending(entityId)
          refreshDevices()
        } else {
          // Re-fetch confirmed state from server
          await refreshDevices()
          clearPending(entityId)
        }
      } catch (err) {
        console.error('REST action failed:', err)
        clearPending(entityId)
        refreshDevices()
      }
    },
    [pendingDevices, clearPending, refreshDevices, send]
  )

  return { deviceMap, connected, pendingDevices, toggleDevice, refreshDevices }
}

export function useRealtimeDevices(levelId: string | null) {
  const { deviceMap, connected, pendingDevices, toggleDevice, refreshDevices } = useRealtimeDeviceControl()
  const [placements, setPlacements] = useState<DevicePlacement[]>([])
  const [loading, setLoading] = useState(false)

  // Fetch placements whenever active level changes
  const fetchPlacements = useCallback(async (lvlId: string) => {
    setLoading(true)
    try {
      const res = await fetch(`/api/levels/${lvlId}/placements`)
      if (res.ok) {
        const payload = await res.json()
        if (payload?.status === 'success' && Array.isArray(payload.data)) {
          setPlacements(payload.data)
        } else {
          setPlacements([])
        }
      } else {
        setPlacements([])
      }
    } catch (err) {
      console.error('Failed to fetch placements:', err)
      setPlacements([])
    } finally {
      setLoading(false)
    }
  }, [])

  // Reload placements when levelId changes
  useEffect(() => {
    if (levelId) {
      fetchPlacements(levelId)
    } else {
      setPlacements([])
    }
  }, [levelId, fetchPlacements])

  return {
    deviceMap,
    devices: Object.values(deviceMap),
    placements,
    pendingDevices,
    loading,
    connected,
    toggleDevice,
    refreshPlacements: () => {
      if (levelId) void fetchPlacements(levelId)
    },
    refreshDevices,
  }
}
