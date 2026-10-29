<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'

import {
  bookingKeys,
  type Booking,
  useConfirmBookingsMutation,
  useHoldBookingsMutation,
  useReleaseBookingsMutation,
  useShowtimeBookingsQuery,
} from '@/api/bookings'
import { useShowtimeQuery } from '@/api/showtimes'
import { useAuth } from '@/composables/useAuth'

import { useQueryClient } from '@tanstack/vue-query'
import { useWebSocket } from '@vueuse/core'

const props = defineProps<{ movieId: string; showtimeId: string }>()
const movieId = computed(() => props.movieId)
const showtimeId = computed(() => props.showtimeId)
const showtimeQuery = useShowtimeQuery(movieId, showtimeId)
const bookingsQuery = useShowtimeBookingsQuery(showtimeId)
const holdBookingsMutation = useHoldBookingsMutation()
const releaseBookingsMutation = useReleaseBookingsMutation()
const confirmBookingsMutation = useConfirmBookingsMutation()
const auth = useAuth()

type SeatEvent = {
  type: 'SEAT_SNAPSHOT' | 'SEAT_UPDATED'
  bookings: Array<{
    id: string
    seat_number: string
    seat_status: string
  }>
}

const rows = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H']
const seatsPerRow = 10
const selectedSeats = ref(new Set<string>())
const pendingSeats = ref(new Set<string>())
const seatActionError = ref('')
const bookingComplete = ref(false)
const occupyingStatuses = new Set(['PENDING', 'LOCKED', 'BOOKED'])
const seatStatuses = computed(() => {
  const statuses = new Map<string, string>()
  for (const booking of bookingsQuery.data.value ?? []) {
    if (occupyingStatuses.has(booking.status)) statuses.set(booking.seat_number, booking.status)
  }
  return statuses
})
const occupiedSeats = computed(
  () => new Set([...seatStatuses.value.keys()].filter((seat) => !selectedSeats.value.has(seat))),
)

const dateTimeFormatter = new Intl.DateTimeFormat(undefined, {
  weekday: 'short',
  month: 'short',
  day: 'numeric',
  hour: 'numeric',
  minute: '2-digit',
})

const queryClient = useQueryClient()

const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
const websocketUrl = computed(
  () => `${protocol}://${location.host}/ws/showtimes/${showtimeId.value}`,
)
const { data } = useWebSocket(websocketUrl, {
  autoReconnect: {
    retries: 3,
    delay: 1000,
  },
})

watch(data, (rawMessage) => {
  if (!rawMessage) return

  let parsed: unknown
  try {
    parsed = JSON.parse(rawMessage) as unknown
  } catch {
    return
  }

  if (
    !parsed ||
    typeof parsed !== 'object' ||
    !('type' in parsed) ||
    (parsed.type !== 'SEAT_SNAPSHOT' && parsed.type !== 'SEAT_UPDATED') ||
    !('bookings' in parsed) ||
    !Array.isArray(parsed.bookings)
  )
    return

  const event = parsed as SeatEvent
  const nextSelection = new Set(selectedSeats.value)
  for (const updated of event.bookings) {
    if (updated.seat_status === 'AVAILABLE') nextSelection.delete(updated.seat_number)
  }
  selectedSeats.value = nextSelection

  queryClient.setQueryData<Booking[]>(bookingKeys.byShowtime(showtimeId.value), (oldBookings) => {
    const updatedBookings = event.bookings.map((updated) => ({
      id: updated.id,
      show_time_id: showtimeId.value,
      user_id: '',
      seat_number: updated.seat_number,
      status: updated.seat_status,
    }))
    if (event.type === 'SEAT_SNAPSHOT') return updatedBookings

    const bookings = [...(oldBookings ?? [])]
    for (const booking of updatedBookings) {
      const index = bookings.findIndex(
        (existing) => existing.id === booking.id || existing.seat_number === booking.seat_number,
      )
      if (index === -1) {
        bookings.push(booking)
      } else {
        bookings[index] = { ...bookings[index], ...booking }
      }
    }
    return bookings
  })
})

const sortedSelectedSeats = computed(() => [...selectedSeats.value].sort())

function seatNumber(row: string, number: number) {
  return `${row}${number}`
}

