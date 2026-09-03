<script setup lang="ts">
import { inferSettingType, type SettingDraft } from '~/composables/useSettingsAdmin'

definePageMeta({
  name: 'admin-settings',
})

const { items, loading, saving, error, fetchSettings, saveSettings } = useSettingsAdmin()
const { uploadFile } = useMediaUpload()
const { hasPermission } = useAuth()

const canUpdate = computed(() => hasPermission('setting.update'))
const canUploadMedia = computed(() => hasPermission('media.upload'))

const drafts = ref<SettingDraft[]>([])
const formError = ref('')
const successMessage = ref('')
const uploadingKey = ref<string | null>(null)
const filePickers = ref<Record<number, File[] | File | null>>({})

const labelForKey: Record<string, string> = {
  site_name: 'Site name',
  logo: 'Logo',
  contact_email: 'Contact email',
  contact_phone: 'Contact phone',
  social_facebook: 'Facebook URL',
  footer_text: 'Footer text',
}

onMounted(async () => {
  await fetchSettings()
  syncDraftsFromItems()
})

function syncDraftsFromItems() {
  drafts.value = items.value.map((item) => ({
    key: item.key,
    value: item.value ?? '',
    type: inferSettingType(item.key, item.value ?? ''),
  }))
}

function settingType(row: SettingDraft) {
  return inferSettingType(row.key, row.value)
}

function settingLabel(key: string) {
  return labelForKey[key] || key
}

function draftIdentity(row: SettingDraft, index: number) {
  return row.key || `__new_${index}`
}

function addSetting() {
  drafts.value.push({
    key: '',
    value: '',
    type: 'string',
    _isNew: true,
  })
}

function removeDraft(index: number) {
  drafts.value.splice(index, 1)
  delete filePickers.value[index]
}

function clearImage(row: SettingDraft) {
  row.value = ''
}

function imagePreviewSrc(value: string) {
  if (!value) return ''
  if (value.startsWith('http://') || value.startsWith('https://') || value.startsWith('/')) {
    return value
  }
  return `/storage/${value.replace(/^\//, '')}`
}

async function onImageFile(row: SettingDraft, index: number, files: File | File[] | null) {
  const list = Array.isArray(files) ? files : files ? [files] : []
  const file = list[0]
  filePickers.value[index] = null
  if (!file) return

  formError.value = ''
  successMessage.value = ''
  uploadingKey.value = draftIdentity(row, index)

  try {
    const uploaded = await uploadFile(file, 'settings')
    row.value = uploaded.url
  } catch (err: unknown) {
    const e = err as { message?: string }
    formError.value = e?.message || 'Image upload failed'
  } finally {
    uploadingKey.value = null
  }
}

function isUploading(row: SettingDraft, index: number) {
  return uploadingKey.value === draftIdentity(row, index)
}

async function onSave() {
  formError.value = ''
  successMessage.value = ''

  const payload = drafts.value
    .map((d) => {
      const key = d.key.trim()
      const value = d.value ?? ''
      return {
        key,
        value,
        type: inferSettingType(key, value),
      }
    })
    .filter((d) => d.key !== '')

  if (!payload.length) {
    formError.value = 'At least one setting with a key is required'
    return
  }

  const keys = payload.map((p) => p.key)
  if (new Set(keys).size !== keys.length) {
    formError.value = 'Duplicate setting keys are not allowed'
    return
  }

  try {
    await saveSettings(payload)
    syncDraftsFromItems()
    successMessage.value = 'Settings saved successfully'
  } catch {
    // error already set in composable
  }
}

async function onReload() {
  formError.value = ''
  successMessage.value = ''
  await fetchSettings()
  syncDraftsFromItems()
}
</script>

