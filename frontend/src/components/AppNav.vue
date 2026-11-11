<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { useAuth } from '@/composables/useAuth'

const auth = useAuth()
</script>

<template>
  <nav class="border-b border-slate-800 bg-slate-950/90 px-4 backdrop-blur sm:px-6">
    <div class="mx-auto flex h-16 max-w-5xl items-center justify-between">
      <RouterLink
        to="/"
        class="flex items-center gap-2 text-lg font-black tracking-tight text-white"
      >
        <span class="rounded bg-red-600 px-2 py-1 text-xs tracking-widest">CINEMA</span>
        <span class="hidden sm:inline">Tickets</span>
      </RouterLink>
      <div class="flex items-center gap-3">
        <template v-if="auth.user.value">
          <span class="hidden max-w-48 truncate text-sm text-slate-400 sm:block">{{
            auth.user.value.email
          }}</span>
          <button
            class="cursor-pointer rounded-lg border border-slate-700 px-4 py-2 text-sm font-semibold text-slate-200 hover:border-slate-500 hover:bg-slate-800"
            type="button"
            @click="auth.signOut"
          >
            Sign out
          </button>
        </template>
        <button
          v-else
          class="cursor-pointer rounded-lg bg-red-600 px-4 py-2 text-sm font-semibold text-white hover:bg-red-500"
          type="button"
          @click="auth.openAuth()"
        >
          Sign in
        </button>
      </div>
    </div>
  </nav>
</template>
