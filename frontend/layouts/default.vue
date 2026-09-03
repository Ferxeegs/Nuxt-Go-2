<script setup lang="ts">
const title = ref(
  "Company Profile - Admin Dashboard"
);
useHead({
  meta: [{ content: title }],
  titleTemplate: (titleChunk) => {
    return titleChunk
      ? `${titleChunk} - Company Profile`
      : "Company Profile - Admin Dashboard";
  },
});
</script>

<template>
  <v-locale-provider>
    <v-app>
      <v-main>
        <LayoutMain />
        <v-container
          fluid
          class="page-wrapper bg-background px-sm-5 px-4 pt-12 rounded-xl"
        >
          <div class="maxWidth">
            <AuthImpersonationBanner />
            <slot />
          </div>
        </v-container>
      </v-main>
    </v-app>
  </v-locale-provider>
</template>
