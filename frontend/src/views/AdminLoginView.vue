<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { loginAdmin } from '@/api/admin'
import '@/styles/admin.css'

const router = useRouter()
const route = useRoute()
const username = ref('')
const password = ref('')
const busy = ref(false)
const error = ref('')
const sessionExpired = computed(() => route.query.expired === '1')

async function login() {
  busy.value = true
  error.value = ''
  try {
    await loginAdmin(username.value, password.value)
    password.value = ''
    await router.replace({ name: 'admin' })
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : 'Unable to connect to the server.'
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <section class="admin mx-auto max-w-5xl">
    <form
      class="mx-auto max-w-sm space-y-5 rounded-xl border border-slate-700 p-6"
      @submit.prevent="login"
    >
      <div>
        <h1 class="text-2xl font-semibold">Admin sign in</h1>
        <p class="mt-2 text-sm text-slate-400">Manage cinema bookings.</p>
      </div>
      <label class="block text-sm"
        >Username<input v-model="username" required autocomplete="username" class="mt-2 w-full"
      /></label>
      <label class="block text-sm"
        >Password<input
          v-model="password"
          required
          type="password"
          autocomplete="current-password"
          class="mt-2 w-full"
      /></label>
      <p v-if="sessionExpired && !error" role="status" class="text-sm text-slate-400">
        Your session expired. Please sign in again.
      </p>
      <p v-if="error" role="alert" class="text-sm text-red-400">{{ error }}</p>
      <button type="submit" class="primary w-full" :disabled="busy">
        {{ busy ? 'Signing in…' : 'Sign in' }}
      </button>
    </form>
  </section>
</template>
