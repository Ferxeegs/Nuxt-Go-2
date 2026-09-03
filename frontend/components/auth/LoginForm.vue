<script setup lang="ts">
const config = useRuntimeConfig()
const isDev = import.meta.dev

const email = ref(
  isDev ? String(config.public.devLoginEmail || '') : '',
)
const password = ref(
  isDev ? String(config.public.devLoginPassword || '') : '',
)
const remember = ref(true)
const errorMessage = ref('')
const submitting = ref(false)

const turnstileSiteKey = computed(() => (config.public.turnstileSiteKey as string) || '')
const turnstileToken = ref('')
const turnstileHost = ref<HTMLElement | null>(null)
let turnstileWidgetId: string | null = null
let scriptLoadPromise: Promise<void> | null = null

const { login } = useAuth()

declare global {
  interface Window {
    turnstile?: {
      render: (
        el: HTMLElement,
        options: {
          sitekey: string
          callback: (token: string) => void
          'expired-callback'?: () => void
          'error-callback'?: () => void
          theme?: 'light' | 'dark' | 'auto'
        },
      ) => string
      reset: (widgetId?: string) => void
      remove: (widgetId?: string) => void
    }
  }
}

function loadTurnstileScript(): Promise<void> {
  if (window.turnstile) {
    return Promise.resolve()
  }
  if (scriptLoadPromise) {
    return scriptLoadPromise
  }
  scriptLoadPromise = new Promise((resolve, reject) => {
    const existing = document.querySelector<HTMLScriptElement>('script[data-cf-turnstile]')
    if (existing) {
      existing.addEventListener('load', () => resolve())
      existing.addEventListener('error', () => reject(new Error('Failed to load captcha')))
      return
    }
    const script = document.createElement('script')
    script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit'
    script.async = true
    script.defer = true
    script.dataset.cfTurnstile = 'true'
    script.onload = () => resolve()
    script.onerror = () => reject(new Error('Failed to load captcha'))
    document.head.appendChild(script)
  })
  return scriptLoadPromise
}

async function mountTurnstile() {
  if (!turnstileSiteKey.value || !turnstileHost.value) {
    return
  }
  try {
    await loadTurnstileScript()
    if (!window.turnstile || !turnstileHost.value) {
      return
    }
    if (turnstileWidgetId) {
      window.turnstile.remove(turnstileWidgetId)
      turnstileWidgetId = null
    }
    turnstileToken.value = ''
    turnstileWidgetId = window.turnstile.render(turnstileHost.value, {
      sitekey: turnstileSiteKey.value,
      callback: (token: string) => {
        turnstileToken.value = token
      },
      'expired-callback': () => {
        turnstileToken.value = ''
      },
      'error-callback': () => {
        turnstileToken.value = ''
      },
      theme: 'light',
    })
  } catch {
    errorMessage.value = 'Failed to load security check. Please refresh the page.'
  }
}

function resetTurnstile() {
  turnstileToken.value = ''
  if (window.turnstile && turnstileWidgetId) {
    window.turnstile.reset(turnstileWidgetId)
  }
}

onMounted(() => {
  if (turnstileSiteKey.value) {
    nextTick(() => {
      void mountTurnstile()
    })
  }
})

onBeforeUnmount(() => {
  if (window.turnstile && turnstileWidgetId) {
    window.turnstile.remove(turnstileWidgetId)
    turnstileWidgetId = null
  }
})

async function onSubmit() {
  errorMessage.value = ''
  if (turnstileSiteKey.value && !turnstileToken.value) {
    errorMessage.value = 'Security check required. Please complete the captcha and try again.'
    return
  }
  submitting.value = true
  try {
    const result = await login(email.value.trim(), password.value, turnstileToken.value)
    if (result.requiresRoleSelection) {
      await navigateTo('/auth/choose-role')
    } else {
      await navigateTo('/')
    }
  } catch (err: unknown) {
    const data = err as { data?: { message?: string }; message?: string }
    errorMessage.value = data?.data?.message || data?.message || 'Login failed'
    resetTurnstile()
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <form @submit.prevent="onSubmit">
    <v-row class="mb-3">
      <v-col cols="12">
        <v-label class="font-weight-medium mb-1">Email</v-label>
        <v-text-field
          v-model="email"
          type="email"
          variant="outlined"
          hide-details="auto"
          color="primary"
          autocomplete="username"
          required
        />
      </v-col>
      <v-col cols="12">
        <v-label class="font-weight-medium mb-1">Password</v-label>
        <v-text-field
          v-model="password"
          variant="outlined"
          type="password"
          hide-details="auto"
          color="primary"
          autocomplete="current-password"
          required
        />
      </v-col>
      <v-col cols="12" class="py-0">
        <div class="d-flex flex-wrap align-center w-100">
          <v-checkbox v-model="remember" hide-details color="primary" label="Remember this device" />
        </div>
      </v-col>
      <v-col v-if="turnstileSiteKey" cols="12">
        <div ref="turnstileHost" class="d-flex justify-center" />
      </v-col>
      <v-col v-if="errorMessage" cols="12">
        <v-alert type="error" variant="tonal" density="compact">{{ errorMessage }}</v-alert>
      </v-col>
      <v-col cols="12">
        <v-btn
          size="large"
          rounded="pill"
          color="primary"
          class="rounded-pill"
          block
          type="submit"
          flat
          :loading="submitting"
        >
          Sign In
        </v-btn>
      </v-col>
    </v-row>
  </form>
</template>
