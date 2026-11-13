<script setup lang="ts">
import type { AdminRole } from '~/composables/useRolesAdmin'

useHead({ title: 'Roles' })

const { items, loading, error, fetchRoles, deleteRole } = useRolesAdmin()
const { hasPermission } = useAuth()

const canCreate = computed(() => hasPermission('role.create'))
const canUpdate = computed(() => hasPermission('role.update'))
const canDelete = computed(() => hasPermission('role.delete'))
const showActions = computed(() => canUpdate.value || canDelete.value)

onMounted(fetchRoles)

async function onDelete(role: AdminRole) {
  if (!confirm(`Delete role ${role.name}?`)) return
  try {
    await deleteRole(role.id)
  } catch (err: unknown) {
    const e = err as { data?: { message?: string } }
    error.value = e?.data?.message || 'Delete failed'
  }
}
</script>

<template>
  <div>
    <AdminPageHeader title="Roles" subtitle="Named permission sets for global and scoped access">
      <v-btn v-if="canCreate" color="primary" prepend-icon="mdi-plus" to="/admin/roles/create">
        New role
      </v-btn>
    </AdminPageHeader>

    <v-card>
      <v-alert v-if="error" type="error" variant="tonal" class="ma-4" density="compact">{{ error }}</v-alert>
      <v-table>
        <thead>
          <tr>
            <th>Role</th>
            <th>Description</th>
            <th>Permissions</th>
            <th v-if="showActions" class="text-right">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td :colspan="showActions ? 4 : 3" class="text-center py-8">Loading…</td>
          </tr>
          <tr v-else-if="!items.length">
            <td :colspan="showActions ? 4 : 3" class="text-center py-8 text-medium-emphasis">
              No roles found
            </td>
          </tr>
          <tr v-for="role in items" :key="role.id">
            <td class="font-weight-medium">{{ role.name }}</td>
            <td class="text-medium-emphasis">{{ role.description || '—' }}</td>
            <td>
              <v-chip size="small" variant="tonal" color="primary">
                {{ (role.permissions || []).length }} permissions
              </v-chip>
            </td>
            <td v-if="showActions" class="text-right text-no-wrap">
              <v-btn
                v-if="canUpdate"
                size="small"
                variant="text"
                prepend-icon="mdi-pencil-outline"
                :to="`/admin/roles/${role.id}`"
              >
                Edit
              </v-btn>
              <v-btn
                v-if="canDelete"
                size="small"
                variant="text"
                color="error"
                prepend-icon="mdi-trash-can-outline"
                @click="onDelete(role)"
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
