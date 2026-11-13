<script setup lang="ts">
defineProps<{ collapsed?: boolean; light?: boolean }>()
</script>

<template>
  <NuxtLink to="/" class="app-brand" :class="{ 'app-brand--light': light }">
    <span class="app-brand__mark">
      <v-icon icon="mdi-domain" size="20" />
    </span>
    <span v-if="!collapsed" class="app-brand__text">
      <span class="app-brand__name">Company Profile</span>
      <span class="app-brand__tag">Admin Console</span>
    </span>
  </NuxtLink>
</template>
