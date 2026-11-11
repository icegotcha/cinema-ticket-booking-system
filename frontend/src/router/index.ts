import { createRouter, createWebHistory } from 'vue-router'
import { hasAdminSession } from '@/api/admin'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/admin',
      name: 'admin',
      component: () => import('@/views/AdminView.vue'),
      meta: { requiresAdmin: true },
    },
    {
      path: '/admin/login',
      name: 'admin-login',
      component: () => import('@/views/AdminLoginView.vue'),
    },
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

router.beforeEach((to) => {
  if (to.meta.requiresAdmin && !hasAdminSession()) return { name: 'admin-login' }
  if (to.name === 'admin-login' && hasAdminSession()) return { name: 'admin' }
})

export default router
