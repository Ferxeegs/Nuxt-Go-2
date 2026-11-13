<script setup lang="ts">
import type { AssignmentDraft } from '~/composables/useUsersAdmin'

const props = defineProps<{ id?: string }>()

const { getUser, createUser, updateUser, resetUserPassword } = useUsersAdmin()
const { items: roles, fetchRoles } = useRolesAdmin()
const { items: scopes, fetchScopes } = useScopesAdmin()
const { hasPermission, isImpersonating } = useAuth()

const isEdit = computed(() => props.id !== undefined)
const canSave = computed(
  () => hasPermission(isEdit.value ? 'user.update' : 'user.create') && !isImpersonating.value,
)
const canReset = computed(() => isEdit.value && hasPermission('user.update'))

const loading = ref(true)
const saving = ref(false)
const resetting = ref(false)
const showPassword = ref(false)
const formError = ref('')
const formSuccess = ref('')
const loadError = ref('')

const form = reactive({
  name: '',
  email: '',
  password: '',
  is_active: true,
  assignments: [] as AssignmentDraft[],
})

const roleOptions = computed(() => roles.value.map((r) => ({ title: r.name, value: r.id })))
const scopeOptions = computed(() =>
  scopes.value.map((s) => ({ title: `${s.name} (${s.code})`, value: s.id })),
)
const defaultScopeId = computed(
  () => scopes.value.find((s) => s.code === 'global')?.id ?? scopes.value[0]?.id ?? null,
)

onMounted(async () => {
  try {
    await Promise.all([fetchRoles(), fetchScopes()])
    if (isEdit.value) {
      const u = await getUser(props.id!)
      form.name = u.name
      form.email = u.email
      form.is_active = u.is_active
      form.assignments = (u.assignments || []).map((a) => ({
        role_id: a.role_id,
        scope_id: a.scope_id,
      }))
    } else {
      form.assignments = [
        {
          role_id: roles.value.find((r) => r.name !== 'SuperAdmin')?.id ?? roles.value[0]?.id ?? null,
          scope_id: defaultScopeId.value,
        },
      ]
    }
    if (!form.assignments.length) {
      form.assignments = [{ role_id: null, scope_id: defaultScopeId.value }]
    }
  } catch (err: unknown) {
    const e = err as { data?: { message?: string } }
    loadError.value = e?.data?.message || 'Failed to load user'
  } finally {
    loading.value = false
  }
})

function addAssignment() {
  form.assignments.push({ role_id: null, scope_id: defaultScopeId.value })
}
function removeAssignment(i: number) {
  form.assignments.splice(i, 1)
}

async function save() {
  formError.value = ''
  formSuccess.value = ''
  const assignments = form.assignments
    .filter((a) => a.role_id && a.scope_id)
    .map((a) => ({ role_id: a.role_id as number, scope_id: a.scope_id as number }))

  if (!form.name.trim() || !form.email.trim()) {
    formError.value = 'Name and email are required'
    return
  }
  if (!assignments.length) {
    formError.value = 'At least one role assignment is required'
    return
  }
  if (!isEdit.value && !form.password) {
    formError.value = 'Password is required'
    return
  }

  saving.value = true
  try {
    if (isEdit.value) {
      await updateUser(props.id!, {
        name: form.name.trim(),
        email: form.email.trim(),
        is_active: form.is_active,
        assignments,
      })
    } else {
      await createUser({
        name: form.name.trim(),
        email: form.email.trim(),
        password: form.password,
        assignments,
      })
    }
    await navigateTo('/admin/users')
  } catch (err: unknown) {
    const e = err as { data?: { message?: string } }
    formError.value = e?.data?.message || 'Save failed'
  } finally {
    saving.value = false
  }
}

async function onResetPassword() {
  formError.value = ''
  formSuccess.value = ''
  if (
    !confirm(
      `Reset password for ${form.email} to the configured default password?\n\nAll active sessions for this user will be signed out.`,
    )
  ) {
    return
  }
  resetting.value = true
  try {
    await resetUserPassword(props.id!)
    formSuccess.value = 'Password reset to the default password. User sessions were signed out.'
  } catch (err: unknown) {
    const e = err as { data?: { message?: string }; message?: string }
    formError.value = e?.data?.message || e?.message || 'Password reset failed'
  } finally {
    resetting.value = false
  }
}
</script>

