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
  { header: 'Overview' },
  {
    title: 'Dashboard',
    icon: 'mdi-view-dashboard-outline',
    to: '/',
  },

  { header: 'Access Control' },
  {
    title: 'Users',
    icon: 'mdi-account-group-outline',
    to: '/admin/users',
    permission: 'user.view',
  },
  {
    title: 'Roles',
    icon: 'mdi-shield-account-outline',
    to: '/admin/roles',
    permission: 'role.view',
  },
  {
    title: 'Scopes',
    icon: 'mdi-map-marker-radius-outline',
    to: '/admin/scopes',
    permission: 'scope.view',
  },

  { header: 'Site' },
  {
    title: 'Web Settings',
    icon: 'mdi-cog-outline',
    to: '/admin/settings',
    permission: 'setting.view',
  },

  { header: 'UI Kit' },
  {
    title: 'Components',
    icon: 'mdi-shape-outline',
    children: [
      { title: 'Alerts', to: '/ui-components/Alerts' },
      { title: 'Buttons', to: '/ui-components/Buttons' },
      { title: 'Cards', to: '/ui-components/Cards' },
      { title: 'Tables', to: '/ui-components/Tables' },
      { title: 'Icons', to: '/pages/Icons' },
      { title: 'Sample Page', to: '/pages/sample-page' },
    ],
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
