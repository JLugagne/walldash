import { useState, useEffect, useRef, useCallback } from 'react'
import type { Device, DevicePlacement } from '../types'

export function useRealtimeDevices(levelId: string | null) {
  const [deviceMap, setDeviceMap] = useState<Record<string, Device>>({})
  const [placements, setPlacements] = useState<DevicePlacement[]>([])
  const [loading, setLoading] = useState(false)
  const [connected, setConnected] = useState(false)

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

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
          if (data.type === 'state_changed' && data.device) {
            const dev = data.device as Device
            setDeviceMap((prev) => ({
              ...prev,
              [dev.id]: dev,
            }))
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

  // Reload placements when levelId changes
  useEffect(() => {
    if (levelId) {
      fetchPlacements(levelId)
    } else {
      setPlacements([])
    }
  }, [levelId, fetchPlacements])

  // Execute toggle action with optimistic update and WebSocket / REST fallback
  const toggleDevice = useCallback(
    async (entityId: string) => {
      // 1. Optimistic update
      setDeviceMap((prev) => {
        const current = prev[entityId]
        if (!current) return prev
        const nextState = current.state === 'on' ? 'off' : 'on'
        return {
          ...prev,
          [entityId]: {
            ...current,
            state: nextState,
          },
        }
      })

      // 2. Try WebSocket send first
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

      // 3. REST fallback
      try {
        const res = await fetch('/api/actions', {
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
          console.warn('REST action fallback failed, refreshing devices')
          fetchDevices()
        }
      } catch (err) {
        console.error('REST action failed:', err)
        fetchDevices()
      }
    },
    [fetchDevices]
  )

  return {
    deviceMap,
    devices: Object.values(deviceMap),
    placements,
    loading,
    connected,
    toggleDevice,
    refreshPlacements: () => levelId && fetchPlacements(levelId),
    refreshDevices: fetchDevices,
  }
}
