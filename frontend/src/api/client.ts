import axios from 'axios'
import { getToken, clearToken } from '../utils/auth'

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api',
  timeout: 15000,
})

apiClient.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    const response = error.response
    if (response && (response.data?.code === 40101 || response.data?.code === 40102 || response.data?.code === 40301)) {
      clearToken()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)