<template>
  <div>
    <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-4" />
    <v-alert v-if="loadError" type="error" variant="tonal" class="mb-4">{{ loadError }}</v-alert>

    <v-row v-if="!loading && !loadError">
      <v-col cols="12" lg="5">
        <v-card class="mb-4">
          <v-card-item>
            <v-card-title class="text-subtitle-1 font-weight-bold">Account</v-card-title>
            <v-card-subtitle>Basic information and sign-in</v-card-subtitle>
          </v-card-item>
          <v-card-text>
            <v-text-field v-model="form.name" label="Full name" class="mb-3" hide-details="auto" />
            <v-text-field
              v-model="form.email"
              label="Email"
              type="email"
              class="mb-3"
              hide-details="auto"
            />
            <v-text-field
              v-if="!isEdit"
              v-model="form.password"
              label="Password"
              :type="showPassword ? 'text' : 'password'"
              :append-inner-icon="showPassword ? 'mdi-eye-off-outline' : 'mdi-eye-outline'"
              hide-details="auto"
              autocomplete="new-password"
              @click:append-inner="showPassword = !showPassword"
            />
            <div v-if="isEdit" class="d-flex align-center justify-space-between">
              <div>
                <div class="text-body-2 font-weight-medium">Active</div>
                <div class="text-caption text-medium-emphasis">Inactive users cannot sign in</div>
              </div>
              <v-switch v-model="form.is_active" color="primary" hide-details inset />
            </div>
          </v-card-text>
        </v-card>

        <v-card v-if="canReset">
          <v-card-item>
            <v-card-title class="text-subtitle-1 font-weight-bold">Security</v-card-title>
            <v-card-subtitle>Reset password to the server default</v-card-subtitle>
          </v-card-item>
          <v-card-text>
            <p class="text-body-2 text-medium-emphasis mb-3">
              Sets the password back to <code>DEFAULT_PASSWORD</code> and signs the user out of
              every active session.
            </p>
            <v-btn
              color="warning"
              variant="tonal"
              prepend-icon="mdi-lock-reset"
              :loading="resetting"
              :disabled="saving || isImpersonating"
              @click="onResetPassword"
            >
              Reset password
            </v-btn>
            <v-alert v-if="formSuccess" type="success" variant="tonal" density="compact" class="mt-3">
              {{ formSuccess }}
            </v-alert>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" lg="7">
        <v-card>
          <v-card-item>
            <div class="d-flex align-center justify-space-between">
              <div>
                <v-card-title class="text-subtitle-1 font-weight-bold pa-0">
                  Role assignments
                </v-card-title>
                <v-card-subtitle class="pa-0">
                  Use scope <code>global</code> for organisation-wide roles
                </v-card-subtitle>
              </div>
              <v-btn size="small" variant="tonal" prepend-icon="mdi-plus" @click="addAssignment">
                Add
              </v-btn>
            </div>
          </v-card-item>
          <v-card-text class="d-flex flex-column ga-3">
            <div v-for="(row, i) in form.assignments" :key="i" class="assignment">
              <v-select
                v-model="row.role_id"
                :items="roleOptions"
                label="Role"
                density="comfortable"
                hide-details
              />
              <v-select
                v-model="row.scope_id"
                :items="scopeOptions"
                label="Scope"
                density="comfortable"
                hide-details
              />
              <v-btn
                icon="mdi-trash-can-outline"
                variant="text"
                color="error"
                size="small"
                aria-label="Remove assignment"
                :disabled="form.assignments.length <= 1"
                @click="removeAssignment(i)"
              />
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-alert v-if="formError" type="error" variant="tonal" density="compact" class="mt-4">
      {{ formError }}
    </v-alert>

    <AdminFormActions
      v-if="!loading && !loadError"
      cancel-to="/admin/users"
      :saving="saving"
      :disabled="!canSave || resetting"
      :save-label="isEdit ? 'Save changes' : 'Create user'"
      @save="save"
    />
  </div>
</template>

<style scoped>
.assignment {
  display: grid;
  grid-template-columns: 1fr 1fr auto;
  gap: 12px;
  align-items: center;
  padding: 12px;
  border: 1px solid rgb(var(--v-theme-borderColor));
  border-radius: 12px;
  background: rgb(var(--v-theme-bglight));
}
@media (max-width: 599px) {
  .assignment {
    grid-template-columns: 1fr auto;
  }
  .assignment > :nth-child(2) {
    grid-column: 1;
  }
}
</style>
