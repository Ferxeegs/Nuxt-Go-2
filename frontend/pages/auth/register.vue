<script setup lang="ts">
/*-For Set Blank Layout-*/
definePageMeta({
  layout: "blank",
});
useHead({ title: "Create account" });
</script>

<template>
  <AuthShell title="Create your account" subtitle="It only takes a minute to get started.">
    <AuthRegisterForm />
    <template #footer>
      Already have an account?
      <NuxtLink to="/auth/login" class="auth-link">Sign in</NuxtLink>
    </template>
  </AuthShell>
</template>
