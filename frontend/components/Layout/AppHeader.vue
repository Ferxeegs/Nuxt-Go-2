<script setup lang="ts">
import sidebarItems from '@/components/Layout/Full/vertical-sidebar/sidebarItem'

const { toggle } = useSidebar()

const route = useRoute()
const titleByPath = computed(() => {
  const map = new Map<string, string>()
  const walk = (items: typeof sidebarItems) => {
    for (const it of items) {
      if (it.to && it.title) map.set(it.to, it.title)
      if (it.children) walk(it.children)
    }
  }
  walk(sidebarItems)
  return map
})

const crumbs = computed(() => {
  const segments = route.path.split('/').filter(Boolean)
  if (!segments.length) return [{ title: 'Dashboard', to: '/', disabled: true }]
  const out: { title: string; to: string; disabled?: boolean }[] = [{ title: 'Home', to: '/' }]
  let acc = ''
  for (const seg of segments) {
    acc += `/${seg}`
    const isId = /^\d+$/.test(seg) || /^[0-9a-f]{8}-[0-9a-f-]{27}$/i.test(seg)
    const title =
      titleByPath.value.get(acc) ||
      (isId ? 'Edit' : '') ||
      seg.replace(/[-_]/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
    out.push({ title, to: acc })
  }
  out[out.length - 1]!.disabled = true
  return out
})
</script>

<template>
  <v-app-bar flat :height="64" class="app-header">
    <div class="app-header__inner">
      <v-btn icon variant="text" size="small" aria-label="Toggle sidebar" @click="toggle">
        <v-icon icon="mdi-menu" size="22" />
      </v-btn>

      <v-breadcrumbs
        :items="crumbs"
        density="compact"
        class="app-header__crumbs d-none d-sm-flex pa-0 ml-2"
      >
        <template #divider><v-icon icon="mdi-chevron-right" size="16" /></template>
      </v-breadcrumbs>

      <v-spacer />

      <v-menu location="bottom end" offset="10">
        <template #activator="{ props }">
          <v-btn
            v-bind="props"
            icon
            variant="text"
            size="small"
            aria-label="Notifications"
            class="mr-2"
          >
            <v-icon icon="mdi-bell-outline" size="22" />
          </v-btn>
        </template>
        <v-card width="300" elevation="8">
          <div class="pa-4 text-subtitle-2 font-weight-bold">Notifications</div>
          <v-divider />
          <div class="pa-6 text-center text-body-2 text-medium-emphasis">
            <v-icon icon="mdi-bell-check-outline" size="32" class="mb-2" /><br />
            You are all caught up
          </div>
        </v-card>
      </v-menu>

      <LayoutUserMenu />
    </div>
  </v-app-bar>
</template>
