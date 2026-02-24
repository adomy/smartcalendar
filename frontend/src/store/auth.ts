import { create } from 'zustand'
import type { User } from '../types'
import { clearToken, getToken, setToken } from '../utils/auth'

type AuthState = {
  token: string
  user?: User
  setAuth: (token: string, user: User) => void
  setUser: (user: User) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: getToken(),
  user: undefined,
  setAuth: (token, user) => {
    setToken(token)
    set({ token, user })
  },
  setUser: (user) => set({ user }),
  logout: () => {
    clearToken()
    set({ token: '', user: undefined })
  },
}))
