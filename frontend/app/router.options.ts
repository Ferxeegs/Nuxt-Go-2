import type { RouterConfig } from '@nuxt/schema'

// Docker/Linux path matching is case-sensitive; keep routes tolerant for
// legacy links while page files use canonical lowercase paths.
export default <RouterConfig>{
  sensitive: false,
  strict: false,
}
