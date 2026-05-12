import { createContext, useContext, useEffect, useMemo, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { AUTH_INVALID_EVENT } from '../lib/request'

const STORAGE_KEY = 'iam.console.auth'
const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const location = useLocation()
  const [state, setState] = useState(() => {
    const raw = localStorage.getItem(STORAGE_KEY)
    return raw ? JSON.parse(raw) : { token: '', user: null }
  })

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
  }, [state])

  useEffect(() => {
    if (state.token && !localStorage.getItem(STORAGE_KEY)) {
      setState({ token: '', user: null })
    }
  }, [location.pathname, state.token])

  useEffect(() => {
    function handleAuthInvalid() {
      setState({ token: '', user: null })
    }

    window.addEventListener(AUTH_INVALID_EVENT, handleAuthInvalid)
    return () => window.removeEventListener(AUTH_INVALID_EVENT, handleAuthInvalid)
  }, [])

  const value = useMemo(
    () => ({
      token: state.token,
      user: state.user,
      login(payload) {
        setState({ token: payload.access_token, user: payload.user })
      },
      logout() {
        setState({ token: '', user: null })
      },
      updateUser(user) {
        setState((prev) => ({ ...prev, user }))
      },
    }),
    [state],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  return useContext(AuthContext)
}
