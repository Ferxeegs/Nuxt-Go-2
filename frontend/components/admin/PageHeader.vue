<script setup lang="ts">
defineProps<{ title: string; subtitle?: string; backTo?: string }>()
</script>

<template>
  <div class="page-header">
    <div class="d-flex align-center ga-3 min-w-0">
      <v-btn
        v-if="backTo"
        :to="backTo"
        icon="mdi-arrow-left"
        variant="tonal"
        size="small"
        aria-label="Back"
      />
      <div class="min-w-0">
        <h1 class="page-header__title">{{ title }}</h1>
        <p v-if="subtitle" class="page-header__subtitle">{{ subtitle }}</p>
      </div>
    </div>
    <div class="d-flex align-center ga-2 flex-wrap">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 20px;
}
.page-header__title {
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: -0.015em;
  line-height: 1.3;
}
.page-header__subtitle {
  font-size: 0.875rem;
  margin: 2px 0 0;
  color: rgb(var(--v-theme-grey100));
}
.min-w-0 {
  min-width: 0;
}
</style>
