import { useQuery } from '@tanstack/vue-query'
import type { MaybeRefOrGetter } from 'vue'
import { computed, toValue } from 'vue'

import type { Movie } from './movies'
import { apiRequest } from './client'

export interface Showtime {
  id: string
  start_time: number
  movie: Movie
}

export const showtimeKeys = {
  byMovie: (movieId: string) => ['showtimes', 'movie', movieId] as const,
  detail: (movieId: string, showtimeId: string) =>
    ['showtimes', 'movie', movieId, showtimeId] as const,
}

export function useMovieShowtimesQuery(movieId: MaybeRefOrGetter<string>) {
  return useQuery({
    queryKey: computed(() => showtimeKeys.byMovie(toValue(movieId))),
    queryFn: () => apiRequest<Showtime[]>(`/showtimes/movies/${toValue(movieId)}`),
    enabled: () => Boolean(toValue(movieId)),
  })
}

export function useShowtimeQuery(
  movieId: MaybeRefOrGetter<string>,
  showtimeId: MaybeRefOrGetter<string>,
) {
  return useQuery({
    queryKey: computed(() => showtimeKeys.detail(toValue(movieId), toValue(showtimeId))),
    queryFn: () =>
      apiRequest<Showtime>(`/showtimes/movies/${toValue(movieId)}/st/${toValue(showtimeId)}`),
    enabled: () => Boolean(toValue(movieId) && toValue(showtimeId)),
  })
}
