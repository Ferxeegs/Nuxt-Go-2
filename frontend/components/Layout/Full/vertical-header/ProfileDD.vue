<script setup lang="ts">
import { UserIcon, ArrowsExchangeIcon } from 'vue-tabler-icons'

const { logout, user, isImpersonating, stopImpersonate, canSwitchRole } = useAuth()
const loggingOut = ref(false)
const stopping = ref(false)

const activeLabel = computed(() => {
  if (!user.value) return ''
  const role = user.value.active_role_name || user.value.role
  const scope = user.value.active_scope_name
  if (role && scope) return `${role} · ${scope}`
  return role || ''
})

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

async function onSwitchRole() {
  await navigateTo('/auth/choose-role')
}

async function onMyProfile() {
  await navigateTo('/profile')
}
</script>

<template>
  <v-menu :close-on-content-click="true">
    <template #activator="{ props }">
      <v-btn class="" variant="text" v-bind="props" icon>
        <v-badge
          v-if="isImpersonating"
          color="warning"
          dot
          location="top end"
        >
          <v-avatar size="35">
            <img
              :src="user?.avatar || '/images/profile/user-1.jpg'"
              height="35"
              alt="user"
              style="object-fit: cover"
            />
          </v-avatar>
        </v-badge>
        <v-avatar v-else size="35">
          <img
            :src="user?.avatar || '/images/profile/user-1.jpg'"
            height="35"
            alt="user"
            style="object-fit: cover"
          />
        </v-avatar>
      </v-btn>
    </template>
    <v-sheet rounded="xl" width="260" elevation="10" class="mt-2">
      <div v-if="user" class="px-4 pt-4 pb-2">
        <div class="text-body-1 font-weight-medium text-truncate">{{ user.name }}</div>
        <div class="text-caption text-medium-emphasis text-truncate">{{ user.email }}</div>
        <div v-if="activeLabel" class="text-caption mt-1 text-truncate">{{ activeLabel }}</div>
        <v-chip
          v-if="isImpersonating"
          class="mt-2"
          size="small"
          color="warning"
          variant="tonal"
        >
          Impersonating
        </v-chip>
      </div>
      <v-list class="py-0" lines="one" density="compact">
        <v-list-item value="profile" color="primary" @click="onMyProfile">
          <template #prepend>
            <UserIcon stroke-width="1.5" size="20" />
          </template>
          <v-list-item-title class="pl-4 text-body-1">My Profile</v-list-item-title>
        </v-list-item>
        <v-list-item
          v-if="canSwitchRole"
          value="switch-role"
          color="primary"
          @click="onSwitchRole"
        >
          <template #prepend>
            <ArrowsExchangeIcon stroke-width="1.5" size="20" />
          </template>
          <v-list-item-title class="pl-4 text-body-1">Switch role</v-list-item-title>
        </v-list-item>
      </v-list>
      <div class="pt-3 pb-4 px-5 text-center d-flex flex-column ga-2">
        <v-btn
          v-if="isImpersonating"
          color="warning"
          variant="flat"
          class="rounded-pill"
          block
          :loading="stopping"
          @click="onStopImpersonate"
        >
          Stop impersonating
        </v-btn>
        <v-btn
          color="primary"
          variant="outlined"
          class="rounded-pill"
          block
          :loading="loggingOut"
          @click="onLogout"
        >
          Logout
        </v-btn>
      </div>
    </v-sheet>
  </v-menu>
</template>
