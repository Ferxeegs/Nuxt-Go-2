<script setup lang="ts">
withDefaults(defineProps<{ title?: string; subtitle?: string; maxWidth?: number }>(), {
  maxWidth: 440,
})
</script>

<template>
  <div class="auth-shell">
    <aside class="auth-shell__aside d-none d-lg-flex">
      <LayoutAppBrand light />
      <div class="auth-shell__pitch">
        <h2 class="auth-shell__headline">Manage your company profile in one place.</h2>
        <p class="auth-shell__lead">
          Content, users, roles and site settings, with fine-grained access control built in.
        </p>
        <ul class="auth-shell__points">
          <li><v-icon icon="mdi-check-circle" size="18" /> Role and scope based permissions</li>
          <li><v-icon icon="mdi-check-circle" size="18" /> Media library with public assets</li>
          <li><v-icon icon="mdi-check-circle" size="18" /> Secure sessions with role switching</li>
        </ul>
      </div>
      <div class="auth-shell__foot">&copy; {{ new Date().getFullYear() }} Company Profile</div>
    </aside>

    <main class="auth-shell__main">
      <div class="auth-shell__card" :style="{ maxWidth: maxWidth + 'px' }">
        <div class="d-lg-none mb-8"><LayoutAppBrand /></div>
        <h1 v-if="title" class="auth-shell__title">{{ title }}</h1>
        <p v-if="subtitle" class="auth-shell__subtitle">{{ subtitle }}</p>
        <slot />
        <div v-if="$slots.footer" class="auth-shell__footer"><slot name="footer" /></div>
      </div>
    </main>
  </div>
</template>
