/** Path prefix → required view permission for page access. */
export const routePermissions: Array<{ prefix: string; permission: string }> = [
  { prefix: '/admin/users', permission: 'user.view' },
  { prefix: '/admin/roles', permission: 'role.view' },
  { prefix: '/admin/scopes', permission: 'scope.view' },
  { prefix: '/admin/settings', permission: 'setting.view' },
]

export function requiredPermissionForPath(path: string): string | null {
  const normalized = path.replace(/\/+$/, '') || '/'
  for (const rule of routePermissions) {
    if (normalized === rule.prefix || normalized.startsWith(`${rule.prefix}/`)) {
      return rule.permission
    }
  }
  return null
}
