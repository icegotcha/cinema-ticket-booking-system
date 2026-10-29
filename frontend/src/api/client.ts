import { firebaseAuth } from '@/firebase'

const API_BASE_URL = import.meta.env.VITE_API_URL ?? '/api'

export async function apiRequest<T>(path: string, options?: RequestInit): Promise<T> {
  const headers = new Headers(options?.headers)
  const currentUser = firebaseAuth.currentUser
  if (currentUser) headers.set('Authorization', `Bearer ${await currentUser.getIdToken()}`)
  const response = await fetch(`${API_BASE_URL}${path}`, { ...options, headers })

  if (!response.ok) {
    const body = (await response.json().catch(() => null)) as { error?: string } | null
    throw new Error(body?.error ?? `Request failed with status ${response.status}`)
  }

  return response.json() as Promise<T>
}
