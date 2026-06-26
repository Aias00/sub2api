export type SidebarNavItem = {
  path: string
  label: string
  icon: unknown
  iconSvg?: string
  hideInSimpleMode?: boolean
  children?: SidebarNavItem[]
  expandOnly?: boolean
  featureFlag?: () => boolean | undefined
}

export type SidebarNavSection = {
  id: string
  items: SidebarNavItem[]
}

export function applySidebarFeatureFlags(items: SidebarNavItem[]): SidebarNavItem[] {
  const out: SidebarNavItem[] = []
  for (const item of items) {
    if (item.featureFlag && item.featureFlag() === false) continue
    if (item.children) {
      out.push({ ...item, children: applySidebarFeatureFlags(item.children) })
    } else {
      out.push(item)
    }
  }
  return out
}

export function buildSidebarVisibleItemMap<K extends string>(
  itemMap: Partial<Record<K, SidebarNavItem>>,
  isSimpleMode: boolean,
): Partial<Record<K, SidebarNavItem>> {
  const visibleMap: Partial<Record<K, SidebarNavItem>> = {}
  for (const [key, item] of Object.entries(itemMap) as Array<[K, SidebarNavItem | undefined]>) {
    if (!item) continue
    const visible = applySidebarFeatureFlags([item]).filter((candidate) =>
      isSimpleMode ? !candidate.hideInSimpleMode : true,
    )
    if (visible.length > 0) {
      visibleMap[key] = visible[0]
    }
  }
  return visibleMap
}

export function buildSidebarSections<K extends string>(
  configuredSections: Array<{ id: string; items: K[] }>,
  defaultSections: Array<{ id: string; items: K[] }>,
  visibleMap: Partial<Record<K, SidebarNavItem>>,
  customItems: SidebarNavItem[],
  fallbackSectionId: string,
  customSectionId: string,
): SidebarNavSection[] {
  const sectionsSource = configuredSections.length > 0 ? configuredSections : defaultSections
  const used = new Set<K>()
  const sections: SidebarNavSection[] = sectionsSource
    .map((section) => ({
      id: section.id,
      items: section.items
        .filter((key): key is K => Boolean(visibleMap[key]))
        .filter((key) => {
          if (used.has(key)) return false
          used.add(key)
          return true
        })
        .map((key) => visibleMap[key]!),
    }))
    .filter((section) => section.items.length > 0)

  const remainingBuiltIns = Object.entries(visibleMap)
    .filter(([key, item]) => Boolean(item) && !used.has(key as K))
    .map(([, item]) => item as SidebarNavItem)
  if (remainingBuiltIns.length > 0) {
    sections.push({ id: fallbackSectionId, items: remainingBuiltIns })
  }
  if (customItems.length > 0) {
    sections.push({ id: customSectionId, items: customItems })
  }
  return sections
}
