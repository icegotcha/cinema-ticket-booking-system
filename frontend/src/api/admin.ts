import type { Booking } from './bookings'

export interface AdminBookings {
  bookings: Booking[] | null
  showtimes: { id: string; movie_id: string; start_time: { T: number } }[] | null
  movies: { id: string; series_title: string }[] | null
}

const API_BASE_URL = import.meta.env.VITE_API_URL ?? '/api'
let accessToken = ''

export class AdminSessionExpiredError extends Error {
  constructor() {
    super('Your session expired. Please sign in again.')
  }
}

export function hasAdminSession() {
  return Boolean(accessToken)
}

export async function loginAdmin(username: string, password: string) {
  const response = await fetch(`${API_BASE_URL}/admin/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
    cache: 'no-store',
  })
  if (response.status === 401) throw new Error('Incorrect username or password. Please try again.')
  if (!response.ok) throw new Error('Unable to sign in. Please try again.')
  const session = (await response.json()) as { access_token: string }
  accessToken = session.access_token
}

export async function getAdminBookings(): Promise<AdminBookings> {
  const response = await fetch(`${API_BASE_URL}/admin/bookings`, {
    headers: { Authorization: `Bearer ${accessToken}` },
    cache: 'no-store',
  })
  if (response.status === 401) {
    accessToken = ''
    throw new AdminSessionExpiredError()
  }
  if (!response.ok) throw new Error('Unable to load bookings. Please try again.')
  return response.json() as Promise<AdminBookings>
}

export async function logoutAdmin() {
  const response = await fetch(`${API_BASE_URL}/admin/logout`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    cache: 'no-store',
  })
  if (!response.ok && response.status !== 401)
    throw new Error('Unable to sign out. Please try again.')
  accessToken = ''
}
