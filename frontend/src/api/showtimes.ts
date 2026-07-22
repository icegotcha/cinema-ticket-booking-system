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
}

export function useMovieShowtimesQuery(movieId: MaybeRefOrGetter<string>) {
  return useQuery({
    queryKey: computed(() => showtimeKeys.byMovie(toValue(movieId))),
    queryFn: () => apiRequest<Showtime[]>(`/showtimes/movies/${toValue(movieId)}`),
    enabled: () => Boolean(toValue(movieId)),
  })
}
