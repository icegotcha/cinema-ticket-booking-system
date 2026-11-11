<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
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
const rows = computed(() =>
  (snapshot.value?.bookings ?? [])
    .map((booking) => {
      const showtime = snapshot.value?.showtimes?.find((s) => s.id === booking.show_time_id)
      const film = snapshot.value?.movies?.find((m) => m.id === showtime?.movie_id)
      const starts = showtime ? new Date(showtime.start_time.T * 1000) : null
      const day = starts
        ? `${starts.getFullYear()}-${String(starts.getMonth() + 1).padStart(2, '0')}-${String(starts.getDate()).padStart(2, '0')}`
        : ''
      return {
        ...booking,
        movieId: film?.id ?? '',
        title: film?.series_title ?? 'Unavailable movie',
        starts,
        day,
      }
    })
    .filter(
      (row) =>
        (!movie.value || row.movieId === movie.value) &&
        (!date.value || row.day === date.value) &&
        row.user_id.toLowerCase().includes(user.value.trim().toLowerCase()),
    ),
)
async function load() {
  busy.value = true
  error.value = ''
  try {
    snapshot.value = await getAdminBookings()
  } catch (cause) {
    if (cause instanceof AdminSessionExpiredError) {
      snapshot.value = null
      await router.replace({ name: 'admin-login', query: { expired: '1' } })
      return
    }
    error.value = cause instanceof Error ? cause.message : 'Unable to connect to the server.'
  } finally {
    busy.value = false
  }
}

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
          {{ rows.length }} of {{ snapshot.bookings?.length ?? 0 }} bookings
        </p>
        <button v-if="movie || date || user" @click="clearFilters">Clear filters</button>
      </div>
      <div class="overflow-x-auto rounded-lg border border-slate-700" :aria-busy="busy">
        <table class="w-full text-left text-sm">
          <caption class="sr-only">
            Cinema bookings filtered by movie, show date, and user ID
          </caption>
          <thead>
            <tr>
              <th scope="col">Booking ID</th>
              <th scope="col">Movie</th>
              <th scope="col">Show date</th>
              <th scope="col">User ID</th>
              <th scope="col">Seat</th>
              <th scope="col">Status</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row.id">
              <td class="font-mono text-xs">{{ row.id }}</td>
              <td>{{ row.title }}</td>
              <td class="whitespace-nowrap">
                {{ row.starts?.toLocaleString() ?? 'Unavailable showtime' }}
              </td>
              <td class="break-all">{{ row.user_id }}</td>
              <td>{{ row.seat_number }}</td>
              <td>
                <span :class="row.status === 'BOOKED' ? 'text-emerald-400' : 'text-slate-300'">{{
                  row.status
                }}</span>
              </td>
            </tr>
            <tr v-if="!rows.length">
              <td colspan="6" class="py-12 text-center text-slate-400">
                {{
                  snapshot.bookings?.length
                    ? 'No bookings match these filters.'
                    : 'No bookings yet. Booked seats will appear here.'
                }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </section>
</template>
