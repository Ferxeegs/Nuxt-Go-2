<script setup lang="ts">
import type { AdminUser, AssignmentDraft } from '~/composables/useUsersAdmin'
import type { AdminRole } from '~/composables/useRolesAdmin'
import type { AdminScope } from '~/composables/useScopesAdmin'

const {
  items,
  loading,
  error,
  search,
  fetchUsers,
  createUser,
  updateUser,
  deleteUser,
  resetUserPassword,
} = useUsersAdmin()
const { items: roles, fetchRoles } = useRolesAdmin()
const { items: scopes, fetchScopes } = useScopesAdmin()
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

const dialog = ref(false)
const saving = ref(false)
const resettingPassword = ref(false)
const formError = ref('')
const formSuccess = ref('')
const editing = ref<AdminUser | null>(null)
const impersonatingId = ref<string | null>(null)

const form = reactive({
  name: '',
  email: '',
  password: '',
  is_active: true,
  assignments: [] as AssignmentDraft[],
})

const roleOptions = computed(() =>
  roles.value.map((r: AdminRole) => ({ title: r.name, value: r.id })),
)
const scopeOptions = computed(() =>
  scopes.value.map((s: AdminScope) => ({ title: `${s.code} — ${s.name}`, value: s.id })),
)

const defaultScopeId = computed(() => {
  const global = scopes.value.find((s) => s.code === 'global')
  return global?.id ?? scopes.value[0]?.id ?? null
})

onMounted(async () => {
  await Promise.all([fetchUsers(), fetchRoles(), fetchScopes()])
})

function openCreate() {
  editing.value = null
  form.name = ''
  form.email = ''
  form.password = ''
  form.is_active = true
  form.assignments = [
    {
      role_id: roles.value.find((r) => r.name !== 'SuperAdmin')?.id ?? roles.value[0]?.id ?? null,
      scope_id: defaultScopeId.value,
    },
  ]
  formError.value = ''
  formSuccess.value = ''
  dialog.value = true
}

function openEdit(user: AdminUser) {
  editing.value = user
  form.name = user.name
  form.email = user.email
  form.password = ''
  form.is_active = user.is_active
  form.assignments = (user.assignments || []).map((a) => ({
    role_id: a.role_id,
    scope_id: a.scope_id,
  }))
  if (!form.assignments.length) {
    form.assignments = [{ role_id: null, scope_id: defaultScopeId.value }]
  }
  formError.value = ''
  formSuccess.value = ''
  dialog.value = true
}

function addAssignment() {
  form.assignments.push({ role_id: null, scope_id: defaultScopeId.value })
}

function removeAssignment(index: number) {
  form.assignments.splice(index, 1)
}

