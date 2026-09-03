<script setup lang="ts">
definePageMeta({
  layout: 'blank',
})

const { user, fetchMe, switchRole, canSwitchRole, roleSelectionPending, accessToken } = useAuth()
const errorMessage = ref('')
const selectingKey = ref<string | null>(null)

onMounted(async () => {
  if (!accessToken.value) {
    await navigateTo('/auth/login')
    return
  }
  if (!user.value) {
    await fetchMe()
  }
  if (!user.value) {
    await navigateTo('/auth/login')
    return
  }
  if (!canSwitchRole.value && !roleSelectionPending.value) {
    await navigateTo('/')
  }
})

const assignments = computed(() => user.value?.assignments || [])

const activeKey = computed(() => {
  if (!user.value?.active_role_id || !user.value?.active_scope_id) return ''
  return `${user.value.active_role_id}:${user.value.active_scope_id}`
})

function assignmentKey(roleId: number, scopeId: number) {
  return `${roleId}:${scopeId}`
}

async function onSelect(roleId: number, scopeId: number) {
  errorMessage.value = ''
  const key = assignmentKey(roleId, scopeId)
  selectingKey.value = key
  try {
    await switchRole(roleId, scopeId)
    await navigateTo('/')
  } catch (err: unknown) {
    const data = err as { data?: { message?: string }; message?: string }
    errorMessage.value = data?.data?.message || data?.message || 'Failed to switch role'
  } finally {
    selectingKey.value = null
  }
}
</script>

<template>
  <div class="authentication">
    <v-container fluid class="pa-3">
      <v-row class="h-100vh d-flex justify-center align-center">
        <v-col cols="12" class="d-flex align-center">
          <div class="boxed-auth-wrap" style="max-width: 520px; width: 100%">
            <v-card rounded="xl" elevation="10" class="px-sm-1 px-0 mx-auto index-2">
              <v-card-item class="pa-sm-8">
                <div class="d-flex justify-center mb-5">
                  <LayoutFullLogoDark />
                </div>
                <div class="text-h5 text-center mb-2">Choose your role</div>
                <div class="text-body-2 text-medium-emphasis text-center mb-6">
                  You have more than one role assignment. Select which context to use for this session.
                </div>

                <v-alert v-if="errorMessage" type="error" variant="tonal" density="compact" class="mb-4">
                  {{ errorMessage }}
                </v-alert>

                <v-list lines="two" class="bg-transparent">
                  <v-list-item
                    v-for="a in assignments"
                    :key="assignmentKey(a.role_id, a.scope_id)"
                    rounded="lg"
                    class="mb-2 border"
                    :active="activeKey === assignmentKey(a.role_id, a.scope_id)"
                    color="primary"
                    @click="onSelect(a.role_id, a.scope_id)"
                  >
                    <v-list-item-title class="font-weight-medium">
                      {{ a.role_name }}
                    </v-list-item-title>
                    <v-list-item-subtitle>
                      Scope: {{ a.scope_name }}
                      <span class="text-medium-emphasis">({{ a.scope_code }})</span>
                    </v-list-item-subtitle>
                    <template #append>
                      <v-progress-circular
                        v-if="selectingKey === assignmentKey(a.role_id, a.scope_id)"
                        indeterminate
                        size="20"
                        width="2"
                      />
                      <v-chip
                        v-else-if="activeKey === assignmentKey(a.role_id, a.scope_id)"
                        size="small"
                        color="primary"
                        variant="tonal"
                      >
                        Active
                      </v-chip>
                    </template>
                  </v-list-item>
                </v-list>
              </v-card-item>
            </v-card>
          </div>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>
