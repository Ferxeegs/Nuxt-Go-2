import { useDisplay } from 'vuetify'

const STORAGE_KEY = 'sidebar-rail'

function readCollapsed(): boolean {
  try {
    return localStorage.getItem(STORAGE_KEY) === '1'
  } catch {
    return false
  }
}

/**
 * Single source of truth for the app sidebar.
 *  - desktop (>= md): permanent drawer, can be collapsed to an icon rail that
 *    expands as an overlay while hovered (state persisted per browser)
 *  - mobile/tablet: temporary drawer toggled by the header button
 */
export function useSidebar() {
  const { mdAndUp } = useDisplay()

  const collapsed = useState<boolean>('sidebar-collapsed', () => readCollapsed())
  const mobileOpen = useState<boolean>('sidebar-mobile-open', () => false)
  const hovering = useState<boolean>('sidebar-hovering', () => false)

  const isDesktop = computed(() => mdAndUp.value)
  /** Drawer is in rail mode (icons only unless hovered). */
  const railMode = computed(() => isDesktop.value && collapsed.value)
  /** Content should render in its compact form right now. */
  const compact = computed(() => railMode.value && !hovering.value)

  // Drawer v-model: always shown on desktop, toggled on mobile.
  const drawerModel = computed({
    get: () => isDesktop.value || mobileOpen.value,
    set: (v: boolean) => {
      if (!isDesktop.value) mobileOpen.value = v
    },
  })

  function toggle() {
    if (isDesktop.value) {
      collapsed.value = !collapsed.value
      hovering.value = false
      try {
        localStorage.setItem(STORAGE_KEY, collapsed.value ? '1' : '0')
      } catch {
        /* storage unavailable: state just won't persist */
      }
    } else {
      mobileOpen.value = !mobileOpen.value
    }
  }

  // Must mirror Vuetify's own expand-on-hover timing (instant) so the drawer
  // width and its content never get out of sync.
  function onEnter() {
    if (railMode.value) hovering.value = true
  }
  function onLeave() {
    hovering.value = false
  }

  watch(isDesktop, (desktop) => {
    hovering.value = false
    if (desktop) mobileOpen.value = false
  })

  return { collapsed, mobileOpen, hovering, isDesktop, railMode, compact, drawerModel, toggle, onEnter, onLeave }
}
