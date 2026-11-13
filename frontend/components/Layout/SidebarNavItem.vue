<script setup lang="ts">
import type { menu } from '@/components/Layout/Full/vertical-sidebar/sidebarItem'

const props = defineProps<{ item: menu; collapsed?: boolean }>()
const route = useRoute()

const isGroupActive = computed(() =>
  Boolean(props.item.children?.some((c) => c.to && route.path.startsWith(c.to))),
)
</script>

<template>
  <!-- Section label -->
  <div v-if="item.header" class="nav-section">
    <span class="nav-section__text">{{ item.header }}</span>
    <span class="nav-section__line" />
  </div>

  <!-- Group with children -->
  <v-list-group v-else-if="item.children" :value="item.title">
    <template #activator="{ props: activatorProps }">
      <v-list-item
        v-bind="activatorProps"
        class="nav-item"
        :class="{ 'nav-item--parent-active': isGroupActive }"
        :title="item.title"
        :aria-label="collapsed ? item.title : undefined"
        :ripple="false"
      >
        <template #prepend>
          <v-icon :icon="item.icon" size="20" class="nav-item__icon" />
        </template>
      </v-list-item>
    </template>
    <v-list-item
      v-for="child in item.children"
      :key="child.to || child.title"
      :to="child.to"
      class="nav-item nav-item--child"
      :title="child.title"
      :ripple="false"
    >
      <template #prepend><span class="nav-item__dot" /></template>
    </v-list-item>
  </v-list-group>

  <!-- Single link -->
  <v-list-item
    v-else
    class="nav-item"
    :to="!item.external ? item.to : undefined"
    :href="item.external ? item.to : undefined"
    :target="item.external ? '_blank' : undefined"
    :disabled="item.disabled"
    :title="item.title"
    :aria-label="collapsed ? item.title : undefined"
    :ripple="false"
  >
    <template #prepend>
      <v-icon :icon="item.icon" size="20" class="nav-item__icon" />
    </template>
    <template v-if="item.chip" #append>
      <v-chip size="x-small" color="primary" variant="tonal">{{ item.chip }}</v-chip>
    </template>
  </v-list-item>
</template>
