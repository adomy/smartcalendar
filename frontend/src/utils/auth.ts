export const TOKEN_KEY = 'smartcalendar_token'

export const getToken = () => {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export const setToken = (token: string) => {
  localStorage.setItem(TOKEN_KEY, token)
}

export const clearToken = () => {
  localStorage.removeItem(TOKEN_KEY)
}
