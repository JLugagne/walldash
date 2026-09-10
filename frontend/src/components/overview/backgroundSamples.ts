export interface BackgroundSample {
  id: string
  name: string
  url: string
}

/**
 * Backgrounds bundled with the application (frontend/public/backgrounds).
 * All are CC0 / public domain — see frontend/public/backgrounds/CREDITS.md.
 */
export const BACKGROUND_SAMPLES: BackgroundSample[] = [
  { id: 'desert-night', name: 'Desert night', url: '/backgrounds/desert-night.jpg' },
  { id: 'sunset-volcano', name: 'Milky Way trail', url: '/backgrounds/sunset-volcano.jpg' },
  { id: 'venice-sunset', name: 'Coastal dusk', url: '/backgrounds/venice-sunset.jpg' },
  { id: 'meadow', name: 'Starlit meadow', url: '/backgrounds/meadow.jpg' },
  { id: 'mountains-lake', name: 'Mountain lake', url: '/backgrounds/mountains-lake.jpg' },
]

export const DEFAULT_BACKGROUND_OPACITY = 95
export const DEFAULT_BACKGROUND_BLUR = 14
export const DEFAULT_BACKGROUND_DIM = 50
