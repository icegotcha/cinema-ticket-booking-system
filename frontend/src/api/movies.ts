import { useQuery } from '@tanstack/vue-query'
import type { MaybeRefOrGetter } from 'vue'
import { computed, toValue } from 'vue'

import { apiRequest } from './client'

export interface Movie {
  id: string
  title: string
  poster_link: string
}

export const movieKeys = {
  all: ['movies'] as const,
  detail: (id: string) => ['movies', id] as const,
}

export function useMoviesQuery() {
  return useQuery({
    queryKey: movieKeys.all,
    queryFn: () => apiRequest<Movie[]>('/movies'),
  })
}

export function useMovieQuery(movieId: MaybeRefOrGetter<string>) {
  return useQuery({
    queryKey: computed(() => movieKeys.detail(toValue(movieId))),
    queryFn: () => apiRequest<Movie>(`/movies/${toValue(movieId)}`),
    enabled: () => Boolean(toValue(movieId)),
  })
}
