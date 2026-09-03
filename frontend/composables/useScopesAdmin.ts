export type AdminScope = {
  id: number
  code: string
  name: string
  description: string
  parent_id?: number | null
  parent_code?: string
  parent_name?: string
  is_active: boolean
}

export function useScopesAdmin() {
  const { api } = useApi()
  const items = ref<AdminScope[]>([])
  const loading = ref(false)
  const error = ref('')
  const search = ref('')

  async function fetchScopes() {
    loading.value = true
    error.value = ''
    try {
      const res = await api<AdminScope[]>('/admin/scopes', {
        query: { page: 1, per_page: 100, search: search.value || undefined },
      })
      items.value = res.data || []
    } catch (err: unknown) {
      const e = err as { data?: { message?: string } }
      error.value = e?.data?.message || 'Failed to load scopes'
    } finally {
      loading.value = false
    }
  }

  async function createScope(payload: {
    code: string
    name: string
    description: string
    parent_id?: number | null
    is_active: boolean
  }) {
    await api('/admin/scopes', {
      method: 'POST',
      body: {
        ...payload,
        parent_id: payload.parent_id || undefined,
      },
    })
    await fetchScopes()
  }

  async function updateScope(
    id: number,
    payload: {
      code: string
      name: string
      description: string
      parent_id?: number | null
      clear_parent?: boolean
      is_active: boolean
    },
  ) {
    await api(`/admin/scopes/${id}`, { method: 'PUT', body: payload })
    await fetchScopes()
  }

  async function deleteScope(id: number) {
    await api(`/admin/scopes/${id}`, { method: 'DELETE' })
    await fetchScopes()
  }

  return { items, loading, error, search, fetchScopes, createScope, updateScope, deleteScope }
}
