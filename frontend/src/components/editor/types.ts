export type SelectionType = 'wall' | 'zone' | 'device' | 'opening'

export interface EditorSelection {
  type: SelectionType
  id: string
  /** Owning wall id when `type` is `opening`. */
  wallId?: string
}

export type OpeningDragMode = 'move' | 'resize-start' | 'resize-end'
