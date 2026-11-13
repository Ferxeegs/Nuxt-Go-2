<script setup lang="ts">
const id = Number(useRoute().params.id)
useHead({ title: 'Edit scope' })
</script>

<template>
  <div>
    <AdminPageHeader title="Edit scope" subtitle="Update scope details" back-to="/admin/scopes" />
    <AdminScopeForm :id="id" :key="id" />
  </div>
</template>
