<script setup lang="ts">
const props = defineProps<{ id?: number }>()

const { items, fetchScopes, getScope, createScope, updateScope } = useScopesAdmin()
const { hasPermission } = useAuth()

const isEdit = computed(() => props.id !== undefined)
const canSave = computed(() => hasPermission(isEdit.value ? 'scope.update' : 'scope.create'))

const loading = ref(true)
const saving = ref(false)
const formError = ref('')
const loadError = ref('')

const form = reactive({
  code: '',
  name: '',
  description: '',
  parent_id: null as number | null,
  is_active: true,
})

const parentOptions = computed(() =>
  items.value
    .filter((s) => !isEdit.value || s.id !== props.id)
    .map((s) => ({ title: `${s.name} (${s.code})`, value: s.id })),
)

onMounted(async () => {
  try {
    await fetchScopes()
    if (isEdit.value) {
      const s = await getScope(props.id!)
      form.code = s.code
      form.name = s.name
      form.description = s.description || ''
      form.parent_id = s.parent_id ?? null
      form.is_active = s.is_active
    }
  } catch (err: unknown) {
    const e = err as { data?: { message?: string } }
    loadError.value = e?.data?.message || 'Failed to load scope'
  } finally {
    loading.value = false
  }
})

const codeLocked = computed(() => isEdit.value && form.code === 'global')

async function save() {
  formError.value = ''
  if (!form.code.trim() || !form.name.trim()) {
    formError.value = 'Code and name are required'
    return
  }
  saving.value = true
  try {
    if (isEdit.value) {
      await updateScope(props.id!, {
        code: form.code.trim(),
        name: form.name.trim(),
        description: form.description,
        parent_id: form.parent_id,
        clear_parent: !form.parent_id,
        is_active: form.is_active,
      })
    } else {
      await createScope({
        code: form.code.trim(),
        name: form.name.trim(),
        description: form.description,
        parent_id: form.parent_id,
        is_active: form.is_active,
      })
    }
    await navigateTo('/admin/scopes')
  } catch (err: unknown) {
    const e = err as { data?: { message?: string } }
    formError.value = e?.data?.message || 'Save failed'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-4" />
    <v-alert v-if="loadError" type="error" variant="tonal" class="mb-4">{{ loadError }}</v-alert>

    <v-row v-if="!loading && !loadError">
      <v-col cols="12" lg="8">
        <v-card>
          <v-card-item>
            <v-card-title class="text-subtitle-1 font-weight-bold">Scope details</v-card-title>
            <v-card-subtitle>An organisational boundary for scoped role assignments</v-card-subtitle>
          </v-card-item>
          <v-card-text>
            <v-row>
              <v-col cols="12" sm="4">
                <v-text-field
                  v-model="form.code"
                  label="Code"
                  :disabled="codeLocked"
                  :hint="codeLocked ? 'The global scope code cannot be changed' : 'Short unique identifier'"
                  persistent-hint
                />
              </v-col>
              <v-col cols="12" sm="8">
                <v-text-field v-model="form.name" label="Name" hide-details="auto" />
              </v-col>
              <v-col cols="12">
                <v-textarea v-model="form.description" label="Description" rows="3" hide-details />
              </v-col>
              <v-col cols="12">
                <v-select
                  v-model="form.parent_id"
                  :items="parentOptions"
                  label="Parent scope"
                  clearable
                  hide-details
                />
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" lg="4">
        <v-card>
          <v-card-item>
            <v-card-title class="text-subtitle-1 font-weight-bold">Status</v-card-title>
          </v-card-item>
          <v-card-text class="d-flex align-center justify-space-between">
            <div>
              <div class="text-body-2 font-weight-medium">Active</div>
              <div class="text-caption text-medium-emphasis">Inactive scopes cannot be assigned</div>
            </div>
            <v-switch v-model="form.is_active" color="primary" hide-details inset />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-alert v-if="formError" type="error" variant="tonal" density="compact" class="mt-4">
      {{ formError }}
    </v-alert>

    <AdminFormActions
      v-if="!loading && !loadError"
      cancel-to="/admin/scopes"
      :saving="saving"
      :disabled="!canSave"
      :save-label="isEdit ? 'Save changes' : 'Create scope'"
      @save="save"
    />
  </div>
</template>
