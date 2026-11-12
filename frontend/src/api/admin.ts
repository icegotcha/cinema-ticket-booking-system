import type { Booking } from './bookings'

export interface AdminBookings {
  page: number
  page_size: number
  total: number
  bookings: Booking[] | null
  showtimes: { id: string; movie_id: string; start_time: { T: number } }[] | null
  movies: { id: string; series_title: string }[] | null
}

const API_BASE_URL = import.meta.env.VITE_API_URL ?? '/api'
const SESSION_KEY = 'cinema.admin.session'

interface AdminSession {
  access_token: string
}

function isStoredSession(value: unknown): value is AdminSession {
  if (typeof value !== 'object' || value === null) return false
  const session = value as Partial<AdminSession>
  return typeof session.access_token === 'string' && session.access_token.length > 0
}

function restoreSession(): AdminSession | null {
  try {
    const saved: unknown = JSON.parse(sessionStorage.getItem(SESSION_KEY) ?? 'null')
    return isStoredSession(saved) ? saved : null
  } catch {
    return null
  }
}

let session = restoreSession()

function saveSession(nextSession: AdminSession | null) {
  session = nextSession
  try {
    if (nextSession) sessionStorage.setItem(SESSION_KEY, JSON.stringify(nextSession))
    else sessionStorage.removeItem(SESSION_KEY)
  } catch {
    // When browser storage is unavailable, keep the session in memory.
  }
}

export class AdminSessionExpiredError extends Error {
  constructor() {
    super('Your session expired. Please sign in again.')
  }
}

export function hasAdminToken() {
  return Boolean(session?.access_token)
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
  const nextSession: unknown = await response.json()
  if (!isStoredSession(nextSession)) throw new Error('Unable to sign in. Please try again.')
  saveSession({ access_token: nextSession.access_token })
}

export async function getAdminBookings(params: URLSearchParams): Promise<AdminBookings> {
  if (!session) throw new AdminSessionExpiredError()
  const response = await fetch(`${API_BASE_URL}/admin/bookings?${params}`, {
    headers: { Authorization: `Bearer ${session.access_token}` },
    cache: 'no-store',
  })
  if (response.status === 401) {
    saveSession(null)
    throw new AdminSessionExpiredError()
  }
  if (!response.ok) throw new Error('Unable to load bookings. Please try again.')
  return response.json() as Promise<AdminBookings>
}

export async function logoutAdmin() {
  if (!session) return
  const response = await fetch(`${API_BASE_URL}/admin/logout`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${session.access_token}` },
    cache: 'no-store',
  })
  if (!response.ok && response.status !== 401)
    throw new Error('Unable to sign out. Please try again.')
  saveSession(null)
}
