<script setup lang="ts">
import { computed, h, onMounted, ref, watch } from 'vue'
import {
  FlexRender,
  createColumnHelper,
  getCoreRowModel,
  useVueTable,
  type PaginationState,
} from '@tanstack/vue-table'
import { useRouter } from 'vue-router'
import {
  AdminSessionExpiredError,
  getAdminBookings,
  logoutAdmin,
  type AdminBookings,
} from '@/api/admin'
import '@/styles/admin.css'

const router = useRouter()
const snapshot = ref<AdminBookings | null>(null)
const busy = ref(false)
const error = ref('')
const movie = ref('')
const date = ref('')
const user = ref('')
const pagination = ref<PaginationState>({ pageIndex: 0, pageSize: 25 })
const rows = computed(() =>
  (snapshot.value?.bookings ?? []).map((booking) => {
    const showtime = snapshot.value?.showtimes?.find((s) => s.id === booking.show_time_id)
    const film = snapshot.value?.movies?.find((m) => m.id === showtime?.movie_id)
    return {
      ...booking,
      title: film?.series_title ?? 'Unavailable movie',
      starts: showtime ? new Date(showtime.start_time.T * 1000) : null,
    }
  }),
)
const column = createColumnHelper<(typeof rows.value)[number]>()
const columns = [
  column.accessor('id', { header: 'Booking ID' }),
  column.accessor('title', { header: 'Movie' }),
  column.accessor('starts', {
    header: 'Show date',
    cell: ({ getValue }) => getValue()?.toLocaleString() ?? 'Unavailable showtime',
  }),
  column.accessor('user_id', { header: 'User ID' }),
  column.accessor('seat_number', { header: 'Seat' }),
  column.accessor('status', {
    header: 'Status',
    cell: ({ getValue }) =>
      h(
        'span',
        {
          class: getValue() === 'BOOKED' ? 'text-emerald-400' : 'text-slate-300',
        },
        getValue(),
      ),
  }),
]
const table = useVueTable({
  get data() {
    return rows.value
  },
  columns,
  getCoreRowModel: getCoreRowModel(),
  getRowId: (row) => row.id,
  manualPagination: true,
  manualFiltering: true,
  manualSorting: true,
  get rowCount() {
    return snapshot.value?.total ?? 0
  },
  state: {
    get pagination() {
      return pagination.value
    },
  },
  onPaginationChange: (updater) => {
    pagination.value = typeof updater === 'function' ? updater(pagination.value) : updater
    void load()
  },
})
let requestId = 0
async function load() {
  const currentRequest = ++requestId
  busy.value = true
  error.value = ''
  const params = new URLSearchParams({
    page: String(pagination.value.pageIndex + 1),
    page_size: String(pagination.value.pageSize),
    movie_id: movie.value,
    user_id: user.value.trim(),
  })
  if (date.value) {
    const start = new Date(`${date.value}T00:00:00`)
    const end = new Date(start)
    end.setDate(end.getDate() + 1)
    params.set('date_from', String(start.getTime() / 1000))
    params.set('date_to', String(end.getTime() / 1000))
  }
  try {
    const result = await getAdminBookings(params)
    if (currentRequest !== requestId) return
    snapshot.value = result
    if (pagination.value.pageIndex > 0 && !result.bookings?.length) {
      table.setPageIndex(Math.max(0, Math.ceil(result.total / pagination.value.pageSize) - 1))
    }
  } catch (cause) {
    if (currentRequest !== requestId) return
    if (cause instanceof AdminSessionExpiredError) {
      snapshot.value = null
      await router.replace({ name: 'admin-login', query: { expired: '1' } })
      return
    }
    error.value = cause instanceof Error ? cause.message : 'Unable to connect to the server.'
  } finally {
    if (currentRequest === requestId) busy.value = false
  }
}

watch([movie, date, user], () => {
  pagination.value = { ...pagination.value, pageIndex: 0 }
  void load()
})
onMounted(load)

