import React from 'react'

export interface ToastProps {
  message: string
}

/**
 * Toast is the single transient banner of the Dashboard view. It is purely presentational and
 * positions itself relative to the nearest positioned ancestor, so the view must render exactly
 * one instance to keep two surfaces from painting over each other.
 */
export const Toast: React.FC<ToastProps> = ({ message }) => (
  <div className="absolute top-4 left-1/2 -translate-x-1/2 z-40 px-4 py-2 rounded-xl bg-red-500/90 text-white text-xs font-semibold shadow-xl">
    {message}
  </div>
)
