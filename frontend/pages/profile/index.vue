<script setup lang="ts">
definePageMeta({
  name: 'profile',
})

const {
  user,
  fetchMe,
  updateProfile,
  changePassword,
  isImpersonating,
  hasPermission,
} = useAuth()
const { uploadFile } = useMediaUpload()

const canUploadMedia = computed(() => hasPermission('media.upload'))
const readOnly = computed(() => isImpersonating.value)

const profileForm = reactive({
  name: '',
  email: '',
  avatar: '',
})
const passwordForm = reactive({
  current_password: '',
  new_password: '',
  confirm_password: '',
})

const profileSaving = ref(false)
const passwordSaving = ref(false)
const avatarUploading = ref(false)
const avatarPicker = ref<File[] | File | null>(null)

const profileError = ref('')
const profileSuccess = ref('')
const passwordError = ref('')
const passwordSuccess = ref('')

const activeRoleLabel = computed(() => {
  if (!user.value) return '—'
  const role = user.value.active_role_name || user.value.role || '—'
  const scope = user.value.active_scope_name
  return scope ? `${role} · ${scope}` : role
})

function formatDateTime(value?: string | null) {
  if (!value) return '—'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  return d.toLocaleString()
}

const avatarSrc = computed(() => {
  const value = profileForm.avatar || user.value?.avatar || ''
  if (!value) return '/images/profile/user-1.jpg'
  if (value.startsWith('http://') || value.startsWith('https://') || value.startsWith('/')) {
    return value
  }
  return `/storage/${value.replace(/^\//, '')}`
})

onMounted(async () => {
  if (!user.value) {
    await fetchMe()
  }
  syncFromUser()
})

watch(
  () => user.value,
  () => syncFromUser(),
)

function syncFromUser() {
  if (!user.value) return
  profileForm.name = user.value.name || ''
  profileForm.email = user.value.email || ''
  profileForm.avatar = user.value.avatar || ''
}

async function onAvatarSelected(files: File | File[] | null) {
  const list = Array.isArray(files) ? files : files ? [files] : []
  const file = list[0]
  avatarPicker.value = null
  if (!file || readOnly.value) return

  profileError.value = ''
  profileSuccess.value = ''
  avatarUploading.value = true
  try {
    const uploaded = await uploadFile(file, 'avatars')
    profileForm.avatar = uploaded.url
  } catch (err: unknown) {
    const e = err as { message?: string }
    profileError.value = e?.message || 'Avatar upload failed'
  } finally {
    avatarUploading.value = false
  }
}

async function onSaveProfile() {
  profileError.value = ''
  profileSuccess.value = ''

  if (readOnly.value) {
    profileError.value = 'Profile cannot be edited while impersonating'
    return
  }
  if (profileForm.name.trim().length < 2) {
    profileError.value = 'Name must be at least 2 characters'
    return
  }
  if (!profileForm.email.trim()) {
    profileError.value = 'Email is required'
    return
  }

  profileSaving.value = true
  try {
    await updateProfile({
      name: profileForm.name.trim(),
      email: profileForm.email.trim(),
      avatar: profileForm.avatar.trim(),
    })
    profileSuccess.value = 'Profile updated successfully'
  } catch (err: unknown) {
    const e = err as { data?: { message?: string }; message?: string }
    profileError.value = e?.data?.message || e?.message || 'Failed to update profile'
  } finally {
    profileSaving.value = false
  }
}

async function onChangePassword() {
  passwordError.value = ''
  passwordSuccess.value = ''

  if (readOnly.value) {
    passwordError.value = 'Password cannot be changed while impersonating'
    return
  }
  if (passwordForm.current_password.length < 6) {
    passwordError.value = 'Current password must be at least 6 characters'
    return
  }
  if (passwordForm.new_password.length < 6) {
    passwordError.value = 'New password must be at least 6 characters'
    return
  }
  if (passwordForm.new_password !== passwordForm.confirm_password) {
    passwordError.value = 'New password and confirmation do not match'
    return
  }
  if (passwordForm.new_password === passwordForm.current_password) {
    passwordError.value = 'New password must be different from the current password'
    return
  }

  passwordSaving.value = true
  try {
    await changePassword({
      current_password: passwordForm.current_password,
      new_password: passwordForm.new_password,
    })
    passwordForm.current_password = ''
    passwordForm.new_password = ''
    passwordForm.confirm_password = ''
    passwordSuccess.value = 'Password updated successfully. Other sessions were signed out.'
  } catch (err: unknown) {
    const e = err as { data?: { message?: string }; message?: string }
    passwordError.value = e?.data?.message || e?.message || 'Failed to change password'
  } finally {
    passwordSaving.value = false
  }
}
</script>

