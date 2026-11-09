import { apiRequest } from './client'

export interface AppUser {
  id: string
  email: string
}

export function loginWithBackend(idToken: string) {
  return apiRequest<AppUser>('/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ id_token: idToken }),
  })
}