async function save() {
  formError.value = ''
  formSuccess.value = ''
  const assignments = form.assignments
    .filter((a) => a.role_id && a.scope_id)
    .map((a) => ({ role_id: a.role_id as number, scope_id: a.scope_id as number }))

  if (!assignments.length) {
    formError.value = 'At least one role assignment is required'
    return
  }

  saving.value = true
  try {
    if (editing.value) {
      await updateUser(editing.value.id, {
        name: form.name,
        email: form.email,
        is_active: form.is_active,
        assignments,
      })
    } else {
      if (!form.password) {
        formError.value = 'Password is required'
        return
      }
      await createUser({
        name: form.name,
        email: form.email,
        password: form.password,
        assignments,
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

async function onResetPassword() {
  if (!editing.value || !canUpdate.value) return
  formError.value = ''
  formSuccess.value = ''
  if (
    !confirm(
      `Reset password for ${editing.value.email} to the configured default password?\n\nAll active sessions for this user will be signed out.`,
    )
  ) {
    return
  }

  resettingPassword.value = true
  try {
    await resetUserPassword(editing.value.id)
    formSuccess.value = 'Password reset to the default password. User sessions were signed out.'
  } catch (err: unknown) {
    const e = err as { data?: { message?: string }; message?: string }
    formError.value = e?.data?.message || e?.message || 'Password reset failed'
  } finally {
    resettingPassword.value = false
  }
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
  if (!confirm(`Impersonate ${target.name} (${target.email})?\n\nYou will see the app as this user until you stop impersonating.`)) {
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

async function onSearch() {
  await fetchUsers()
}
</script>

<template>
  <v-container fluid>
    <div class="d-flex flex-wrap align-center justify-space-between mb-4 ga-3">
      <div>
        <h1 class="text-h5 font-weight-bold">Users</h1>
        <p class="text-body-2 text-medium-emphasis mb-0">
          Manage accounts and role assignments (use scope <code>global</code> for org-wide roles)
        </p>
      </div>
      <v-btn v-if="canCreate" color="primary" @click="openCreate">New user</v-btn>
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
            <th>Name</th>
            <th>Email</th>
            <th>Roles</th>
            <th>Assignments</th>
            <th>Active</th>
            <th v-if="showActions" class="text-right">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td :colspan="showActions ? 6 : 5" class="text-center py-8">Loading…</td>
          </tr>
          <tr v-else-if="!items.length">
            <td :colspan="showActions ? 6 : 5" class="text-center py-8 text-medium-emphasis">No users found</td>
          </tr>
          <tr v-for="user in items" :key="user.id">
            <td>{{ user.name }}</td>
            <td>{{ user.email }}</td>
            <td>
              <span v-if="!(user.roles || []).length" class="text-medium-emphasis">—</span>
              <span v-else>{{ user.roles.map((r) => r.name).join(', ') }}</span>
            </td>
            <td>
              <span v-if="!(user.assignments || []).length" class="text-medium-emphasis">—</span>
              <div v-else class="d-flex flex-column">
                <span v-for="(a, i) in user.assignments" :key="i" class="text-caption">
                  {{ a.role_name }} @ {{ a.scope_code || a.scope_name }}
                </span>
              </div>
            </td>
            <td>{{ user.is_active ? 'Yes' : 'No' }}</td>
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
                :disabled="isImpersonating"
                @click="openEdit(user)"
              >
                Edit
              </v-btn>
              <v-btn
                v-if="canDelete"
                size="small"
                variant="text"
                color="error"
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

    <v-dialog v-model="dialog" max-width="640" persistent>
      <v-card>
        <v-card-title>{{ editing ? 'Edit user' : 'Create user' }}</v-card-title>
        <v-card-text>
          <v-text-field v-model="form.name" label="Name" class="mb-2" />
          <v-text-field v-model="form.email" label="Email" type="email" class="mb-2" />
          <v-text-field
            v-if="!editing"
            v-model="form.password"
            label="Password"
            type="password"
            class="mb-2"
          />
          <v-switch v-if="editing" v-model="form.is_active" label="Active" color="primary" class="mb-4" />

          <template v-if="editing && canUpdate">
            <v-divider class="mb-4" />
            <div class="mb-4">
              <div class="text-subtitle-2 mb-1">Reset password</div>
              <p class="text-body-2 text-medium-emphasis mb-3">
                Set this user's password back to the server
                <code>DEFAULT_PASSWORD</code>
                and sign out their other sessions.
              </p>
              <v-btn
                color="warning"
                variant="tonal"
                :loading="resettingPassword"
                :disabled="saving || isImpersonating"
                @click="onResetPassword"
              >
                Reset password
              </v-btn>
            </div>
            <v-divider class="mb-4" />
          </template>

          <div class="d-flex align-center justify-space-between mb-2">
            <div class="text-subtitle-2">Role assignments</div>
            <v-btn size="small" variant="tonal" @click="addAssignment">Add</v-btn>
          </div>
          <div v-for="(row, index) in form.assignments" :key="index" class="d-flex ga-2 mb-2 align-center">
            <v-select v-model="row.role_id" :items="roleOptions" label="Role" density="compact" hide-details />
            <v-select v-model="row.scope_id" :items="scopeOptions" label="Scope" density="compact" hide-details />
            <v-btn
              size="small"
              variant="text"
              color="error"
              :disabled="form.assignments.length <= 1"
              @click="removeAssignment(index)"
            >
              Remove
            </v-btn>
          </div>
          <v-alert v-if="formSuccess" type="success" variant="tonal" density="compact" class="mt-2">
            {{ formSuccess }}
          </v-alert>
          <v-alert v-if="formError" type="error" variant="tonal" density="compact" class="mt-2">
            {{ formError }}
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" :disabled="resettingPassword" @click="dialog = false">Cancel</v-btn>
          <v-btn color="primary" :loading="saving" :disabled="resettingPassword" @click="save">Save</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>
