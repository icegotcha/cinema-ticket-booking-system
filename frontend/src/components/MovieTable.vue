<script setup lang="ts">
import { computed, h } from 'vue'
import { FlexRender, createColumnHelper, getCoreRowModel, useVueTable } from '@tanstack/vue-table'
import { useRouter } from 'vue-router'

import { type Movie, useMoviesQuery } from '@/api/movies'

const { data, error, isPending, isError, refetch } = useMoviesQuery()
const movies = computed(() => data.value ?? [])
const router = useRouter()

function openShowtimes(movie: Movie) {
  void router.push({ name: 'movie-showtimes', params: { movieId: movie.id } })
}

const posterFallback =
  'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="72" height="104" viewBox="0 0 72 104"%3E%3Crect width="72" height="104" rx="6" fill="%23202938"/%3E%3Cpath d="M25 45h22v17H25zM29 41l4 4m6-4 4 4" fill="none" stroke="%2364748b" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/%3E%3C/svg%3E'

const columnHelper = createColumnHelper<Movie>()
const columns = [
  columnHelper.display({
    id: 'number',
    header: '#',
    cell: ({ row }) => h('span', { class: 'text-slate-500 tabular-nums' }, row.index + 1),
  }),
  columnHelper.accessor('poster_link', {
    header: 'Poster',
    cell: ({ getValue, row }) =>
      h('img', {
        class: 'block h-20 w-14 rounded-md bg-slate-800 object-cover shadow-lg',
        src: getValue() || posterFallback,
        alt: `${row.original.title} poster`,
        loading: 'lazy',
        width: 56,
        height: 80,
        onError: (event: Event) => {
          const image = event.currentTarget as HTMLImageElement
          image.onerror = null
          image.src = posterFallback
        },
      }),
  }),
  columnHelper.accessor('title', {
    header: 'Title',
    cell: ({ getValue }) => h('span', { class: 'font-semibold' }, getValue()),
  }),
]

const table = useVueTable({
  get data() {
    return movies.value
  },
  columns,
  getCoreRowModel: getCoreRowModel(),
})
</script>

<template>
  <div class="overflow-hidden rounded-2xl border border-slate-800 bg-slate-900/80 shadow-2xl">
    <div v-if="isPending" class="px-6 py-14 text-center text-slate-400" role="status">
      Loading movies…
    </div>

    <div v-else-if="isError" class="px-6 py-14 text-center text-slate-400" role="alert">
      <strong class="mb-1.5 block text-slate-100">Couldn’t load movies</strong>
      <span>{{ error?.message }}</span>
      <div>
        <button
          class="mt-4 cursor-pointer rounded-lg bg-red-600 px-4 py-2 font-medium text-white transition-colors hover:bg-red-500 focus-visible:ring-2 focus-visible:ring-red-400 focus-visible:outline-none"
          type="button"
          @click="refetch()"
        >
          Try again
        </button>
      </div>
    </div>

    <div v-else-if="movies.length === 0" class="px-6 py-14 text-center text-slate-400">
      <strong class="mb-1.5 block text-slate-100">No movies found</strong>
      <span>Check back again soon.</span>
    </div>

    <div v-else class="overflow-x-auto">
      <table class="w-full border-collapse">
        <thead class="bg-slate-900 text-xs tracking-wider text-slate-400 uppercase">
          <tr v-for="headerGroup in table.getHeaderGroups()" :key="headerGroup.id">
            <th
              v-for="header in headerGroup.headers"
              :key="header.id"
              class="px-4 py-4 text-left font-semibold sm:px-5"
              :class="{
                'w-20': header.column.id === 'number',
                'w-24': header.column.id === 'poster_link',
              }"
              :colspan="header.colSpan"
              scope="col"
            >
              <FlexRender
                v-if="!header.isPlaceholder"
                :render="header.column.columnDef.header"
                :props="header.getContext()"
              />
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in table.getRowModel().rows"
            :key="row.id"
            class="cursor-pointer border-t border-slate-800 transition-colors hover:bg-slate-800/70 focus-visible:bg-slate-800/70 focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-red-400 focus-visible:outline-none"
            tabindex="0"
            :aria-label="`View showtimes for ${row.original.title}`"
            @click="openShowtimes(row.original)"
            @keydown.enter="openShowtimes(row.original)"
            @keydown.space.prevent="openShowtimes(row.original)"
          >
            <td v-for="cell in row.getVisibleCells()" :key="cell.id" class="px-4 py-4 sm:px-5">
              <FlexRender :render="cell.column.columnDef.cell" :props="cell.getContext()" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
