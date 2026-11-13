<script setup lang="ts">
import type { AdminRole, PermissionOption } from '~/composables/useRolesAdmin'

const props = defineProps<{ id?: number }>()

const { permissions, fetchPermissions, getRole, createRole, updateRole } = useRolesAdmin()
const { hasPermission } = useAuth()

const isEdit = computed(() => props.id !== undefined)
const canSave = computed(() => hasPermission(isEdit.value ? 'role.update' : 'role.create'))

const loading = ref(true)
const saving = ref(false)
const formError = ref('')
const loadError = ref('')

const form = reactive({
  name: '',
  description: '',
  permission_ids: [] as number[],
})

// ---- permission grouping ---------------------------------------------------
const STD_ACTIONS = ['view', 'create', 'update', 'delete']
const GROUP_META: Record<string, { label: string; icon: string }> = {
  user: { label: 'Users', icon: 'mdi-account-group-outline' },
  role: { label: 'Roles', icon: 'mdi-shield-account-outline' },
  scope: { label: 'Scopes', icon: 'mdi-map-marker-radius-outline' },
  setting: { label: 'Web settings', icon: 'mdi-cog-outline' },
  media: { label: 'Media', icon: 'mdi-image-multiple-outline' },
  activity: { label: 'Activity logs', icon: 'mdi-history' },
  invoice: { label: 'Invoices', icon: 'mdi-receipt-text-outline' },
}

type PermItem = PermissionOption & { action: string }
type Group = {
  key: string
  label: string
  icon: string
  items: PermItem[]
  byAction: Record<string, PermItem | undefined>
  others: PermItem[]
}

const groups = computed<Group[]>(() => {
  const map = new Map<string, PermItem[]>()
  for (const p of permissions.value) {
    const [key, ...rest] = p.name.split('.')
    const action = rest.join('.') || p.name
    if (!map.has(key!)) map.set(key!, [])
    map.get(key!)!.push({ ...p, action })
  }
  const out: Group[] = []
  for (const [key, items] of map) {
    items.sort((x, y) => x.action.localeCompare(y.action))
    const meta = GROUP_META[key!] || {
      label: key!.charAt(0).toUpperCase() + key!.slice(1),
      icon: 'mdi-key-outline',
    }
    const byAction: Record<string, PermItem | undefined> = {}
    for (const i of items) if (STD_ACTIONS.includes(i.action)) byAction[i.action] = i
    out.push({
      key: key!,
      ...meta,
      items,
      byAction,
      others: items.filter((i) => !STD_ACTIONS.includes(i.action)),
    })
  }
  return out
})

const selected = computed(() => new Set(form.permission_ids))
const total = computed(() => permissions.value.length)

function groupState(g: Group) {
  const n = g.items.filter((i) => selected.value.has(i.id)).length
  return { n, all: n === g.items.length, some: n > 0 && n < g.items.length }
}

function toggle(id: number, on: boolean) {
  const set = new Set(form.permission_ids)
  if (on) set.add(id)
  else set.delete(id)
  form.permission_ids = [...set]
}

function toggleGroup(g: Group, on: boolean) {
  const set = new Set(form.permission_ids)
  for (const i of g.items) (on ? set.add(i.id) : set.delete(i.id))
  form.permission_ids = [...set]
}

function selectAll() {
  form.permission_ids = permissions.value.map((p) => p.id)
}
function clearAll() {
  form.permission_ids = []
}
function readOnlyPreset() {
  form.permission_ids = permissions.value.filter((p) => p.name.endsWith('.view')).map((p) => p.id)
}

// ---- load / save -----------------------------------------------------------
onMounted(async () => {
  try {
    await fetchPermissions()
    if (isEdit.value) {
      const role: AdminRole = await getRole(props.id!)
      form.name = role.name
      form.description = role.description || ''
      const names = new Set(role.permissions || [])
      form.permission_ids = permissions.value.filter((p) => names.has(p.name)).map((p) => p.id)
    }
  } catch (err: unknown) {
    const e = err as { data?: { message?: string } }
    loadError.value = e?.data?.message || 'Failed to load role'
  } finally {
    loading.value = false
  }
})

