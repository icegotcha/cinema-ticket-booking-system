import {
  GoogleAuthProvider,
  onAuthStateChanged,
  signInWithPopup,
  signOut as firebaseSignOut,
  type User,
} from 'firebase/auth'
import { readonly, ref } from 'vue'

import { loginWithBackend, type AppUser } from '@/api/auth'
import { firebaseAuth } from '@/firebase'

const user = ref<User | null>(null)
const profile = ref<AppUser | null>(null)
const ready = ref(false)
const modalOpen = ref(false)

const authError = ref('')
let session: { firebaseUser: User; promise: Promise<void> } | null = null

// Popup completion and the auth observer share the same synchronization attempt.
function syncBackend(firebaseUser: User): Promise<void> {
  if (session?.firebaseUser === firebaseUser) return session.promise

  user.value = null
  profile.value = null
  authError.value = ''
  ready.value = false
  const promise = (async () => {
    try {
      const token = await firebaseUser.getIdToken()
      const backendProfile = await loginWithBackend(token)
      if (firebaseAuth.currentUser !== firebaseUser) return
      profile.value = backendProfile
      user.value = firebaseUser
    } catch (error) {
      if (firebaseAuth.currentUser === firebaseUser) {
        authError.value = error instanceof Error ? error.message : 'Login failed. Please try again.'
        // Preserve the original login error even if Firebase sign-out fails.
        await firebaseSignOut(firebaseAuth).catch(() => {})
      }
      throw error
    } finally {
      if (session?.firebaseUser === firebaseUser || !firebaseAuth.currentUser) ready.value = true
    }
  })()
  session = { firebaseUser, promise }
  return promise
}

onAuthStateChanged(firebaseAuth, (firebaseUser) => {
  if (firebaseUser) {
    // The error is displayed through authError and also propagated to popup login.
    void syncBackend(firebaseUser).catch(() => {})
  } else {
    session = null
    user.value = null
    profile.value = null
    ready.value = true
  }
})

export function useAuth() {
  function openAuth() {
    authError.value = ''
    modalOpen.value = true
  }

  async function loginWithGoogle() {
    const credential = await signInWithPopup(firebaseAuth, new GoogleAuthProvider())
    await syncBackend(credential.user)
    modalOpen.value = false
  }

  async function signOut() {
    await firebaseSignOut(firebaseAuth)
    session = null
    user.value = null
    profile.value = null
  }

  return {
    user: readonly(user),
    profile: readonly(profile),
    ready: readonly(ready),
    authError: readonly(authError),
    modalOpen,
    openAuth,
    closeAuth: () => (modalOpen.value = false),
    loginWithGoogle,
    signOut,
  }
}
