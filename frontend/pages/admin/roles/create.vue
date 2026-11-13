<script setup lang="ts">
useHead({ title: 'New role' })
</script>

<template>
  <div>
    <AdminPageHeader title="New role" subtitle="Define a named set of permissions" back-to="/admin/roles" />
    <AdminRoleForm />
  </div>
</template>
