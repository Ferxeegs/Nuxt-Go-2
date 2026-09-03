import { requiredPermissionForPath } from '~/utils/permissions'

export default defineNuxtRouteMiddleware(async (to) => {
  const { accessToken, fetchMe, user, roleSelectionPending, canSwitchRole, hasPermission } =
    useAuth()

  if (to.path === '/auth/choose-role') {
    if (!accessToken.value) {
      return navigateTo('/auth/login')
    }
    if (!user.value) {
      const me = await fetchMe()
      if (!me) {
        return navigateTo('/auth/login')
      }
    }
    if (!canSwitchRole.value && !roleSelectionPending.value) {
      return navigateTo('/')
    }
    return
  }

  if (to.path.startsWith('/auth')) {
    return
  }

  if (!accessToken.value) {
    return navigateTo('/auth/login')
  }

  if (!user.value) {
    const me = await fetchMe()
    if (!me) {
      return navigateTo('/auth/login')
    }
  }

  if (roleSelectionPending.value) {
    return navigateTo('/auth/choose-role')
  }

  const required = requiredPermissionForPath(to.path)
  if (required && !hasPermission(required)) {
    return navigateTo('/')
  }
})