<template>
  <v-container fluid>
    <div class="mb-4">
      <h1 class="text-h5 font-weight-bold">My Profile</h1>
      <p class="text-body-2 text-medium-emphasis mb-0">
        Manage your account details and password
      </p>
    </div>

    <v-alert
      v-if="isImpersonating"
      type="warning"
      variant="tonal"
      density="compact"
      class="mb-4"
    >
      You are impersonating another user. Profile and password changes are disabled.
    </v-alert>

    <v-row>
      <v-col cols="12" md="7">
        <v-card rounded="lg" elevation="1">
          <v-card-title class="text-subtitle-1 font-weight-medium">Profile information</v-card-title>
          <v-card-text>
            <v-alert
              v-if="profileError"
              type="error"
              variant="tonal"
              density="compact"
              class="mb-4"
            >
              {{ profileError }}
            </v-alert>
            <v-alert
              v-if="profileSuccess"
              type="success"
              variant="tonal"
              density="compact"
              class="mb-4"
            >
              {{ profileSuccess }}
            </v-alert>

            <div class="d-flex flex-wrap ga-4 align-start mb-5">
              <v-avatar size="96" rounded="lg">
                <img :src="avatarSrc" alt="Avatar" width="96" height="96" style="object-fit: cover" />
              </v-avatar>
              <div class="flex-grow-1" style="min-width: 220px">
                <v-file-input
                  v-if="!readOnly && canUploadMedia"
                  v-model="avatarPicker"
                  label="Upload avatar"
                  accept="image/png,image/jpeg,image/webp,.png,.jpg,.jpeg,.webp"
                  prepend-icon=""
                  density="comfortable"
                  hide-details="auto"
                  show-size
                  clearable
                  :loading="avatarUploading"
                  :disabled="avatarUploading"
                  @update:model-value="(files) => onAvatarSelected(files as File | File[] | null)"
                />
                <v-text-field
                  v-model="profileForm.avatar"
                  class="mt-3"
                  label="Avatar URL"
                  density="comfortable"
                  hide-details="auto"
                  :disabled="readOnly"
                  hint="Optional image URL (RustFS upload fills this automatically)"
                  persistent-hint
                />
              </div>
            </div>

            <v-text-field
              v-model="profileForm.name"
              label="Full name"
              density="comfortable"
              hide-details="auto"
              class="mb-3"
              :disabled="readOnly"
            />
            <v-text-field
              v-model="profileForm.email"
              label="Email"
              type="email"
              density="comfortable"
              hide-details="auto"
              class="mb-3"
              :disabled="readOnly"
            />

            <v-row dense class="mb-2">
              <v-col cols="12" sm="6">
                <div class="text-caption text-medium-emphasis">Active role</div>
                <div class="text-body-2">{{ activeRoleLabel }}</div>
              </v-col>
              <v-col cols="12" sm="6">
                <div class="text-caption text-medium-emphasis">Status</div>
                <div class="text-body-2">{{ user?.is_active ? 'Active' : 'Inactive' }}</div>
              </v-col>
              <v-col v-if="user?.last_login_at" cols="12" sm="6">
                <div class="text-caption text-medium-emphasis">Last login</div>
                <div class="text-body-2">{{ formatDateTime(user.last_login_at) }}</div>
              </v-col>
              <v-col v-if="user?.created_at" cols="12" sm="6">
                <div class="text-caption text-medium-emphasis">Member since</div>
                <div class="text-body-2">{{ formatDateTime(user.created_at) }}</div>
              </v-col>
            </v-row>
          </v-card-text>
          <v-card-actions class="px-4 pb-4">
            <v-spacer />
            <v-btn
              color="primary"
              :loading="profileSaving"
              :disabled="readOnly || avatarUploading"
              @click="onSaveProfile"
            >
              Save profile
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-col>

      <v-col cols="12" md="5">
        <v-card rounded="lg" elevation="1">
          <v-card-title class="text-subtitle-1 font-weight-medium">Change password</v-card-title>
          <v-card-text>
            <v-alert
              v-if="passwordError"
              type="error"
              variant="tonal"
              density="compact"
              class="mb-4"
            >
              {{ passwordError }}
            </v-alert>
            <v-alert
              v-if="passwordSuccess"
              type="success"
              variant="tonal"
              density="compact"
              class="mb-4"
            >
              {{ passwordSuccess }}
            </v-alert>

            <v-text-field
              v-model="passwordForm.current_password"
              label="Current password"
              type="password"
              autocomplete="current-password"
              density="comfortable"
              hide-details="auto"
              class="mb-3"
              :disabled="readOnly"
            />
            <v-text-field
              v-model="passwordForm.new_password"
              label="New password"
              type="password"
              autocomplete="new-password"
              density="comfortable"
              hide-details="auto"
              class="mb-3"
              :disabled="readOnly"
              hint="Minimum 6 characters"
              persistent-hint
            />
            <v-text-field
              v-model="passwordForm.confirm_password"
              label="Confirm new password"
              type="password"
              autocomplete="new-password"
              density="comfortable"
              hide-details="auto"
              :disabled="readOnly"
            />
          </v-card-text>
          <v-card-actions class="px-4 pb-4">
            <v-spacer />
            <v-btn
              color="primary"
              variant="tonal"
              :loading="passwordSaving"
              :disabled="readOnly"
              @click="onChangePassword"
            >
              Update password
            </v-btn>
          </v-card-actions>
        </v-card>

        <v-card v-if="(user?.assignments || []).length" rounded="lg" elevation="1" class="mt-4">
          <v-card-title class="text-subtitle-1 font-weight-medium">Role assignments</v-card-title>
          <v-card-text class="pt-0">
            <v-list density="compact" class="bg-transparent">
              <v-list-item
                v-for="(a, i) in user?.assignments || []"
                :key="i"
                :title="a.role_name"
                :subtitle="`${a.scope_name} (${a.scope_code})`"
              >
                <template v-if="a.role_id === user?.active_role_id && a.scope_id === user?.active_scope_id" #append>
                  <v-chip size="small" color="primary" variant="tonal">Active</v-chip>
                </template>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>
