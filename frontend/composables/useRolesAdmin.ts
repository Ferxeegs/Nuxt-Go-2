export type AdminRole = {
  id: number
  name: string
  description: string
  permissions: string[]
}

export type PermissionOption = {
  id: number
  name: string
  description: string
}

export function useRolesAdmin() {
  const { api } = useApi()
  const items = ref<AdminRole[]>([])
  const permissions = ref<PermissionOption[]>([])
  const loading = ref(false)
  const error = ref('')

  async function fetchRoles() {
    loading.value = true
    error.value = ''
    try {
      const res = await api<AdminRole[]>('/admin/roles')
      items.value = res.data || []
    } catch (err: unknown) {
      const e = err as { data?: { message?: string } }
      error.value = e?.data?.message || 'Failed to load roles'
    } finally {
      loading.value = false
    }
  }

  async function fetchPermissions() {
    const res = await api<PermissionOption[]>('/admin/roles/permissions')
    permissions.value = res.data || []
  }

  async function createRole(payload: {
    name: string
    description: string
    permission_ids: number[]
  }) {
    await api('/admin/roles', { method: 'POST', body: payload })
    await fetchRoles()
  }

  async function updateRole(
    id: number,
    payload: { name: string; description: string; permission_ids: number[] },
  ) {
    await api(`/admin/roles/${id}`, { method: 'PUT', body: payload })
    await fetchRoles()
  }

  async function deleteRole(id: number) {
    await api(`/admin/roles/${id}`, { method: 'DELETE' })
    await fetchRoles()
  }

  async function getRole(id: number) {
    const res = await api<AdminRole>(`/admin/roles/${id}`)
    return res.data as AdminRole
  }

  return {
    getRole,
    items,
    permissions,
    loading,
    error,
    fetchRoles,
    fetchPermissions,
    createRole,
    updateRole,
    deleteRole,
  }
}
