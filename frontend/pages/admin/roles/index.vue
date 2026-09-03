<script setup lang="ts">
import type { AdminRole } from '~/composables/useRolesAdmin'

const {
  items,
  permissions,
  loading,
  error,
  fetchRoles,
  fetchPermissions,
  createRole,
  updateRole,
  deleteRole,
} = useRolesAdmin()
const { hasPermission } = useAuth()

const canCreate = computed(() => hasPermission('role.create'))
const canUpdate = computed(() => hasPermission('role.update'))
const canDelete = computed(() => hasPermission('role.delete'))
const showActions = computed(() => canUpdate.value || canDelete.value)

const dialog = ref(false)
const saving = ref(false)
const formError = ref('')
const editing = ref<AdminRole | null>(null)

const form = reactive({
  name: '',
  description: '',
  permission_ids: [] as number[],
})

onMounted(async () => {
  await Promise.all([fetchRoles(), fetchPermissions()])
})

function openCreate() {
  editing.value = null
  form.name = ''
  form.description = ''
  form.permission_ids = []
  formError.value = ''
  dialog.value = true
}

function openEdit(role: AdminRole) {
  editing.value = role
  form.name = role.name
  form.description = role.description || ''
  const names = new Set(role.permissions || [])
  form.permission_ids = permissions.value.filter((p) => names.has(p.name)).map((p) => p.id)
  formError.value = ''
  dialog.value = true
}

async function save() {
  formError.value = ''
  saving.value = true
  try {
    const payload = {
      name: form.name,
      description: form.description,
      permission_ids: form.permission_ids,
    }
    if (editing.value) {
      await updateRole(editing.value.id, payload)
    } else {
      await createRole(payload)
    }
    dialog.value = false
  } catch (err: unknown) {
    const e = err as { data?: { message?: string } }
    formError.value = e?.data?.message || 'Save failed'
  } finally {
    saving.value = false
  }
}

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
  <v-container fluid>
    <div class="d-flex flex-wrap align-center justify-space-between mb-4 ga-3">
      <div>
        <h1 class="text-h5 font-weight-bold">Roles</h1>
        <p class="text-body-2 text-medium-emphasis mb-0">Named permission sets for global and scoped access</p>
      </div>
      <v-btn v-if="canCreate" color="primary" @click="openCreate">New role</v-btn>
    </div>

    <v-card rounded="lg" elevation="1">
      <v-alert v-if="error" type="error" variant="tonal" class="ma-4" density="compact">{{ error }}</v-alert>
      <v-table>
        <thead>
          <tr>
            <th>Name</th>
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
            <td :colspan="showActions ? 4 : 3" class="text-center py-8 text-medium-emphasis">No roles found</td>
          </tr>
          <tr v-for="role in items" :key="role.id">
            <td>{{ role.name }}</td>
            <td>{{ role.description || '—' }}</td>
            <td>
              <span class="text-caption">{{ (role.permissions || []).length }} permissions</span>
            </td>
            <td v-if="showActions" class="text-right">
              <v-btn v-if="canUpdate" size="small" variant="text" @click="openEdit(role)">Edit</v-btn>
              <v-btn v-if="canDelete" size="small" variant="text" color="error" @click="onDelete(role)">
                Delete
              </v-btn>
            </td>
          </tr>
        </tbody>
      </v-table>
    </v-card>

    <v-dialog v-model="dialog" max-width="640" persistent>
      <v-card>
        <v-card-title>{{ editing ? 'Edit role' : 'Create role' }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="form.name" label="Name" class="mb-2" />
          <v-textarea v-model="form.description" label="Description" rows="2" class="mb-4" />
          <div class="text-subtitle-2 mb-2">Permissions</div>
          <v-checkbox
            v-for="perm in permissions"
            :key="perm.id"
            v-model="form.permission_ids"
            :label="`${perm.name}${perm.description ? ' — ' + perm.description : ''}`"
            :value="perm.id"
            density="compact"
            hide-details
            color="primary"
          />
          <v-alert v-if="formError" type="error" variant="tonal" density="compact" class="mt-3">
            {{ formError }}
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" @click="dialog = false">Cancel</v-btn>
          <v-btn color="primary" :loading="saving" @click="save">Save</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>
