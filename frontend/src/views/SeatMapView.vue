<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'

import { useCreateBookingsMutation, useShowtimeBookingsQuery } from '@/api/bookings'
import { useShowtimeQuery } from '@/api/showtimes'

const props = defineProps<{ movieId: string; showtimeId: string }>()
const movieId = computed(() => props.movieId)
const showtimeId = computed(() => props.showtimeId)
const showtimeQuery = useShowtimeQuery(movieId, showtimeId)
const bookingsQuery = useShowtimeBookingsQuery(showtimeId)
const createBookingsMutation = useCreateBookingsMutation()

const rows = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H']
const seatsPerRow = 10
const selectedSeats = ref(new Set<string>())
const bookingComplete = ref(false)
const occupiedSeats = computed(
  () => new Set((bookingsQuery.data.value ?? []).map((booking) => booking.seat_number)),
)

const dateTimeFormatter = new Intl.DateTimeFormat(undefined, {
  weekday: 'short',
  month: 'short',
  day: 'numeric',
  hour: 'numeric',
  minute: '2-digit',
})

const sortedSelectedSeats = computed(() => [...selectedSeats.value].sort())

function seatNumber(row: string, number: number) {
  return `${row}${number}`
}

function toggleSeat(seat: string) {
  if (occupiedSeats.value.has(seat)) return

  bookingComplete.value = false
  createBookingsMutation.reset()

  const nextSelection = new Set(selectedSeats.value)
  if (nextSelection.has(seat)) {
    nextSelection.delete(seat)
  } else {
    nextSelection.add(seat)
  }
  selectedSeats.value = nextSelection
}

function isSelected(seat: string) {
  return selectedSeats.value.has(seat)
}

async function createBooking() {
  if (selectedSeats.value.size === 0) return

  try {
    await createBookingsMutation.mutateAsync({
      show_time_id: props.showtimeId,
      seat_numbers: sortedSelectedSeats.value,
    })
    selectedSeats.value = new Set()
    bookingComplete.value = true
  } catch {
    await bookingsQuery.refetch()
  }
}

function retrySeatMap() {
  void showtimeQuery.refetch()
  void bookingsQuery.refetch()
}
</script>

