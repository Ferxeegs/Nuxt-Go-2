<script setup lang="ts">
const id = String(useRoute().params.id)
useHead({ title: 'Edit user' })
</script>

<template>
  <div>
    <AdminPageHeader title="Edit user" subtitle="Update account and role assignments" back-to="/admin/users" />
    <AdminUserForm :id="id" :key="id" />
  </div>
</template>