<template>
  <v-container fluid>
    <div class="d-flex flex-wrap align-center justify-space-between mb-4 ga-3">
      <div>
        <h1 class="text-h5 font-weight-bold">Web Settings</h1>
        <p class="text-body-2 text-medium-emphasis mb-0">
          Site configuration. Image fields upload to MinIO and store the public URL.
        </p>
      </div>
      <div class="d-flex ga-2">
        <v-btn variant="tonal" :disabled="loading || saving" @click="onReload">Reload</v-btn>
        <v-btn
          v-if="canUpdate"
          color="primary"
          :loading="saving"
          :disabled="loading || Boolean(uploadingKey)"
          @click="onSave"
        >
          Save
        </v-btn>
      </div>
    </div>

    <v-alert v-if="error" type="error" variant="tonal" density="compact" class="mb-4">{{ error }}</v-alert>
    <v-alert v-if="formError" type="error" variant="tonal" density="compact" class="mb-4">{{ formError }}</v-alert>
    <v-alert v-if="successMessage" type="success" variant="tonal" density="compact" class="mb-4">
      {{ successMessage }}
    </v-alert>

    <div v-if="loading" class="text-center py-10 text-medium-emphasis">Loading…</div>

    <template v-else>
      <div v-if="!drafts.length" class="text-center py-8 text-medium-emphasis mb-4">
        No settings yet. Add one below.
      </div>

      <v-row>
        <v-col
          v-for="(row, index) in drafts"
          :key="row._isNew ? `new-${index}` : row.key"
          cols="12"
          :md="settingType(row) === 'image' || settingType(row) === 'text' || settingType(row) === 'json' ? 12 : 6"
        >
          <v-card rounded="lg" elevation="1" class="h-100">
            <v-card-text>
              <div class="d-flex flex-wrap align-center justify-space-between mb-3 ga-2">
                <div>
                  <div class="text-subtitle-1 font-weight-medium">
                    {{ row._isNew ? 'New setting' : settingLabel(row.key) }}
                  </div>
                  <div v-if="!row._isNew" class="text-caption text-medium-emphasis">{{ row.key }}</div>
                </div>
                <v-btn
                  v-if="canUpdate"
                  size="small"
                  variant="text"
                  color="error"
                  @click="removeDraft(index)"
                >
                  Remove
                </v-btn>
              </div>

              <v-row dense>
                <v-col v-if="row._isNew" cols="12">
                  <v-text-field
                    v-model="row.key"
                    label="Key"
                    density="comfortable"
                    hide-details="auto"
                    :disabled="!canUpdate"
                    hint="Type is inferred from the key (e.g. logo → image, footer_text → string)"
                    persistent-hint
                  />
                </v-col>

                <v-col cols="12">
                  <template v-if="settingType(row) === 'image'">
                    <div class="d-flex flex-wrap ga-4 align-start">
                      <div
                        class="d-flex align-center justify-center rounded-lg overflow-hidden bg-grey-lighten-4"
                        style="width: 140px; height: 140px; border: 1px dashed rgba(0,0,0,0.2)"
                      >
                        <img
                          v-if="row.value"
                          :src="imagePreviewSrc(row.value)"
                          alt="Preview"
                          style="max-width: 100%; max-height: 100%; object-fit: contain"
                        />
                        <span v-else class="text-caption text-medium-emphasis px-2 text-center">
                          No image
                        </span>
                      </div>

                      <div class="flex-grow-1" style="min-width: 220px">
                        <v-file-input
                          v-if="canUpdate && canUploadMedia"
                          v-model="filePickers[index]"
                          label="Upload image"
                          accept="image/png,image/jpeg,image/webp,.png,.jpg,.jpeg,.webp"
                          prepend-icon=""
                          density="comfortable"
                          hide-details="auto"
                          show-size
                          clearable
                          :loading="isUploading(row, index)"
                          :disabled="isUploading(row, index)"
                          @update:model-value="(files) => onImageFile(row, index, files as File | File[] | null)"
                        />
                        <v-alert
                          v-else-if="canUpdate && !canUploadMedia"
                          type="info"
                          variant="tonal"
                          density="compact"
                          class="mb-3"
                        >
                          Image upload requires <code>media.upload</code> permission. You can still paste a URL below.
                        </v-alert>
                        <v-text-field
                          v-model="row.value"
                          :class="canUpdate && canUploadMedia ? 'mt-3' : ''"
                          label="Image URL"
                          density="comfortable"
                          hide-details="auto"
                          :disabled="!canUpdate"
                          hint="Filled automatically after upload to MinIO"
                          persistent-hint
                        />
                        <v-btn
                          v-if="canUpdate && row.value"
                          class="mt-2"
                          size="small"
                          variant="text"
                          color="error"
                          @click="clearImage(row)"
                        >
                          Clear image
                        </v-btn>
                      </div>
                    </div>
                  </template>

                  <v-textarea
                    v-else-if="settingType(row) === 'text' || settingType(row) === 'json'"
                    v-model="row.value"
                    :label="settingType(row) === 'json' ? 'JSON value' : 'Value'"
                    density="comfortable"
                    rows="4"
                    hide-details="auto"
                    auto-grow
                    :disabled="!canUpdate"
                  />

                  <v-text-field
                    v-else
                    v-model="row.value"
                    label="Value"
                    density="comfortable"
                    hide-details="auto"
                    :disabled="!canUpdate"
                  />
                </v-col>
              </v-row>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-btn v-if="canUpdate" class="mt-4" variant="tonal" @click="addSetting">Add setting</v-btn>
    </template>
  </v-container>
</template>
