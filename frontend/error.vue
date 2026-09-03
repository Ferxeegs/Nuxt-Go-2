<template>
  <div class="d-flex justify-center align-center text-center h-100">
    <div>
      <img :src="errorImg" width="500" :alt="String(statusCode)" />
      <h1 class="text-h1 pt-3">{{ statusCode }}</h1>
      <h4 class="text-h4 my-8">
        {{ message }}
      </h4>
      <v-btn rounded="pill" color="primary" class="mb-4" to="/">
        Go Back to Home
      </v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ error: { statusCode?: number; message?: string; statusMessage?: string } }>()

// Keep as a runtime string so Vite/Nuxt do not rewrite this to virtual:public.
const errorImg = "/images/backgrounds/errorimg.svg"

const statusCode = computed(() => props.error?.statusCode || 500)
const message = computed(
  () =>
    props.error?.statusMessage ||
    props.error?.message ||
    (statusCode.value === 404
      ? 'This page you are looking for could not be found.'
      : 'Something went wrong.'),
)
</script>
