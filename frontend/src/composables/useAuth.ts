import {
  GoogleAuthProvider,
  onAuthStateChanged,
  signInWithPopup,
  signOut as firebaseSignOut,
  type User,
} from 'firebase/auth'
import { readonly, ref } from 'vue'

import { loginWithBackend, signupWithBackend, type AppUser } from '@/api/auth'
import { firebaseAuth } from '@/firebase'

const user = ref<User | null>(firebaseAuth.currentUser)
const profile = ref<AppUser | null>(null)
const ready = ref(false)
const modalOpen = ref(false)

async function syncBackend(firebaseUser: User, createIfMissing = false) {
  const email = firebaseUser.email
  if (!email) throw new Error('Your account does not have an email address.')

  const token = await firebaseUser.getIdToken()
  try {
    profile.value = await loginWithBackend(email, token)
  } catch (error) {
    if (!createIfMissing || !(error instanceof Error) || error.message !== 'user not found')
      throw error
    await signupWithBackend(firebaseUser.uid, email)
    profile.value = await loginWithBackend(email, token)
  }
}

onAuthStateChanged(firebaseAuth, async (firebaseUser) => {
  user.value = firebaseUser
  if (firebaseUser) {
    try {
      await syncBackend(firebaseUser, true)
    } catch {
      profile.value = null
    }
  } else {
    profile.value = null
  }
  ready.value = true
})

export function useAuth() {
  function openAuth() {
    modalOpen.value = true
  }

  async function loginWithGoogle() {
    const credential = await signInWithPopup(firebaseAuth, new GoogleAuthProvider())
    await syncBackend(credential.user, true)
    modalOpen.value = false
  }

  async function signOut() {
    await firebaseSignOut(firebaseAuth)
  }

  return {
    user: readonly(user),
    profile: readonly(profile),
    ready: readonly(ready),
    modalOpen,
    openAuth,
    closeAuth: () => (modalOpen.value = false),
    loginWithGoogle,
    signOut,
  }
}
