<script setup lang="ts">
const { logout, user, isImpersonating, stopImpersonate, canSwitchRole } = useAuth()
const loggingOut = ref(false)
const stopping = ref(false)

const role = computed(() => user.value?.active_role_name || user.value?.role || '')
const scope = computed(() => user.value?.active_scope_name || '')
const initials = computed(() =>
  (user.value?.name || '?')
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase())
    .join(''),
)

async function onLogout() {
  loggingOut.value = true
  try {
    await logout()
    await navigateTo('/auth/login')
  } finally {
    loggingOut.value = false
  }
}

async function onStopImpersonate() {
  stopping.value = true
  try {
    await stopImpersonate()
  } finally {
    stopping.value = false
  }
}
</script>

<template>
  <v-menu location="bottom end" offset="10" :close-on-content-click="true">
    <template #activator="{ props }">
      <button v-bind="props" class="user-trigger" type="button">
        <v-badge v-if="isImpersonating" dot color="warning" location="top end">
          <v-avatar size="36" color="primary" variant="tonal">
            <v-img v-if="user?.avatar" :src="user.avatar" cover />
            <span v-else class="text-caption font-weight-bold">{{ initials }}</span>
          </v-avatar>
        </v-badge>
        <v-avatar v-else size="36" color="primary" variant="tonal">
          <v-img v-if="user?.avatar" :src="user.avatar" cover />
          <span v-else class="text-caption font-weight-bold">{{ initials }}</span>
        </v-avatar>
        <span class="user-trigger__meta d-none d-md-flex">
          <span class="user-trigger__name">{{ user?.name || 'Account' }}</span>
          <span class="user-trigger__role">{{ role }}</span>
        </span>
        <v-icon icon="mdi-chevron-down" size="18" class="d-none d-md-inline text-medium-emphasis" />
      </button>
    </template>

    <v-card class="user-menu" min-width="270" elevation="8">
      <div v-if="user" class="user-menu__head">
        <div class="text-subtitle-2 font-weight-bold text-truncate">{{ user.name }}</div>
        <div class="text-caption text-medium-emphasis text-truncate">{{ user.email }}</div>
        <div class="d-flex flex-wrap ga-1 mt-2">
          <v-chip v-if="role" size="x-small" color="primary" variant="tonal">{{ role }}</v-chip>
          <v-chip v-if="scope" size="x-small" variant="tonal">{{ scope }}</v-chip>
          <v-chip v-if="isImpersonating" size="x-small" color="warning" variant="tonal">
            Impersonating
          </v-chip>
        </div>
      </div>
      <v-divider />
      <v-list density="compact" nav class="pa-2">
        <v-list-item to="/profile" prepend-icon="mdi-account-outline" title="My profile" rounded="lg" />
        <v-list-item
          v-if="canSwitchRole"
          to="/auth/choose-role"
          prepend-icon="mdi-swap-horizontal"
          title="Switch role"
          rounded="lg"
        />
        <v-list-item
          v-if="isImpersonating"
          prepend-icon="mdi-account-switch-outline"
          title="Stop impersonating"
          rounded="lg"
          :disabled="stopping"
          @click="onStopImpersonate"
        />
      </v-list>
      <v-divider />
      <div class="pa-2">
        <v-btn
          block
          variant="text"
          color="error"
          prepend-icon="mdi-logout"
          class="justify-start"
          :loading="loggingOut"
          @click="onLogout"
        >
          Sign out
        </v-btn>
      </div>
    </v-card>
  </v-menu>
</template>
