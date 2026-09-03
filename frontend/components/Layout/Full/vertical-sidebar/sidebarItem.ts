export interface menu {
  header?: string
  title?: string
  icon?: any
  to?: string
  chip?: string
  BgColor?: string
  chipBgColor?: string
  chipColor?: string
  chipVariant?: string
  chipIcon?: string
  children?: menu[]
  disabled?: boolean
  type?: string
  subCaption?: string
  external?: boolean
  /** If set, item is shown only when the user has this permission. */
  permission?: string
}

const sidebarItem: menu[] = [
  { header: 'Home' },
  {
    title: 'Dashboard',
    icon: 'graph-new-linear',
    to: '/',
  },

  { header: 'Access Control' },
  {
    title: 'Users',
    icon: 'user-linear',
    to: '/admin/users',
    permission: 'user.view',
  },
  {
    title: 'Roles',
    icon: 'shield-user-outline',
    to: '/admin/roles',
    permission: 'role.view',
  },
  {
    title: 'Scopes',
    icon: 'map-point-outline',
    to: '/admin/scopes',
    permission: 'scope.view',
  },

  { header: 'Site' },
  {
    title: 'Web Settings',
    icon: 'settings-minimalistic-linear',
    to: '/admin/settings',
    permission: 'setting.view',
  },

  { header: 'ui' },
  {
    title: 'Alert',
    icon: 'volume-small-broken',
    to: '/ui-components/Alerts',
  },
  {
    title: 'Button',
    icon: 'tag-horizontal-outline',
    to: '/ui-components/Buttons',
  },
  {
    title: 'Cards',
    icon: 'cardholder-linear',
    to: '/ui-components/Cards',
  },
  {
    title: 'Tables',
    icon: 'suspension-outline',
    to: '/ui-components/Tables',
  },

  { header: 'Auth' },
  {
    title: 'Login',
    icon: 'user-linear',
    to: '/auth/login',
  },
  {
    title: 'Register',
    icon: 'user-hand-up-linear',
    to: '/auth/register',
  },

  { header: 'Extra' },
  {
    title: 'Icons',
    icon: 'ghost-smile-outline',
    to: '/pages/Icons',
  },
  {
    title: 'Sample Page',
    icon: 'book-bookmark-linear',
    to: '/pages/sample-page',
  },
]

/** Hide items the user cannot access; drop section headers with no visible children. */
export function filterSidebarByPermission(
  items: menu[],
  can: (permission: string) => boolean,
): menu[] {
  const out: menu[] = []
  let pendingHeader: menu | null = null

  for (const item of items) {
    if (item.header) {
      pendingHeader = item
      continue
    }

    if (item.permission && !can(item.permission)) {
      continue
    }

    if (item.children?.length) {
      const children = item.children.filter((child) => !child.permission || can(child.permission))
      if (!children.length) continue
      if (pendingHeader) {
        out.push(pendingHeader)
        pendingHeader = null
      }
      out.push({ ...item, children })
      continue
    }

    if (pendingHeader) {
      out.push(pendingHeader)
      pendingHeader = null
    }
    out.push(item)
  }

  return out
}

export default sidebarItem
