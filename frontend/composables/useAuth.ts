export type AuthAssignment = {
  role_id: number
  role_name: string
  scope_id: number
  scope_code: string
  scope_name: string
}

export type AuthUser = {
  id: string
  name: string
  email: string
  role: string
  role_id?: number
  roles: Array<{ id: number; name: string }>
  avatar?: string
  is_active: boolean
  permissions: string[]
  impersonator_id?: string | null
  assignments: AuthAssignment[]
  active_role_id?: number
  active_scope_id?: number
  active_role_name?: string
  active_scope_code?: string
  active_scope_name?: string
  requires_role_selection?: boolean
  can_switch_role?: boolean
  last_login_at?: string | null
  created_at?: string
}

type TokenPair = {
  access_token: string
  refresh_token: string
  token_type: string
  expires_at: string
  requires_role_selection?: boolean
  active_role_id?: number
  active_scope_id?: number
}

export function useAuth() {
  const { api, accessToken, refreshToken, clearSessionAndRedirect } = useApi()
  const config = useRuntimeConfig()
  const cookieSecure = Boolean(config.public.cookieSecure)

  const user = useState<AuthUser | null>('auth-user', () => null)
  const loading = useState<boolean>('auth-loading', () => false)

  const roleSelectionPending = useCookie<boolean | null>('role_selection_pending', {
    sameSite: 'lax',
    secure: cookieSecure,
    maxAge: 60 * 60 * 12,
    default: () => null,
  })

  const isImpersonating = computed(() => Boolean(user.value?.impersonator_id))
  const canSwitchRole = computed(() => Boolean(user.value?.can_switch_role || (user.value?.assignments?.length || 0) > 1))

  function hasPermission(permission: string) {
    if (!user.value) return false
    if (user.value.role === 'SuperAdmin') return true
    return (user.value.permissions || []).includes(permission)
  }

  function applyTokens(tokens: TokenPair) {
    accessToken.value = tokens.access_token
    refreshToken.value = tokens.refresh_token
  }

  function markRoleSelectionPending(required: boolean) {
    roleSelectionPending.value = required ? true : null
  }

  function clearRoleSelectionPending() {
    roleSelectionPending.value = null
  }

  async function login(email: string, password: string, turnstileToken = '') {
    const res = await api<TokenPair>('/auth/login', {
      method: 'POST',
      auth: false,
      body: {
        email,
        password,
        turnstile_token: turnstileToken || undefined,
      },
    })

    if (!res.data?.access_token) {
      throw new Error(res.message || 'Login failed')
    }

    applyTokens(res.data)
    markRoleSelectionPending(Boolean(res.data.requires_role_selection))
    await fetchMe()
    return {
      user: user.value,
      requiresRoleSelection: Boolean(res.data.requires_role_selection),
    }
  }

  async function fetchMe() {
    if (!accessToken.value) {
      user.value = null
      return null
    }
    loading.value = true
    try {
      const res = await api<AuthUser>('/auth/me')
      user.value = res.data || null
      return user.value
    } catch {
      user.value = null
      return null
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      if (accessToken.value) {
        await api('/auth/logout', { method: 'POST' })
      }
    } catch {
      // ignore logout API errors; clear local session regardless
    } finally {
      user.value = null
      clearRoleSelectionPending()
      await clearSessionAndRedirect()
    }
  }

  async function switchRole(roleId: number, scopeId: number) {
    const res = await api<TokenPair>('/auth/switch-role', {
      method: 'POST',
      body: { role_id: roleId, scope_id: scopeId },
    })
    if (!res.data?.access_token) {
      throw new Error(res.message || 'Failed to switch role')
    }
    applyTokens(res.data)
    clearRoleSelectionPending()
    await fetchMe()
    return user.value
  }

  async function impersonate(targetUserId: string) {
    if (isImpersonating.value) {
      throw new Error('Already impersonating another user; stop first')
    }
    const res = await api<TokenPair>('/auth/impersonate', {
      method: 'POST',
      body: { user_id: targetUserId },
    })
    if (!res.data?.access_token) {
      throw new Error(res.message || 'Impersonation failed')
    }
    applyTokens(res.data)
    markRoleSelectionPending(Boolean(res.data.requires_role_selection))
    await fetchMe()
    if (res.data.requires_role_selection) {
      await navigateTo('/auth/choose-role')
    } else {
      await navigateTo('/')
    }
    return user.value
  }

  async function stopImpersonate() {
    if (!isImpersonating.value) {
      throw new Error('Not currently impersonating')
    }
    const res = await api<TokenPair>('/auth/stop-impersonate', {
      method: 'POST',
    })
    if (!res.data?.access_token) {
      throw new Error(res.message || 'Failed to stop impersonation')
    }
    applyTokens(res.data)
    markRoleSelectionPending(Boolean(res.data.requires_role_selection))
    await fetchMe()
    if (res.data.requires_role_selection) {
      await navigateTo('/auth/choose-role')
    } else {
      await navigateTo('/admin/users')
    }
    return user.value
  }

  async function updateProfile(payload: { name?: string; email?: string; avatar?: string }) {
    if (isImpersonating.value) {
      throw new Error('Profile cannot be edited while impersonating')
    }
    const res = await api<AuthUser>('/auth/me', {
      method: 'PATCH',
      body: payload,
    })
    if (!res.data) {
      throw new Error(res.message || 'Failed to update profile')
    }
    user.value = res.data
    return user.value
  }

  async function changePassword(payload: { current_password: string; new_password: string }) {
    if (isImpersonating.value) {
      throw new Error('Password cannot be changed while impersonating')
    }
    const res = await api('/auth/me/password', {
      method: 'PATCH',
      body: payload,
    })
    if (res.status === 'error') {
      throw new Error(res.message || 'Failed to change password')
    }
    return true
  }

  function canImpersonateTarget(target: {
    id: string
    is_active: boolean
    roles?: Array<{ name: string }>
  }) {
    if (!user.value || isImpersonating.value) return false
    if (!hasPermission('user.impersonate')) return false
    if (target.id === user.value.id) return false
    if (!target.is_active) return false
    if ((target.roles || []).some((r) => r.name === 'SuperAdmin')) return false
    return true
  }

  function isAuthenticated() {
    return Boolean(accessToken.value)
  }

  return {
    user,
    loading,
    isImpersonating,
    canSwitchRole,
    roleSelectionPending,
    hasPermission,
    canImpersonateTarget,
    login,
    logout,
    fetchMe,
    updateProfile,
    changePassword,
    switchRole,
    clearRoleSelectionPending,
    impersonate,
    stopImpersonate,
    isAuthenticated,
    accessToken,
  }
}
