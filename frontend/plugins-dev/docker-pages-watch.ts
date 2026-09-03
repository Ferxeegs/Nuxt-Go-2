import { resolve } from 'node:path'
import type { Plugin, ViteDevServer } from 'vite'

/**
 * Docker bind mounts (esp. Windows/macOS) often fail to propagate "add file"
 * events into Vite's module graph. Nuxt may regenerate `.nuxt/routes.mjs` on
 * disk while Vite keeps serving a stale transform → client 404 for new pages.
 *
 * Restart the Vite dev server when pages are added/removed so routes reload.
 */
export function dockerPagesWatchPlugin(pagesDir: string): Plugin {
  const normalizedPages = resolve(pagesDir).replace(/\\/g, '/')

  return {
    name: 'docker-pages-watch',
    apply: 'serve',
    configureServer(server: ViteDevServer) {
      const restart = debounce(() => {
        // eslint-disable-next-line no-console
        console.info('[docker-pages-watch] pages changed — restarting Vite to refresh routes')
        server.restart()
      }, 400)

      const onFsEvent = (file: string) => {
        const normalized = file.replace(/\\/g, '/')
        if (!normalized.startsWith(normalizedPages)) return
        if (!normalized.endsWith('.vue')) return
        restart()
      }

      server.watcher.add(normalizedPages)
      server.watcher.on('add', onFsEvent)
      server.watcher.on('unlink', onFsEvent)
    },
  }
}

function debounce(fn: () => void, wait: number) {
  let timer: ReturnType<typeof setTimeout> | null = null
  return () => {
    if (timer) clearTimeout(timer)
    timer = setTimeout(fn, wait)
  }
}
