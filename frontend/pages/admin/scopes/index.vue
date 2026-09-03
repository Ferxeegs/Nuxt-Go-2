<script setup lang="ts">
import type { AdminScope } from '~/composables/useScopesAdmin'

const {
  items,
  loading,
  error,
  search,
  fetchScopes,
  createScope,
  updateScope,
  deleteScope,
} = useScopesAdmin()
const { hasPermission } = useAuth()

const canCreate = computed(() => hasPermission('scope.create'))
const canUpdate = computed(() => hasPermission('scope.update'))
const canDelete = computed(() => hasPermission('scope.delete'))
const showActions = computed(() => canUpdate.value || canDelete.value)

const dialog = ref(false)
const saving = ref(false)
const formError = ref('')
const editing = ref<AdminScope | null>(null)

const form = reactive({
  code: '',
  name: '',
  description: '',
  parent_id: null as number | null,
  is_active: true,
})

const parentOptions = computed(() =>
  items.value
    .filter((s) => !editing.value || s.id !== editing.value.id)
    .map((s) => ({ title: `${s.code} — ${s.name}`, value: s.id })),
)

onMounted(async () => {
  await fetchScopes()
})

function openCreate() {
  editing.value = null
  form.code = ''
  form.name = ''
  form.description = ''
  form.parent_id = null
  form.is_active = true
  formError.value = ''
  dialog.value = true
}

function openEdit(scope: AdminScope) {
  editing.value = scope
  form.code = scope.code
  form.name = scope.name
  form.description = scope.description || ''
  form.parent_id = scope.parent_id ?? null
  form.is_active = scope.is_active
  formError.value = ''
  dialog.value = true
}

async function save() {
  formError.value = ''
  saving.value = true
  try {
    if (editing.value) {
      await updateScope(editing.value.id, {
        code: form.code,
        name: form.name,
        description: form.description,
        parent_id: form.parent_id,
        clear_parent: !form.parent_id,
        is_active: form.is_active,
      })
    } else {
      await createScope({
        code: form.code,
        name: form.name,
        description: form.description,
        parent_id: form.parent_id,
        is_active: form.is_active,
      })
    }
    dialog.value = false
  } catch (err: unknown) {
    const e = err as { data?: { message?: string } }
    formError.value = e?.data?.message || 'Save failed'
  } finally {
    saving.value = false
  }
}

async function onDelete(scope: AdminScope) {
  if (!confirm(`Delete scope ${scope.code}?`)) return
  try {
    await deleteScope(scope.id)
  } catch (err: unknown) {
    const e = err as { data?: { message?: string } }
    error.value = e?.data?.message || 'Delete failed'
  }
}

async function onSearch() {
  await fetchScopes()
}
</script>

<template>
  <v-container fluid>
    <div class="d-flex flex-wrap align-center justify-space-between mb-4 ga-3">
      <div>
        <h1 class="text-h5 font-weight-bold">Scopes</h1>
        <p class="text-body-2 text-medium-emphasis mb-0">Organizational boundaries for scoped role assignments</p>
      </div>
      <v-btn v-if="canCreate" color="primary" @click="openCreate">New scope</v-btn>
    </div>

    <v-card rounded="lg" elevation="1">
      <v-card-text class="d-flex flex-wrap ga-3">
        <v-text-field
          v-model="search"
          label="Search"
          density="compact"
          hide-details
          style="max-width: 280px"
          @keyup.enter="onSearch"
        />
        <v-btn variant="tonal" @click="onSearch">Search</v-btn>
      </v-card-text>

      <v-alert v-if="error" type="error" variant="tonal" class="ma-4" density="compact">{{ error }}</v-alert>

      <v-table>
        <thead>
          <tr>
            <th>Code</th>
            <th>Name</th>
            <th>Parent</th>
            <th>Active</th>
            <th v-if="showActions" class="text-right">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td :colspan="showActions ? 5 : 4" class="text-center py-8">Loading…</td>
          </tr>
          <tr v-else-if="!items.length">
            <td :colspan="showActions ? 5 : 4" class="text-center py-8 text-medium-emphasis">No scopes found</td>
          </tr>
          <tr v-for="scope in items" :key="scope.id">
            <td>{{ scope.code }}</td>
            <td>{{ scope.name }}</td>
            <td>{{ scope.parent_code || '—' }}</td>
            <td>{{ scope.is_active ? 'Yes' : 'No' }}</td>
            <td v-if="showActions" class="text-right">
              <v-btn v-if="canUpdate" size="small" variant="text" @click="openEdit(scope)">Edit</v-btn>
              <v-btn v-if="canDelete" size="small" variant="text" color="error" @click="onDelete(scope)">
                Delete
              </v-btn>
            </td>
          </tr>
        </tbody>
      </v-table>
    </v-card>

    <v-dialog v-model="dialog" max-width="560" persistent>
      <v-card>
        <v-card-title>{{ editing ? 'Edit scope' : 'Create scope' }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="form.code" label="Code" class="mb-2" :disabled="editing?.code === 'global'" />
          <v-text-field v-model="form.name" label="Name" class="mb-2" />
          <v-textarea v-model="form.description" label="Description" rows="2" class="mb-2" />
          <v-select
            v-model="form.parent_id"
            :items="parentOptions"
            label="Parent scope"
            clearable
            class="mb-2"
          />
          <v-switch v-model="form.is_active" label="Active" color="primary" />
          <v-alert v-if="formError" type="error" variant="tonal" density="compact" class="mt-2">
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
