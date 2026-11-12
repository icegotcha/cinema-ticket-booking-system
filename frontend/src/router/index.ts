import { createRouter, createWebHistory } from 'vue-router'
import { hasAdminToken } from '@/api/admin'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/admin',
      name: 'admin',
      component: () => import('@/views/AdminView.vue'),
      meta: { requiresAdminSession: true, isAdminPage: true },
    },
    {
      path: '/admin/login',
      name: 'admin-login',
      component: () => import('@/views/AdminLoginView.vue'),
      meta: { isAdminPage: true },
    },
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
    },
    {
      path: '/movies/:movieId/showtimes',
      component: () => import('@/views/ShowtimesView.vue'),
      props: true,
    },
    {
      path: '/movies/:movieId/showtimes/:showtimeId/seats',
      name: 'seat-map',
      component: () => import('@/views/SeatMapView.vue'),
      props: true,
    },
    {
      path: '/movies/:movieId/showtimes',
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
  if (to.meta.requiresAdminSession && !hasAdminToken()) return { name: 'admin-login' }
  if (to.name === 'admin-login' && hasAdminToken()) return { name: 'admin' }
})

export default router
