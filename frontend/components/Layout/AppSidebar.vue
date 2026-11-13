<script setup lang="ts">
import sidebarItems, {
  filterSidebarByPermission,
} from '@/components/Layout/Full/vertical-sidebar/sidebarItem'

const { hasPermission } = useAuth()
const { isDesktop, railMode, compact, drawerModel, mobileOpen, onEnter, onLeave } = useSidebar()

const menu = computed(() => filterSidebarByPermission(sidebarItems, hasPermission))

// Close the mobile drawer after navigating.
const route = useRoute()
watch(() => route.path, () => {
  mobileOpen.value = false
})
</script>

<template>
  <v-navigation-drawer
    v-model="drawerModel"
    :permanent="isDesktop"
    :temporary="!isDesktop"
    :rail="railMode"
    :expand-on-hover="railMode"
    :rail-width="76"
    :width="268"
    :border="false"
    class="app-sidebar"
    @mouseenter="onEnter"
    @mouseleave="onLeave"
  >
    <div class="app-sidebar__brand">
      <LayoutAppBrand :collapsed="compact" />
    </div>

    <perfect-scrollbar class="app-sidebar__scroll" :options="{ suppressScrollX: true }">
      <v-list nav density="comfortable" class="app-sidebar__list" :class="{ 'is-rail': compact }">
        <LayoutSidebarNavItem
          v-for="(item, i) in menu"
          :key="(item.title || item.header || '') + i"
          :item="item"
          :rail="compact"
        />
      </v-list>
    </perfect-scrollbar>
  </v-navigation-drawer>
</template>
