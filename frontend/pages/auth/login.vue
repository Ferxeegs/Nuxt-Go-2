<script setup lang="ts">
/*-For Set Blank Layout-*/
definePageMeta({
  layout: "blank",
});
useHead({ title: "Sign in" });
</script>

<template>
  <AuthShell title="Welcome back" subtitle="Sign in to continue to your dashboard.">
    <AuthLoginForm />
    <template #footer>
      New here?
      <NuxtLink to="/auth/register" class="auth-link">Create an account</NuxtLink>
    </template>
  </AuthShell>
</template>
