<script setup lang="ts">
import type { AdminUser } from '~/composables/useUsersAdmin'

useHead({ title: 'Users' })

const { items, loading, error, search, fetchUsers, deleteUser } = useUsersAdmin()
const {
  user: authUser,
  canImpersonateTarget,
  impersonate,
  isImpersonating,
  hasPermission,
} = useAuth()

const canCreate = computed(() => hasPermission('user.create'))
const canUpdate = computed(() => hasPermission('user.update'))
const canDelete = computed(() => hasPermission('user.delete'))
const showActions = computed(
  () => canUpdate.value || canDelete.value || hasPermission('user.impersonate'),
)

const impersonatingId = ref<string | null>(null)

onMounted(fetchUsers)

function initials(name: string) {
  return name
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase())
    .join('')
}

async function onDelete(user: AdminUser) {
  if (!confirm(`Delete user ${user.email}?`)) return
  try {
    await deleteUser(user.id)
  } catch (err: unknown) {
    const e = err as { data?: { message?: string } }
    error.value = e?.data?.message || 'Delete failed'
  }
}

async function onImpersonate(target: AdminUser) {
  if (!canImpersonateTarget(target)) return
  if (
    !confirm(
      `Impersonate ${target.name} (${target.email})?\n\nYou will see the app as this user until you stop impersonating.`,
    )
  ) {
    return
  }
  impersonatingId.value = target.id
  error.value = ''
  try {
    await impersonate(target.id)
  } catch (err: unknown) {
    const e = err as { data?: { message?: string }; message?: string }
    error.value = e?.data?.message || e?.message || 'Impersonation failed'
  } finally {
    impersonatingId.value = null
  }
}
</script>

<template>
  <div>
    <AdminPageHeader title="Users" subtitle="Manage accounts and their role assignments">
      <v-btn v-if="canCreate" color="primary" prepend-icon="mdi-plus" to="/admin/users/create">
        New user
      </v-btn>
    </AdminPageHeader>

    <v-card>
      <div class="d-flex flex-wrap ga-3 pa-4">
        <v-text-field
          v-model="search"
          placeholder="Search by name or email…"
          prepend-inner-icon="mdi-magnify"
          density="compact"
          hide-details
          clearable
          style="max-width: 320px"
          @keyup.enter="fetchUsers"
          @click:clear="() => { search = ''; fetchUsers() }"
        />
        <v-btn variant="tonal" @click="fetchUsers">Search</v-btn>
      </div>

      <v-alert v-if="error" type="error" variant="tonal" class="mx-4 mb-4" density="compact">{{ error }}</v-alert>

      <v-table>
        <thead>
          <tr>
            <th>User</th>
            <th>Assignments</th>
            <th>Status</th>
            <th v-if="showActions" class="text-right">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td :colspan="showActions ? 4 : 3" class="text-center py-8">Loading…</td>
          </tr>
          <tr v-else-if="!items.length">
            <td :colspan="showActions ? 4 : 3" class="text-center py-8 text-medium-emphasis">
              No users found
            </td>
          </tr>
          <tr v-for="user in items" :key="user.id">
            <td>
              <div class="d-flex align-center ga-3 py-1">
                <v-avatar size="36" color="primary" variant="tonal">
                  <v-img v-if="user.avatar" :src="user.avatar" cover />
                  <span v-else class="text-caption font-weight-bold">{{ initials(user.name) }}</span>
                </v-avatar>
                <div class="min-w-0">
                  <div class="font-weight-medium">{{ user.name }}</div>
                  <div class="text-caption text-medium-emphasis">{{ user.email }}</div>
                </div>
              </div>
            </td>
            <td>
              <span v-if="!(user.assignments || []).length" class="text-medium-emphasis">—</span>
              <div v-else class="d-flex flex-wrap ga-1">
                <v-chip
                  v-for="(a, i) in user.assignments"
                  :key="i"
                  size="small"
                  variant="tonal"
                  color="primary"
                >
                  {{ a.role_name }}
                  <span class="text-medium-emphasis ml-1">· {{ a.scope_code || a.scope_name }}</span>
                </v-chip>
              </div>
            </td>
            <td>
              <v-chip size="small" variant="tonal" :color="user.is_active ? 'success' : undefined">
                {{ user.is_active ? 'Active' : 'Inactive' }}
              </v-chip>
            </td>
            <td v-if="showActions" class="text-right text-no-wrap">
              <v-btn
                v-if="canImpersonateTarget(user)"
                size="small"
                variant="text"
                color="warning"
                :loading="impersonatingId === user.id"
                :disabled="isImpersonating"
                @click="onImpersonate(user)"
              >
                Impersonate
              </v-btn>
              <v-btn
                v-if="canUpdate"
                size="small"
                variant="text"
                prepend-icon="mdi-pencil-outline"
                :disabled="isImpersonating"
                :to="`/admin/users/${user.id}`"
              >
                Edit
              </v-btn>
              <v-btn
                v-if="canDelete"
                size="small"
                variant="text"
                color="error"
                prepend-icon="mdi-trash-can-outline"
                :disabled="isImpersonating || user.id === authUser?.id"
                @click="onDelete(user)"
              >
                Delete
              </v-btn>
            </td>
          </tr>
        </tbody>
      </v-table>
    </v-card>
  </div>
</template>
