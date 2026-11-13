<script setup lang="ts">
import type { AdminScope } from '~/composables/useScopesAdmin'

useHead({ title: 'Scopes' })

const { items, loading, error, search, fetchScopes, deleteScope } = useScopesAdmin()
const { hasPermission } = useAuth()

const canCreate = computed(() => hasPermission('scope.create'))
const canUpdate = computed(() => hasPermission('scope.update'))
const canDelete = computed(() => hasPermission('scope.delete'))
const showActions = computed(() => canUpdate.value || canDelete.value)

onMounted(fetchScopes)

async function onDelete(scope: AdminScope) {
  if (!confirm(`Delete scope ${scope.code}?`)) return
  try {
    await deleteScope(scope.id)
  } catch (err: unknown) {
    const e = err as { data?: { message?: string } }
    error.value = e?.data?.message || 'Delete failed'
  }
}
</script>

<template>
  <div>
    <AdminPageHeader title="Scopes" subtitle="Organisational boundaries for scoped role assignments">
      <v-btn v-if="canCreate" color="primary" prepend-icon="mdi-plus" to="/admin/scopes/create">
        New scope
      </v-btn>
    </AdminPageHeader>

    <v-card>
      <div class="d-flex flex-wrap ga-3 pa-4">
        <v-text-field
          v-model="search"
          placeholder="Search scopes…"
          prepend-inner-icon="mdi-magnify"
          density="compact"
          hide-details
          clearable
          style="max-width: 320px"
          @keyup.enter="fetchScopes"
          @click:clear="() => { search = ''; fetchScopes() }"
        />
        <v-btn variant="tonal" @click="fetchScopes">Search</v-btn>
      </div>

      <v-alert v-if="error" type="error" variant="tonal" class="mx-4 mb-4" density="compact">{{ error }}</v-alert>

      <v-table>
        <thead>
          <tr>
            <th>Code</th>
            <th>Name</th>
            <th>Parent</th>
            <th>Status</th>
            <th v-if="showActions" class="text-right">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td :colspan="showActions ? 5 : 4" class="text-center py-8">Loading…</td>
          </tr>
          <tr v-else-if="!items.length">
            <td :colspan="showActions ? 5 : 4" class="text-center py-8 text-medium-emphasis">
              No scopes found
            </td>
          </tr>
          <tr v-for="scope in items" :key="scope.id">
            <td><code>{{ scope.code }}</code></td>
            <td class="font-weight-medium">{{ scope.name }}</td>
            <td class="text-medium-emphasis">{{ scope.parent_code || '—' }}</td>
            <td>
              <v-chip size="small" variant="tonal" :color="scope.is_active ? 'success' : undefined">
                {{ scope.is_active ? 'Active' : 'Inactive' }}
              </v-chip>
            </td>
            <td v-if="showActions" class="text-right text-no-wrap">
              <v-btn
                v-if="canUpdate"
                size="small"
                variant="text"
                prepend-icon="mdi-pencil-outline"
                :to="`/admin/scopes/${scope.id}`"
              >
                Edit
              </v-btn>
              <v-btn
                v-if="canDelete"
                size="small"
                variant="text"
                color="error"
                prepend-icon="mdi-trash-can-outline"
                @click="onDelete(scope)"
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
