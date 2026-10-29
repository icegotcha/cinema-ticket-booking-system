import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import type { MaybeRefOrGetter } from 'vue'
import { computed, toValue } from 'vue'

import { apiRequest } from './client'

export interface Booking {
  id: string
  show_time_id: string
  user_id: string
  seat_number: string
  status: string
  expires_at?: string
}

interface BookingSeatsRequest {
  show_time_id: string
  seat_numbers: string[]
}

export const bookingKeys = {
  byShowtime: (showtimeId: string) => ['bookings', 'showtime', showtimeId] as const,
}

export function useShowtimeBookingsQuery(showtimeId: MaybeRefOrGetter<string>) {
  return useQuery({
    queryKey: computed(() => bookingKeys.byShowtime(toValue(showtimeId))),
    queryFn: () =>
      apiRequest<Booking[]>(`/bookings?show_time_id=${encodeURIComponent(toValue(showtimeId))}`),
    enabled: () => Boolean(toValue(showtimeId)),
  })
}

export function useHoldBookingsMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: BookingSeatsRequest) =>
      apiRequest<Booking[]>('/bookings/holds', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(request),
      }),
    onSuccess: (bookings) => {
      const showtimeId = bookings[0]?.show_time_id
      if (showtimeId) {
        void queryClient.invalidateQueries({ queryKey: bookingKeys.byShowtime(showtimeId) })
      }
    },
  })
}

export function useReleaseBookingsMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: BookingSeatsRequest) =>
      apiRequest<Booking[]>('/bookings/holds', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(request),
      }),
    onSuccess: (_bookings, request) => {
      void queryClient.invalidateQueries({
        queryKey: bookingKeys.byShowtime(request.show_time_id),
      })
    },
  })
}

export function useConfirmBookingsMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: BookingSeatsRequest) =>
      apiRequest<Booking[]>('/bookings/confirm', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(request),
      }),
    onSuccess: (bookings) => {
      const showtimeId = bookings[0]?.show_time_id
      if (showtimeId) {
        void queryClient.invalidateQueries({ queryKey: bookingKeys.byShowtime(showtimeId) })
      }
    },
  })
}
