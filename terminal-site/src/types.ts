export type SectionId = 'terminal' | 'usage' | 'readme'

export type Side = 'left' | 'right' | 'top' | 'bottom'

export type PaneNode = PaneLeaf | PaneSplit

export interface PaneLeaf {
  type: 'leaf'
  paneId: number
  sectionId: SectionId
}

export interface PaneSplit {
  type: 'split'
  dir: 'row' | 'col'
  sizes: [number, number]
  children: [PaneNode, PaneNode]
}

export const SECTION_LABEL: Record<SectionId, string> = {
  terminal: 'shunpo',
  usage: 'usage',
  readme: 'readme',
}

export const SECTION_IDS: SectionId[] = ['terminal', 'usage', 'readme']
