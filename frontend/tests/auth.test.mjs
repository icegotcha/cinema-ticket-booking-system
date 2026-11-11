import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'
import vm from 'node:vm'
import ts from 'typescript'
import * as vue from 'vue'

const source = readFileSync(new URL('../src/composables/useAuth.ts', import.meta.url), 'utf8')
const compiled = ts.transpileModule(source, {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
}).outputText

function setup(login) {
  const firebaseAuth = { currentUser: null }
  let observer
  let requests = 0
  let signOuts = 0
  const firebaseUser = { uid: 'firebase-uid', getIdToken: async () => 'verified-token' }
  const modules = {
    vue,
    '@/firebase': { firebaseAuth },
    '@/api/auth': {
      loginWithBackend: async (token) => {
        requests++
        assert.equal(token, 'verified-token')
        return login()
      },
    },
    'firebase/auth': {
      GoogleAuthProvider: class {},
      onAuthStateChanged: (_, callback) => {
        observer = callback
      },
      signInWithPopup: async () => {
        firebaseAuth.currentUser = firebaseUser
        observer(firebaseUser)
        return { user: firebaseUser }
      },
      signOut: async () => {
        signOuts++
        firebaseAuth.currentUser = null
        observer(null)
      },
    },
  }
  const exports = {}
  vm.runInNewContext(compiled, {
    exports,
    Error,
    require: (name) => {
      assert.ok(name in modules, `Unexpected import ${name}`)
      return modules[name]
    },
  })
  return {
    auth: exports.useAuth(),
    restore: () => {
      firebaseAuth.currentUser = firebaseUser
      observer(firebaseUser)
    },
    requests: () => requests,
    signOuts: () => signOuts,
  }
}

const profile = { id: 'mongo-id', email: 'viewer@example.com' }
const flush = () => new Promise((resolve) => setImmediate(resolve))

test('popup and observer share one backend login and wait for its success', async () => {
  let resolveLogin
  const state = setup(
    () =>
      new Promise((resolve) => {
        resolveLogin = resolve
      }),
  )
  state.auth.openAuth()
  const pending = state.auth.loginWithGoogle()
  await flush()
  assert.equal(state.requests(), 1)
  assert.equal(state.auth.user.value, null)
  assert.equal(state.auth.modalOpen.value, true)
  resolveLogin(profile)
  await pending
  assert.equal(state.auth.user.value.uid, 'firebase-uid')
  assert.equal(state.auth.profile.value.id, 'mongo-id')
  assert.equal(state.auth.modalOpen.value, false)
  assert.equal(state.auth.ready.value, true)
})

test('backend failure signs out, stays unauthenticated, and permits retry', async () => {
  let fail = true
  const state = setup(() => {
    if (fail) throw new Error('database error')
    return profile
  })
  state.auth.openAuth()
  await assert.rejects(state.auth.loginWithGoogle(), /database error/)
  assert.equal(state.requests(), 1)
  assert.equal(state.signOuts(), 1)
  assert.equal(state.auth.user.value, null)
  assert.equal(state.auth.profile.value, null)
  assert.equal(state.auth.authError.value, 'database error')
  assert.equal(state.auth.modalOpen.value, true)
  fail = false
  await state.auth.loginWithGoogle()
  assert.equal(state.requests(), 2)
  assert.equal(state.auth.profile.value.id, 'mongo-id')
  assert.equal(state.auth.authError.value, '')
})

test('restored Firebase sessions also require backend success', async () => {
  const state = setup(() => profile)
  state.restore()
  await flush()
  assert.equal(state.requests(), 1)
  assert.equal(state.auth.user.value.uid, 'firebase-uid')
  assert.equal(state.auth.ready.value, true)
})

test('a late login response cannot restore a signed-out user', async () => {
  let resolveLogin
  const state = setup(
    () =>
      new Promise((resolve) => {
        resolveLogin = resolve
      }),
  )
  state.restore()
  await flush()
  await state.auth.signOut()
  resolveLogin(profile)
  await flush()
  assert.equal(state.auth.user.value, null)
  assert.equal(state.auth.profile.value, null)
})
