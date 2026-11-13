<script setup lang="ts">
useHead({
  titleTemplate: (chunk) => (chunk ? `${chunk} · Company Profile` : 'Company Profile · Admin'),
})
</script>

<template>
  <v-locale-provider>
    <v-app class="app-shell">
      <LayoutAppSidebar />
      <LayoutAppHeader />
      <v-main class="app-main">
        <div class="app-content">
          <AuthImpersonationBanner />
          <slot />
        </div>
      </v-main>
    </v-app>
  </v-locale-provider>
</template>
