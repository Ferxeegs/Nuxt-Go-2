export type AdminUser = {
  id: string
  name: string
  email: string
  roles: Array<{ id: number; name: string }>
  avatar?: string
  is_active: boolean
  assignments: Array<{
    role_id: number
    role_name?: string
    scope_id: number
    scope_code?: string
    scope_name?: string
  }>
}

export type AssignmentDraft = { role_id: number | null; scope_id: number | null }

export function useUsersAdmin() {
  const { api } = useApi()
  const items = ref<AdminUser[]>([])
  const loading = ref(false)
  const error = ref('')
  const search = ref('')

  async function fetchUsers() {
    loading.value = true
    error.value = ''
    try {
      const res = await api<AdminUser[]>('/admin/users', {
        query: { page: 1, per_page: 100, search: search.value || undefined },
      })
      items.value = res.data || []
    } catch (err: unknown) {
      const e = err as { data?: { message?: string } }
      error.value = e?.data?.message || 'Failed to load users'
    } finally {
      loading.value = false
    }
  }

  async function createUser(payload: {
    name: string
    email: string
    password: string
    assignments: Array<{ role_id: number; scope_id: number }>
  }) {
    await api('/admin/users', { method: 'POST', body: payload })
    await fetchUsers()
  }

  async function updateUser(
    id: string,
    payload: {
      name: string
      email: string
      is_active: boolean
      assignments: Array<{ role_id: number; scope_id: number }>
    },
  ) {
    await api(`/admin/users/${id}`, { method: 'PUT', body: payload })
    await fetchUsers()
  }

  async function deleteUser(id: string) {
    await api(`/admin/users/${id}`, { method: 'DELETE' })
    await fetchUsers()
  }

  async function resetUserPassword(id: string) {
    await api(`/admin/users/${id}/reset-password`, { method: 'POST' })
  }

  async function getUser(id: string) {
    const res = await api<AdminUser>(`/admin/users/${id}`)
    return res.data as AdminUser
  }

  return {
    getUser,
    items,
    loading,
    error,
    search,
    fetchUsers,
    createUser,
    updateUser,
    deleteUser,
    resetUserPassword,
  }
}
