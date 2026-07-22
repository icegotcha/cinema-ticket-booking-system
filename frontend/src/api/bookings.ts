import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import type { MaybeRefOrGetter } from 'vue'
import { computed, toValue } from 'vue'

import { apiRequest } from './client'

export interface Booking {
  id: string
  show_time_id: string
  user_id: string
  seat_number: string
}

interface CreateBookingsRequest {
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

export function useCreateBookingsMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: CreateBookingsRequest) =>
      apiRequest<Booking[]>('/bookings', {
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
