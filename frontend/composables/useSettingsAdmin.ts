export type WebSettingItem = {
  key: string
  value: string
  type: 'string' | 'text' | 'json' | 'image' | string
}

export type SettingDraft = {
  key: string
  value: string
  type: string
  _isNew?: boolean
}

/** Known keys + naming heuristics. Type is never picked manually in the UI. */
const knownKeyTypes: Record<string, string> = {
  site_name: 'string',
  logo: 'image',
  contact_email: 'string',
  contact_phone: 'string',
  social_facebook: 'string',
  footer_text: 'string',
}

export function inferSettingType(key: string, value = ''): string {
  const k = key.trim().toLowerCase()
  if (!k) return 'string'
  if (knownKeyTypes[k]) return knownKeyTypes[k]

  if (/(^|_)(logo|image|avatar|banner|icon|thumbnail|cover)(_|$)/.test(k)) {
    return 'image'
  }
  if (/(^|_)(json|config|meta)(_|$)/.test(k) || looksLikeJson(value)) {
    return 'json'
  }
  if (/(^|_)(text|body|content|description|html|bio)(_|$)/.test(k) || value.includes('\n')) {
    return 'text'
  }
  return 'string'
}

function looksLikeJson(value: string) {
  const v = value.trim()
  if (!v || (v[0] !== '{' && v[0] !== '[')) return false
  try {
    JSON.parse(v)
    return true
  } catch {
    return false
  }
}

export function useSettingsAdmin() {
  const { api } = useApi()
  const items = ref<WebSettingItem[]>([])
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')

  async function fetchSettings() {
    loading.value = true
    error.value = ''
    try {
      const res = await api<WebSettingItem[]>('/admin/settings')
      items.value = res.data || []
    } catch (err: unknown) {
      const e = err as { data?: { message?: string } }
      error.value = e?.data?.message || 'Failed to load settings'
    } finally {
      loading.value = false
    }
  }

  async function saveSettings(settings: Array<{ key: string; value: string; type: string }>) {
    saving.value = true
    error.value = ''
    try {
      const res = await api<WebSettingItem[]>('/admin/settings', {
        method: 'PUT',
        body: { settings },
      })
      items.value = res.data || []
    } catch (err: unknown) {
      const e = err as { data?: { message?: string } }
      error.value = e?.data?.message || 'Failed to save settings'
      throw err
    } finally {
      saving.value = false
    }
  }

  return { items, loading, saving, error, fetchSettings, saveSettings }
}
