import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
    },
    {
      path: '/movies/:movieId/showtimes',
      name: 'movie-showtimes',
      component: () => import('@/views/ShowtimesView.vue'),
      props: true,
    },
    {
      path: '/movies/:movieId/showtimes/:showtimeId/seats',
      name: 'seat-map',
      component: () => import('@/views/SeatMapView.vue'),
      props: true,
    },
  ],
  scrollBehavior: () => ({ top: 0 }),
})

export default router
