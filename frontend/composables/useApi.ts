export type ApiEnvelope<T> = {
  status: string
  message: string
  data?: T
  meta?: {
    page: number
    per_page: number
    total_data: number
    total_pages: number
  }
}

type ApiOptions = {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: unknown
  query?: Record<string, string | number | boolean | undefined | null>
  auth?: boolean
  _retried?: boolean
}

export function useApi() {
  const config = useRuntimeConfig()
  const cookieSecure = Boolean(config.public.cookieSecure)

  // Tokens are readable by JS so the SPA can send Authorization: Bearer.
  // HttpOnly is not possible here — only a server Set-Cookie can set HttpOnly,
  // and then JS could not attach the bearer header.
  const cookieOpts = {
    sameSite: 'lax' as const,
    secure: cookieSecure,
    httpOnly: false,
    watch: true,
  }

  const accessToken = useCookie<string | null>('access_token', cookieOpts)
  const refreshToken = useCookie<string | null>('refresh_token', cookieOpts)

  const baseURL = computed(() => {
    const raw = (config.public.apiUrl as string) || ''
    return `${raw.replace(/\/$/, '')}/api/v1`
  })

  async function refreshAccessToken(): Promise<boolean> {
    if (!refreshToken.value) {
      return false
    }
    try {
      const res = await $fetch<ApiEnvelope<{
        access_token: string
        refresh_token: string
        token_type: string
        expires_at: string
      }>>(`${baseURL.value}/auth/refresh`, {
        method: 'POST',
        body: { refresh_token: refreshToken.value },
      })
      if (!res?.data?.access_token) {
        return false
      }
      accessToken.value = res.data.access_token
      if (res.data.refresh_token) {
        refreshToken.value = res.data.refresh_token
      }
      return true
    } catch {
      return false
    }
  }

  async function clearSessionAndRedirect() {
    accessToken.value = null
    refreshToken.value = null
    if (import.meta.client && !window.location.pathname.startsWith('/auth')) {
      await navigateTo('/auth/login')
    }
  }

  async function api<T>(path: string, options: ApiOptions = {}): Promise<ApiEnvelope<T>> {
    const headers: Record<string, string> = {}
    if (options.auth !== false && accessToken.value) {
      headers.Authorization = `Bearer ${accessToken.value}`
    }

    try {
      return await $fetch<ApiEnvelope<T>>(`${baseURL.value}${path}`, {
        method: options.method || 'GET',
        body: options.body as BodyInit | Record<string, unknown> | null | undefined,
        query: options.query as Record<string, unknown> | undefined,
        headers,
      })
    } catch (err: unknown) {
      const status = (err as { statusCode?: number; status?: number })?.statusCode
        ?? (err as { statusCode?: number; status?: number })?.status
      if (status === 401 && options.auth !== false && !options._retried) {
        const refreshed = await refreshAccessToken()
        if (refreshed) {
          return api<T>(path, { ...options, _retried: true })
        }
        await clearSessionAndRedirect()
      }
      throw err
    }
  }

  return { api, accessToken, refreshToken, refreshAccessToken, clearSessionAndRedirect, baseURL }
}
