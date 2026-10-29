<script setup lang="ts">
import { RouterView } from 'vue-router'
import AppNav from '@/components/AppNav.vue'
import AuthModal from '@/components/AuthModal.vue'
</script>

<template>
  <div class="min-h-screen bg-slate-950 text-slate-100">
    <AppNav />
    <main class="px-4 py-10 sm:px-6 sm:py-16"><RouterView /></main>
    <AuthModal />
  </div>
</template>
