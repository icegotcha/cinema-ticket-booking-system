<script setup lang="ts">
import MovieTable from '@/components/MovieTable.vue'
</script>

<template>
  <section class="mx-auto w-full max-w-5xl" aria-labelledby="movies-heading">
    <header class="mb-7">
      <p class="mb-1.5 text-xs font-bold tracking-[0.16em] text-red-400 uppercase">Now showing</p>
      <h1 id="movies-heading" class="text-4xl leading-tight font-bold tracking-tight sm:text-5xl">
        Movies
      </h1>
      <p class="mt-2.5 text-slate-400">Browse the movies currently available for booking.</p>
    </header>

    <MovieTable />
  </section>
</template>