function withSeat(seats: Set<string>, seat: string, included: boolean) {
  const updated = new Set(seats)
  if (included) updated.add(seat)
  else updated.delete(seat)
  return updated
}

async function holdSeat(seat: string) {
  await holdBookingsMutation.mutateAsync({
    show_time_id: props.showtimeId,
    seat_numbers: [seat],
  })
  selectedSeats.value = withSeat(selectedSeats.value, seat, true)
}

async function releaseSeat(seat: string) {
  await releaseBookingsMutation.mutateAsync({
    show_time_id: props.showtimeId,
    seat_numbers: [seat],
  })
  selectedSeats.value = withSeat(selectedSeats.value, seat, false)
}

async function toggleSeat(seat: string) {
  if (occupiedSeats.value.has(seat)) return
  if (pendingSeats.value.has(seat)) return
  if (!auth.user.value) {
    auth.openAuth()
    return
  }

  bookingComplete.value = false
  seatActionError.value = ''
  const wasSelected = selectedSeats.value.has(seat)
  pendingSeats.value = withSeat(pendingSeats.value, seat, true)

  try {
    if (wasSelected) await releaseSeat(seat)
    else await holdSeat(seat)
  } catch (error) {
    seatActionError.value = error instanceof Error ? error.message : 'Could not update this seat.'
    if (wasSelected) selectedSeats.value = withSeat(selectedSeats.value, seat, false)
    await bookingsQuery.refetch()
  } finally {
    pendingSeats.value = withSeat(pendingSeats.value, seat, false)
  }
}

function isSelected(seat: string) {
  return selectedSeats.value.has(seat)
}

async function confirmBooking() {
  if (selectedSeats.value.size === 0) return
  if (!auth.user.value) {
    auth.openAuth()
    return
  }

  try {
    await confirmBookingsMutation.mutateAsync({
      show_time_id: props.showtimeId,
      seat_numbers: sortedSelectedSeats.value,
    })
    selectedSeats.value = new Set()
    bookingComplete.value = true
  } catch {
    selectedSeats.value = new Set()
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
                    ? seatStatuses.get(seatNumber(row, number)) === 'LOCKED'
                      ? 'cursor-not-allowed border-amber-700 bg-amber-950 text-amber-500'
                      : 'cursor-not-allowed border-slate-800 bg-slate-800 text-slate-600'
                    : isSelected(seatNumber(row, number))
                      ? 'border-red-500 bg-red-600 text-white'
                      : pendingSeats.has(seatNumber(row, number))
                        ? 'cursor-wait border-amber-500 bg-amber-600 text-white'
                        : 'border-slate-600 bg-slate-700 text-slate-200 hover:border-slate-400 hover:bg-slate-600',
                ]"
                type="button"
                :disabled="
                  occupiedSeats.has(seatNumber(row, number)) ||
                  pendingSeats.has(seatNumber(row, number))
                "
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
            <span class="h-4 w-4 rounded-t-md border border-amber-700 bg-amber-950" /> Locked
          </div>
          <div class="flex items-center gap-2">
            <span class="h-4 w-4 rounded-t-md border border-slate-800 bg-slate-800" /> Booked
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
          <p v-if="selectedSeats.size" class="mt-1 text-xs text-amber-400">
            Seats are held for five minutes. Confirm before the hold expires.
          </p>
          <p v-if="bookingComplete" class="mt-2 text-sm font-medium text-emerald-400" role="status">
            Your seats have been booked successfully.
          </p>
          <p
            v-else-if="seatActionError || confirmBookingsMutation.isError.value"
            class="mt-2 text-sm font-medium text-red-400"
            role="alert"
          >
            {{ seatActionError || confirmBookingsMutation.error.value?.message }} Please select
            available seats and try again.
          </p>
        </div>
        <button
          class="rounded-lg bg-red-600 px-6 py-3 font-semibold text-white transition-colors hover:bg-red-500 disabled:cursor-not-allowed disabled:bg-slate-800 disabled:text-slate-500"
          type="button"
          :disabled="selectedSeats.size === 0 || confirmBookingsMutation.isPending.value"
          @click="confirmBooking"
        >
          {{ confirmBookingsMutation.isPending.value ? 'Confirming…' : 'Confirm booking' }}
        </button>
      </footer>
    </template>
  </section>
</template>
