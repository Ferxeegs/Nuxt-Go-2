<script setup lang="ts">
const { user, isImpersonating, stopImpersonate } = useAuth()
const stopping = ref(false)
const stopError = ref('')

async function onStop() {
  stopError.value = ''
  stopping.value = true
  try {
    await stopImpersonate()
  } catch (err: unknown) {
    const e = err as { data?: { message?: string }; message?: string }
    stopError.value = e?.data?.message || e?.message || 'Failed to stop impersonation'
  } finally {
    stopping.value = false
  }
}
</script>

<template>
  <v-alert
    v-if="isImpersonating && user"
    type="warning"
    variant="tonal"
    density="comfortable"
    class="mb-4"
    border="start"
  >
    <div class="d-flex flex-wrap align-center justify-space-between ga-3">
      <div>
        <div class="font-weight-medium">Impersonation active</div>
        <div class="text-body-2">
          You are viewing the app as <strong>{{ user.name }}</strong>
          <span class="text-medium-emphasis"> ({{ user.email }})</span>.
          Profile and password changes are disabled until you exit.
        </div>
        <div v-if="stopError" class="text-error text-caption mt-1">{{ stopError }}</div>
      </div>
      <v-btn color="warning" variant="flat" :loading="stopping" @click="onStop">
        Stop impersonating
      </v-btn>
    </div>
  </v-alert>
</template>
