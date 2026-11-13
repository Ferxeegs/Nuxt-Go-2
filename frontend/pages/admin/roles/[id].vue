<script setup lang="ts">
const id = Number(useRoute().params.id)
useHead({ title: 'Edit role' })
</script>

<template>
  <div>
    <AdminPageHeader title="Edit role" subtitle="Update details and permissions" back-to="/admin/roles" />
    <AdminRoleForm :id="id" :key="id" />
  </div>
</template>
