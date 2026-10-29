<script setup lang="ts">
import { ref, watch } from 'vue'

import { useAuth } from '@/composables/useAuth'

const auth = useAuth()
const errorMessage = ref('')
const submitting = ref(false)

watch(
  () => auth.modalOpen.value,
  (isOpen) => {
    if (!isOpen) return
    errorMessage.value = ''
  },
)

function friendlyError(error: unknown) {
  if (!(error instanceof Error)) return 'Something went wrong. Please try again.'
  const messages: Record<string, string> = {
    'auth/popup-closed-by-user': 'Google sign-in was cancelled.',
  }
  const code = (error as Error & { code?: string }).code
  return (code && messages[code]) || error.message
}

async function submitGoogle() {
  errorMessage.value = ''
  submitting.value = true
  try {
    await auth.loginWithGoogle()
  } catch (error) {
    errorMessage.value = friendlyError(error)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="auth.modalOpen.value"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 p-4 backdrop-blur-sm"
      role="presentation"
      @mousedown.self="auth.closeAuth"
      @keydown.esc="auth.closeAuth"
    >
      <section
        class="relative w-full max-w-md rounded-2xl border border-slate-700 bg-slate-900 p-6 shadow-2xl sm:p-8"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="'auth-modal-title'"
      >
        <button
          class="absolute top-4 right-4 cursor-pointer rounded-lg p-2 text-xl text-slate-400 hover:bg-slate-800 hover:text-white"
          type="button"
          aria-label="Close"
          @click="auth.closeAuth"
        >
          ×
        </button>

        <p class="mb-2 text-xs font-bold tracking-[0.16em] text-red-400 uppercase">Cinema</p>
        <h2 id="auth-modal-title" class="text-3xl font-bold tracking-tight text-white">Welcome</h2>
        <p class="mt-2 text-sm text-slate-400">Sign in with Google to reserve your seats.</p>

        <button
          class="mt-6 flex w-full cursor-pointer items-center justify-center gap-3 rounded-lg border border-slate-600 bg-white px-4 py-3 font-semibold text-slate-900 transition hover:bg-slate-100 disabled:cursor-wait disabled:opacity-60"
          type="button"
          :disabled="submitting"
          @click="submitGoogle"
        >
          <svg class="h-5 w-5" viewBox="0 0 24 24" aria-hidden="true">
            <path
              fill="#4285F4"
              d="M21.6 12.2c0-.7-.1-1.4-.2-2H12v3.9h5.4a4.7 4.7 0 0 1-2 3v2.6h3.3c1.9-1.8 2.9-4.4 2.9-7.5Z"
            />
            <path
              fill="#34A853"
              d="M12 22c2.7 0 5-.9 6.7-2.3l-3.3-2.6c-.9.6-2.1 1-3.4 1a5.9 5.9 0 0 1-5.5-4.1H3.1v2.7A10 10 0 0 0 12 22Z"
            />
            <path fill="#FBBC05" d="M6.5 14a6 6 0 0 1 0-3.9V7.4H3.1a10 10 0 0 0 0 9.3L6.5 14Z" />
            <path
              fill="#EA4335"
              d="M12 5.9c1.5 0 2.8.5 3.9 1.5l2.9-2.9A9.8 9.8 0 0 0 3.1 7.4l3.4 2.7A5.9 5.9 0 0 1 12 5.9Z"
            />
          </svg>
          Continue with Google
        </button>

        <p v-if="errorMessage" class="mt-4 text-sm text-red-400" role="alert">
          {{ errorMessage }}
        </p>
      </section>
    </div>
  </Teleport>
</template>
