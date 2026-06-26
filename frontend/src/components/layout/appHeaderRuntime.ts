type HeaderUserLike = {
  username?: string | null
  email?: string | null
}

type HeaderMenuLike = {
  id: string
  label?: string | null
}

export function resolveHeaderDisplayName(user: HeaderUserLike | null | undefined): string {
  if (!user) return ''
  return user.username || user.email?.split('@')[0] || ''
}

export function resolveHeaderUserInitials(user: HeaderUserLike | null | undefined): string {
  if (!user) return ''
  if (user.username) {
    return user.username.substring(0, 2).toUpperCase()
  }
  if (user.email) {
    const localPart = user.email.split('@')[0]
    return localPart.substring(0, 2).toUpperCase()
  }
  return ''
}

export function resolveCompactUserDropdown(showPrimaryActions: boolean, showContactSupport: boolean) {
  return !showPrimaryActions && !showContactSupport
}

export function resolveHeaderPageTitle(options: {
  routeName?: string | symbol | null
  routeCustomId?: string
  routeMetaTitleKey?: string
  routeMetaTitle?: string
  publicMenuItems?: HeaderMenuLike[]
  adminMenuItems?: HeaderMenuLike[]
  isAdmin?: boolean
  translate: (key: string) => string
}): string {
  if (options.routeName === 'CustomPage' && options.routeCustomId) {
    const publicItem = options.publicMenuItems?.find((item) => item.id === options.routeCustomId)
    const adminItem = options.isAdmin
      ? options.adminMenuItems?.find((item) => item.id === options.routeCustomId)
      : undefined
    const menuItem = publicItem ?? adminItem
    if (menuItem?.label) return menuItem.label
  }
  if (options.routeMetaTitleKey) {
    return options.translate(options.routeMetaTitleKey)
  }
  return options.routeMetaTitle || ''
}
