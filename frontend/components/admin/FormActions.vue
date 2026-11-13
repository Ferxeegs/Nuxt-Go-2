<script setup lang="ts">
defineProps<{ saving?: boolean; cancelTo: string; saveLabel?: string; disabled?: boolean }>()
defineEmits<{ (e: 'save'): void }>()
</script>

<template>
  <div class="form-actions">
    <div class="form-actions__inner">
      <div class="flex-grow-1"><slot /></div>
      <v-btn variant="text" :to="cancelTo" :disabled="saving">Cancel</v-btn>
      <v-btn
        color="primary"
        variant="flat"
        :loading="saving"
        :disabled="disabled"
        @click="$emit('save')"
      >
        {{ saveLabel || 'Save changes' }}
      </v-btn>
    </div>
  </div>
</template>

<style scoped>
.form-actions {
  position: sticky;
  bottom: 0;
  z-index: 5;
  margin: 24px -16px -48px;
  padding: 0 16px;
  background: rgba(var(--v-theme-background), 0.9);
  backdrop-filter: blur(8px);
  border-top: 1px solid rgb(var(--v-theme-borderColor));
}
.form-actions__inner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 0;
}
@media (min-width: 960px) {
  .form-actions {
    margin: 24px -28px -56px;
    padding: 0 28px;
  }
}
</style>
