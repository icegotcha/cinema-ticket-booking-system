import { apiRequest } from './client'

export interface AppUser {
  id: string
  email: string
}

export function loginWithBackend(email: string, idToken: string) {
  return apiRequest<AppUser>('/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, id_token: idToken }),
  })
}

export function signupWithBackend(uid: string, email: string) {
  return apiRequest<null>('/auth/signup', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ uid, email }),
  })
}