async function save() {
  formError.value = ''
  if (!form.name.trim()) {
    formError.value = 'Role name is required'
    return
  }
  saving.value = true
  try {
    const payload = {
      name: form.name.trim(),
      description: form.description,
      permission_ids: form.permission_ids,
    }
    if (isEdit.value) await updateRole(props.id!, payload)
    else await createRole(payload)
    await navigateTo('/admin/roles')
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
      <v-col cols="12" lg="4">
        <v-card>
          <v-card-item>
            <v-card-title class="text-subtitle-1 font-weight-bold">Role details</v-card-title>
            <v-card-subtitle>Name and purpose of this role</v-card-subtitle>
          </v-card-item>
          <v-card-text>
            <v-text-field v-model="form.name" label="Name" class="mb-3" hide-details="auto" />
            <v-textarea v-model="form.description" label="Description" rows="3" hide-details />
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" lg="8">
        <v-card>
          <v-card-item>
            <div class="d-flex flex-wrap align-center justify-space-between ga-3">
              <div>
                <v-card-title class="text-subtitle-1 font-weight-bold pa-0">Permissions</v-card-title>
                <v-card-subtitle class="pa-0">
                  <strong>{{ form.permission_ids.length }}</strong> of {{ total }} enabled. Tick what this
                  role is allowed to do in each module.
                </v-card-subtitle>
              </div>
              <div class="d-flex flex-wrap ga-2">
                <v-btn size="small" variant="tonal" @click="selectAll">Allow all</v-btn>
                <v-btn size="small" variant="tonal" @click="readOnlyPreset">View only</v-btn>
                <v-btn size="small" variant="text" @click="clearAll">Clear</v-btn>
              </div>
            </div>
          </v-card-item>
          <v-divider />

          <div class="perm-scroll">
            <table class="perm-table">
              <thead>
                <tr>
                  <th class="perm-table__module">Module</th>
                  <th v-for="a in STD_ACTIONS" :key="a">{{ a }}</th>
                  <th class="perm-table__other">Other</th>
                  <th>All</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="g in groups" :key="g.key" :class="{ 'is-on': groupState(g).n > 0 }">
                  <td class="perm-table__module">
                    <div class="d-flex align-center ga-3">
                      <v-icon :icon="g.icon" size="20" class="text-medium-emphasis" />
                      <div>
                        <div class="font-weight-medium">{{ g.label }}</div>
                        <div class="text-caption text-medium-emphasis">
                          {{ groupState(g).n }} of {{ g.items.length }} enabled
                        </div>
                      </div>
                    </div>
                  </td>

                  <td v-for="a in STD_ACTIONS" :key="a" class="text-center">
                    <v-checkbox-btn
                      v-if="g.byAction[a]"
                      :model-value="selected.has(g.byAction[a]!.id)"
                      color="primary"
                      density="comfortable"
                      :aria-label="`${g.label} ${a}`"
                      @update:model-value="(v: boolean | null) => toggle(g.byAction[a]!.id, Boolean(v))"
                    />
                    <span v-else class="text-disabled">&ndash;</span>
                  </td>

                  <td class="perm-table__other">
                    <div v-if="g.others.length" class="d-flex flex-wrap ga-1 justify-start">
                      <v-chip
                        v-for="o in g.others"
                        :key="o.id"
                        size="small"
                        :variant="selected.has(o.id) ? 'flat' : 'outlined'"
                        :color="selected.has(o.id) ? 'primary' : undefined"
                        :prepend-icon="selected.has(o.id) ? 'mdi-check' : undefined"
                        class="text-capitalize"
                        @click="toggle(o.id, !selected.has(o.id))"
                      >
                        {{ o.action }}
                      </v-chip>
                    </div>
                    <span v-else class="text-disabled">&ndash;</span>
                  </td>

                  <td class="text-center">
                    <v-checkbox-btn
                      :model-value="groupState(g).all"
                      :indeterminate="groupState(g).some"
                      color="primary"
                      density="comfortable"
                      :aria-label="`Allow everything in ${g.label}`"
                      @update:model-value="(v: boolean | null) => toggleGroup(g, Boolean(v))"
                    />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </v-card>
      </v-col>
    </v-row>

    <v-alert v-if="formError" type="error" variant="tonal" density="compact" class="mt-4">
      {{ formError }}
    </v-alert>

    <AdminFormActions
      v-if="!loading && !loadError"
      cancel-to="/admin/roles"
      :saving="saving"
      :disabled="!canSave"
      :save-label="isEdit ? 'Save changes' : 'Create role'"
      @save="save"
    />
  </div>
</template>

<style scoped>
.perm-scroll {
  overflow-x: auto;
}
.perm-table {
  width: 100%;
  min-width: 640px;
  border-collapse: collapse;
}
.perm-table th {
  padding: 12px 8px;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  text-align: center;
  color: rgb(var(--v-theme-grey100));
  background: rgb(var(--v-theme-bglight));
  border-bottom: 1px solid rgb(var(--v-theme-borderColor));
}
.perm-table td {
  padding: 8px;
  border-bottom: 1px solid rgb(var(--v-theme-borderColor));
  vertical-align: middle;
}
.perm-table tbody tr:last-child td {
  border-bottom: 0;
}
.perm-table tbody tr:hover {
  background: rgb(var(--v-theme-hoverColor));
}
.perm-table tbody tr.is-on {
  background: rgba(var(--v-theme-primary), 0.035);
}
.perm-table__module {
  text-align: left !important;
  padding-left: 20px !important;
  min-width: 190px;
}
.perm-table__other {
  text-align: left !important;
  min-width: 140px;
}
.perm-table td.text-center :deep(.v-selection-control) {
  justify-content: center;
}
</style>