function clearFilters() {
  movie.value = ''
  date.value = ''
  user.value = ''
}
async function logout() {
  busy.value = true
  error.value = ''
  try {
    await logoutAdmin()
    snapshot.value = null
    await router.replace({ name: 'admin-login' })
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : 'Unable to connect to the server.'
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <section class="admin mx-auto max-w-5xl">
    <header class="mb-8 flex flex-wrap items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-semibold">Booking dashboard</h1>
        <p class="mt-2 text-sm text-slate-400">Cinema Tickets · Administrator</p>
      </div>
      <div class="flex gap-3">
        <button :disabled="busy" @click="load()">{{ busy ? 'Refreshing…' : 'Refresh' }}</button
        ><button :disabled="busy" @click="logout">Sign out</button>
      </div>
    </header>
    <p v-if="error" role="alert" class="mb-4 text-red-400">{{ error }}</p>
    <p v-if="busy && !snapshot" role="status" class="text-sm text-slate-400">Loading bookings…</p>
    <template v-if="snapshot">
      <div class="mb-6 grid gap-4 sm:grid-cols-3">
        <label class="text-sm"
          >Movie<select v-model="movie" class="mt-2 w-full">
            <option value="">All movies</option>
            <option v-for="film in snapshot.movies ?? []" :key="film.id" :value="film.id">
              {{ film.series_title }}
            </option>
          </select></label
        >
        <label class="text-sm"
          >Show date (local time)<input v-model="date" type="date" class="mt-2 w-full"
        /></label>
        <label class="text-sm"
          >User ID<input
            v-model="user"
            type="search"
            placeholder="Filter by user ID"
            class="mt-2 w-full"
        /></label>
      </div>
      <div class="mb-3 flex items-center justify-between gap-4">
        <p aria-live="polite" class="text-sm text-slate-400">
          {{ rows.length }} of {{ snapshot.total }} bookings
        </p>
        <button v-if="movie || date || user" @click="clearFilters">Clear filters</button>
      </div>
      <div class="overflow-x-auto rounded-lg border border-slate-700" :aria-busy="busy">
        <table class="w-full text-left text-sm">
          <caption class="sr-only">
            Cinema bookings filtered by movie, show date, and user ID
          </caption>
          <thead>
            <tr v-for="group in table.getHeaderGroups()" :key="group.id">
              <th
                v-for="header in group.headers"
                :key="header.id"
                scope="col"
                :aria-sort="header.column.id === 'starts' ? 'descending' : undefined"
              >
                <FlexRender :render="header.column.columnDef.header" :props="header.getContext()" />
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in table.getRowModel().rows" :key="row.id">
              <td
                v-for="cell in row.getVisibleCells()"
                :key="cell.id"
                :class="{
                  'font-mono text-xs': cell.column.id === 'id',
                  'whitespace-nowrap': cell.column.id === 'starts',
                  'break-all': cell.column.id === 'user_id',
                }"
              >
                <FlexRender :render="cell.column.columnDef.cell" :props="cell.getContext()" />
              </td>
            </tr>
            <tr v-if="!rows.length">
              <td colspan="6" class="py-12 text-center text-slate-400">
                {{
                  movie || date || user
                    ? 'No bookings match these filters.'
                    : 'No bookings yet. Booked seats will appear here.'
                }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <nav
        aria-label="Booking pagination"
        class="mt-4 flex flex-wrap items-center justify-between gap-4"
      >
        <label class="text-sm"
          >Rows per page
          <select
            :value="pagination.pageSize"
            :disabled="busy"
            @change="table.setPageSize(Number(($event.target as HTMLSelectElement).value))"
          >
            <option v-for="size in [10, 25, 50, 100]" :key="size" :value="size">{{ size }}</option>
          </select>
        </label>
        <p aria-live="polite" class="text-sm text-slate-400">
          Page {{ pagination.pageIndex + 1 }} of {{ Math.max(1, table.getPageCount()) }}
        </p>
        <div class="flex gap-3">
          <button :disabled="busy || !table.getCanPreviousPage()" @click="table.previousPage()">
            Previous
          </button>
          <button :disabled="busy || !table.getCanNextPage()" @click="table.nextPage()">
            Next
          </button>
        </div>
      </nav>
    </template>
  </section>
</template>
