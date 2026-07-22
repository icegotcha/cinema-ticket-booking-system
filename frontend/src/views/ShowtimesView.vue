<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

import { useMovieQuery } from '@/api/movies'
import { type Showtime, useMovieShowtimesQuery } from '@/api/showtimes'

const props = defineProps<{ movieId: string }>()
const movieId = computed(() => props.movieId)

const movieQuery = useMovieQuery(movieId)
const showtimesQuery = useMovieShowtimesQuery(movieId)

const posterFallback =
  'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="200" height="300" viewBox="0 0 200 300"%3E%3Crect width="200" height="300" rx="12" fill="%23202938"/%3E%3Cpath d="M70 128h60v46H70zM80 116l12 12m20-12 12 12" fill="none" stroke="%2364748b" stroke-width="6" stroke-linecap="round" stroke-linejoin="round"/%3E%3C/svg%3E'

const dateFormatter = new Intl.DateTimeFormat(undefined, {
  weekday: 'long',
  month: 'long',
  day: 'numeric',
})
const timeFormatter = new Intl.DateTimeFormat(undefined, {
  hour: 'numeric',
  minute: '2-digit',
})

function showtimeDate(showtime: Showtime) {
  return new Date(showtime.start_time * 1000)
}

const groupedShowtimes = computed(() => {
  const groups = new Map<string, { date: Date; showtimes: Showtime[] }>()

  for (const showtime of showtimesQuery.data.value ?? []) {
    const date = showtimeDate(showtime)
    const key = `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`
    const group = groups.get(key)

    if (group) {
      group.showtimes.push(showtime)
    } else {
      groups.set(key, { date, showtimes: [showtime] })
    }
  }

  return [...groups.values()].sort((a, b) => a.date.getTime() - b.date.getTime())
})

function useFallbackPoster(event: Event) {
  const image = event.currentTarget as HTMLImageElement
  image.onerror = null
  image.src = posterFallback
}
</script>

<template>
  <section class="mx-auto w-full max-w-5xl" aria-labelledby="showtimes-heading">
    <RouterLink
      to="/"
      class="mb-8 inline-flex items-center gap-2 text-sm font-medium text-slate-400 transition-colors hover:text-white"
    >
      <span aria-hidden="true">←</span> All movies
    </RouterLink>

    <div v-if="movieQuery.isPending.value" class="py-16 text-center text-slate-400">
      Loading movie…
    </div>

    <div v-else-if="movieQuery.isError.value" class="py-16 text-center text-slate-400" role="alert">
      <strong class="mb-1.5 block text-slate-100">Couldn’t load this movie</strong>
      <button
        class="mt-4 cursor-pointer rounded-lg bg-red-600 px-4 py-2 font-medium text-white hover:bg-red-500"
        type="button"
        @click="movieQuery.refetch()"
      >
        Try again
      </button>
    </div>

    <template v-else-if="movieQuery.data.value">
      <header class="mb-10 flex items-end gap-5 sm:gap-7">
        <img
          class="h-40 w-28 shrink-0 rounded-xl bg-slate-800 object-cover shadow-2xl sm:h-52 sm:w-36"
          :src="movieQuery.data.value.poster_link || posterFallback"
          :alt="`${movieQuery.data.value.title} poster`"
          width="144"
          height="208"
          @error="useFallbackPoster"
        />
        <div class="pb-1">
          <p class="mb-1.5 text-xs font-bold tracking-[0.16em] text-red-400 uppercase">
            Select a showtime
          </p>
          <h1
            id="showtimes-heading"
            class="text-3xl leading-tight font-bold tracking-tight sm:text-5xl"
          >
            {{ movieQuery.data.value.title }}
          </h1>
        </div>
      </header>

      <div v-if="showtimesQuery.isPending.value" class="py-14 text-center text-slate-400">
        Loading showtimes…
      </div>

      <div
        v-else-if="showtimesQuery.isError.value"
        class="rounded-2xl border border-slate-800 bg-slate-900/80 px-6 py-14 text-center text-slate-400"
        role="alert"
      >
        <strong class="mb-1.5 block text-slate-100">Couldn’t load showtimes</strong>
        <button
          class="mt-4 cursor-pointer rounded-lg bg-red-600 px-4 py-2 font-medium text-white hover:bg-red-500"
          type="button"
          @click="showtimesQuery.refetch()"
        >
          Try again
        </button>
      </div>

      <div
        v-else-if="groupedShowtimes.length === 0"
        class="rounded-2xl border border-slate-800 bg-slate-900/80 px-6 py-14 text-center text-slate-400"
      >
        <strong class="mb-1.5 block text-slate-100">No showtimes available</strong>
        Check back again soon.
      </div>

      <div v-else class="space-y-4">
        <article
          v-for="group in groupedShowtimes"
          :key="group.date.toISOString()"
          class="rounded-2xl border border-slate-800 bg-slate-900/80 p-5 sm:p-6"
        >
          <h2 class="mb-4 font-semibold text-slate-200">{{ dateFormatter.format(group.date) }}</h2>
          <div class="flex flex-wrap gap-3">
            <button
              v-for="showtime in group.showtimes"
              :key="showtime.id"
              class="cursor-pointer rounded-lg border border-slate-700 bg-slate-800 px-4 py-2.5 font-semibold text-slate-100 transition-colors hover:border-red-500 hover:bg-red-600 focus-visible:ring-2 focus-visible:ring-red-400 focus-visible:outline-none"
              type="button"
            >
              {{ timeFormatter.format(showtimeDate(showtime)) }}
            </button>
          </div>
        </article>
      </div>
    </template>
  </section>
</template>