<template>
  <section class="mx-auto w-full max-w-5xl" aria-labelledby="seat-map-heading">
    <RouterLink
      :to="{ name: 'movie-showtimes', params: { movieId } }"
      class="mb-8 inline-flex items-center gap-2 text-sm font-medium text-slate-400 transition-colors hover:text-white"
    >
      <span aria-hidden="true">←</span> Back to showtimes
    </RouterLink>

    <div
      v-if="showtimeQuery.isPending.value || bookingsQuery.isPending.value"
      class="py-16 text-center text-slate-400"
    >
      Loading seat map…
    </div>

    <div
      v-else-if="showtimeQuery.isError.value || bookingsQuery.isError.value"
      class="rounded-2xl border border-slate-800 bg-slate-900/80 px-6 py-14 text-center text-slate-400"
      role="alert"
    >
      <strong class="mb-1.5 block text-slate-100">Couldn’t load this showtime</strong>
      <button
        class="mt-4 cursor-pointer rounded-lg bg-red-600 px-4 py-2 font-medium text-white hover:bg-red-500"
        type="button"
        @click="retrySeatMap"
      >
        Try again
      </button>
    </div>

    <template v-else-if="showtimeQuery.data.value">
      <header class="mb-8">
        <p class="mb-1.5 text-xs font-bold tracking-[0.16em] text-red-400 uppercase">
          Select your seats
        </p>
        <h1 id="seat-map-heading" class="text-3xl font-bold tracking-tight sm:text-5xl">
          {{ showtimeQuery.data.value.movie.title }}
        </h1>
        <p class="mt-2.5 text-slate-400">
          {{ dateTimeFormatter.format(new Date(showtimeQuery.data.value.start_time * 1000)) }}
        </p>
      </header>

      <div class="rounded-2xl border border-slate-800 bg-slate-900/80 p-4 shadow-2xl sm:p-8">
        <div class="mx-auto mb-12 max-w-2xl">
          <div class="h-2 rounded-[50%] bg-slate-200 shadow-[0_6px_22px_rgba(226,232,240,0.35)]" />
          <p
            class="mt-3 text-center text-xs font-semibold tracking-[0.24em] text-slate-500 uppercase"
          >
            Screen
          </p>
        </div>

        <div class="overflow-x-auto pb-3">
          <div class="mx-auto w-max space-y-2.5" aria-label="Cinema seat map">
            <div v-for="row in rows" :key="row" class="flex items-center gap-2">
              <span class="w-5 text-center text-xs font-semibold text-slate-500">{{ row }}</span>
              <button
                v-for="number in seatsPerRow"
                :key="seatNumber(row, number)"
                class="h-8 w-8 cursor-pointer rounded-t-lg rounded-b-sm border text-[10px] font-bold transition-colors focus-visible:ring-2 focus-visible:ring-red-400 focus-visible:outline-none sm:h-10 sm:w-10 sm:text-xs"
                :class="[
                  number === 6 ? 'ml-4 sm:ml-7' : '',
                  occupiedSeats.has(seatNumber(row, number))
                    ? 'cursor-not-allowed border-slate-800 bg-slate-800 text-slate-600'
                    : isSelected(seatNumber(row, number))
                      ? 'border-red-500 bg-red-600 text-white'
                      : 'border-slate-600 bg-slate-700 text-slate-200 hover:border-slate-400 hover:bg-slate-600',
                ]"
                type="button"
                :disabled="occupiedSeats.has(seatNumber(row, number))"
                :aria-label="`Seat ${seatNumber(row, number)}${occupiedSeats.has(seatNumber(row, number)) ? ', occupied' : ''}`"
                :aria-pressed="isSelected(seatNumber(row, number))"
                @click="toggleSeat(seatNumber(row, number))"
              >
                {{ number }}
              </button>
              <span class="w-5 text-center text-xs font-semibold text-slate-500">{{ row }}</span>
            </div>
          </div>
        </div>

        <div class="mt-8 flex flex-wrap justify-center gap-x-6 gap-y-3 text-sm text-slate-400">
          <div class="flex items-center gap-2">
            <span class="h-4 w-4 rounded-t-md border border-slate-600 bg-slate-700" /> Available
          </div>
          <div class="flex items-center gap-2">
            <span class="h-4 w-4 rounded-t-md border border-red-500 bg-red-600" /> Selected
          </div>
          <div class="flex items-center gap-2">
            <span class="h-4 w-4 rounded-t-md border border-slate-800 bg-slate-800" /> Occupied
          </div>
        </div>
      </div>

      <footer
        class="mt-5 flex flex-col gap-4 rounded-2xl border border-slate-800 bg-slate-900 p-5 sm:flex-row sm:items-center sm:justify-between"
      >
        <div>
          <p class="text-sm text-slate-400">
            {{ selectedSeats.size }} {{ selectedSeats.size === 1 ? 'seat' : 'seats' }} selected
          </p>
          <p class="mt-1 min-h-6 font-semibold text-slate-100">
            {{
              sortedSelectedSeats.length ? sortedSelectedSeats.join(', ') : 'Choose a seat above'
            }}
          </p>
          <p v-if="bookingComplete" class="mt-2 text-sm font-medium text-emerald-400" role="status">
            Your seats have been booked successfully.
          </p>
          <p
            v-else-if="createBookingsMutation.isError.value"
            class="mt-2 text-sm font-medium text-red-400"
            role="alert"
          >
            {{ createBookingsMutation.error.value?.message }} Please select available seats and try
            again.
          </p>
        </div>
        <button
          class="rounded-lg bg-red-600 px-6 py-3 font-semibold text-white transition-colors hover:bg-red-500 disabled:cursor-not-allowed disabled:bg-slate-800 disabled:text-slate-500"
          type="button"
          :disabled="selectedSeats.size === 0 || createBookingsMutation.isPending.value"
          @click="createBooking"
        >
          {{ createBookingsMutation.isPending.value ? 'Booking…' : 'Book seats' }}
        </button>
      </footer>
    </template>
  </section>
</template>
