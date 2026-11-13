<script setup lang="ts">
useHead({ title: 'New scope' })
</script>

<template>
  <div>
    <AdminPageHeader title="New scope" subtitle="Add an organisational boundary" back-to="/admin/scopes" />
    <AdminScopeForm />
  </div>
</template>
