<script setup lang="ts">
useHead({ title: 'New user' })
</script>

<template>
  <div>
    <AdminPageHeader title="New user" subtitle="Create an account and assign roles" back-to="/admin/users" />
    <AdminUserForm />
  </div>
</template>
