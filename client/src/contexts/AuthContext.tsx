import { createContext, useCallback, useContext, useEffect, useState } from 'react'
import * as authApi from '../api/auth'
import { ApiRequestError, UNAUTHORIZED_EVENT } from '../api/client'

export interface AuthUser {
  id?: number
  email: string
  firstName: string
  lastName: string
  role: 'student' | 'staff' | 'maintainer' | 'admin' | 'pending'
  mustChangePw: boolean
}

interface AuthContextValue {
  user: AuthUser | null
  isLoading: boolean
  login(email: string, password: string): Promise<void>
  logout(): Promise<void>
  refreshUser(): Promise<void>
}

const SESSION_KEY = 'ticketa_user'

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  // Restore session from localStorage on mount. The session cookie has a
  // 7-day Max-Age and survives browser restarts, so the display metadata must
  // too — sessionStorage would make the user appear logged out while their
  // cookie is still valid.
  useEffect(() => {
    try {
      let stored = localStorage.getItem(SESSION_KEY)
      // Migrate sessions persisted by older builds into sessionStorage
      if (!stored) {
        stored = sessionStorage.getItem(SESSION_KEY)
        if (stored) {
          localStorage.setItem(SESSION_KEY, stored)
          sessionStorage.removeItem(SESSION_KEY)
        }
      }
      if (stored) {
        setUser(JSON.parse(stored) as AuthUser)
      }
    } catch {
      localStorage.removeItem(SESSION_KEY)
      sessionStorage.removeItem(SESSION_KEY)
    } finally {
      setIsLoading(false)
    }
  }, [])

  // The server can invalidate the session at any moment (expiry, account
  // deactivation). When any API call gets a 401, drop the stale client-side
  // session so route guards redirect to the login page.
  useEffect(() => {
    const onUnauthorized = () => {
      localStorage.removeItem(SESSION_KEY)
      setUser(null)
    }
    window.addEventListener(UNAUTHORIZED_EVENT, onUnauthorized)
    return () => window.removeEventListener(UNAUTHORIZED_EVENT, onUnauthorized)
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    // Server validates credentials, sets the HttpOnly session cookie and
    // returns the user profile — no need for a separate /api/me call.
    const apiUser = await authApi.login(email.trim().toLowerCase(), password)
    const newUser: AuthUser = {
      id: apiUser.id,
      email: apiUser.email,
      firstName: apiUser.first_name,
      lastName: apiUser.last_name,
      role: apiUser.user_type,
      mustChangePw: apiUser.must_change_pw,
    }
    localStorage.setItem(SESSION_KEY, JSON.stringify(newUser))
    setUser(newUser)
  }, [])

  // refreshUser aktualizuje metadata přihlášeného uživatele z /api/me.
  // Volá se po změně hesla, aby se vymazal flag mustChangePw.
  const refreshUser = useCallback(async () => {
    try {
      const apiUser = await authApi.me()
      const updated: AuthUser = {
        id: apiUser.id,
        email: apiUser.email,
        firstName: apiUser.first_name,
        lastName: apiUser.last_name,
        role: apiUser.user_type,
        mustChangePw: apiUser.must_change_pw,
      }
      localStorage.setItem(SESSION_KEY, JSON.stringify(updated))
      setUser(updated)
    } catch {
      // Pokud /api/me selže, necháme aktuální stav beze změny
    }
  }, [])

  const logout = useCallback(async () => {
    try {
      // Invalidate the session server-side so the cookie cannot be reused.
      await authApi.logout()
    } finally {
      // Always clear local state even if the server request fails.
      localStorage.removeItem(SESSION_KEY)
      setUser(null)
    }
  }, [])

  return (
    <AuthContext.Provider value={{ user, isLoading, login, logout, refreshUser }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside <AuthProvider>')
  return ctx
}

export { ApiRequestError }
