import { useState, useEffect, useRef, useCallback } from 'react'
import type { Device, DevicePlacement } from '../types'
import { apiFetch } from '../api'

export function useRealtimeDevices(levelId: string | null) {
  const [deviceMap, setDeviceMap] = useState<Record<string, Device>>({})
  const [placements, setPlacements] = useState<DevicePlacement[]>([])
  const [pendingDevices, setPendingDevices] = useState<Record<string, boolean>>({})
  const [loading, setLoading] = useState(false)
  const [connected, setConnected] = useState(false)

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const pendingTimersRef = useRef<Record<string, ReturnType<typeof setTimeout>>>({})

  // Fetch initial devices list
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

  // Setup WebSocket connection with auto-reconnection
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
          const data = JSON.parse(event.data)

          // Real-time confirmed state update from Home Assistant
          if (data.type === 'state_changed' && data.device) {
            const dev = data.device as Device
            setDeviceMap((prev) => ({
              ...prev,
              [dev.id]: dev,
            }))
            clearPending(dev.id)
          } else if (data.type === 'action_success' && data.entity_id) {
            // Action acknowledged by Home Assistant; fetch fresh state if not received via broadcast
            const entityId = data.entity_id as string
            setTimeout(() => {
              fetchDevices().finally(() => clearPending(entityId))
            }, 600)
          } else if (data.type === 'error' && data.entity_id) {
            console.warn(`Home Assistant action failed for ${data.entity_id}:`, data.error)
            clearPending(data.entity_id as string)
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
      for (const t of Object.values(pendingTimersRef.current)) {
        clearTimeout(t)
      }
    }
  }, [fetchDevices, clearPending])

  // Reload placements when levelId changes
  useEffect(() => {
    if (levelId) {
      fetchPlacements(levelId)
    } else {
      setPlacements([])
    }
  }, [levelId, fetchPlacements])

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
        fetchDevices()
      }, 15000)

      // 1. Try WebSocket send first
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.send(
          JSON.stringify({
            type: 'action',
            entity_id: entityId,
            action: 'toggle',
          })
        )
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
          fetchDevices()
        } else {
          // Re-fetch confirmed state from server
          await fetchDevices()
          clearPending(entityId)
        }
      } catch (err) {
        console.error('REST action failed:', err)
        clearPending(entityId)
        fetchDevices()
      }
    },
    [fetchDevices, pendingDevices, clearPending]
  )

  return {
    deviceMap,
    devices: Object.values(deviceMap),
    placements,
    pendingDevices,
    loading,
    connected,
    toggleDevice,
    refreshPlacements: () => levelId && fetchPlacements(levelId),
    refreshDevices: fetchDevices,
  }
}
