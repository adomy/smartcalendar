import { useEffect } from 'react'
import { Navigate, Route, Routes, useLocation } from 'react-router-dom'
import { getProfile } from './api'
import AppLayout from './components/AppLayout'
import AdminUsers from './pages/AdminUsers'
import CalendarPage from './pages/Calendar'
import Login from './pages/Login'
import OperationLogs from './pages/OperationLogs'
import Profile from './pages/Profile'
import Register from './pages/Register'
import { useAuthStore } from './store/auth'
import './App.css'

const ProtectedLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { token } = useAuthStore()
  if (!token) {
    return <Navigate to="/login" replace />
  }
  return <AppLayout>{children}</AppLayout>
}

const AdminRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { user } = useAuthStore()
  if (user?.role !== 'admin') {
    return <Navigate to="/calendar" replace />
  }
  return <>{children}</>
}

function App() {
  const { token, user, setUser } = useAuthStore()
  const location = useLocation()

  useEffect(() => {
    if (token && !user) {
      getProfile().then((response) => {
        if (response.data.code === 0) {
          setUser(response.data.data)
        }
      })
    }
  }, [token, user, setUser])

  if (!token && location.pathname === '/') {
    return <Navigate to="/login" replace />
  }

  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route
        path="/calendar"
        element={
          <ProtectedLayout>
            <CalendarPage />
          </ProtectedLayout>
        }
      />
      <Route
        path="/profile"
        element={
          <ProtectedLayout>
            <Profile />
          </ProtectedLayout>
        }
      />
      <Route
        path="/operation-logs"
        element={
          <ProtectedLayout>
            <OperationLogs />
          </ProtectedLayout>
        }
      />
      <Route
        path="/admin/users"
        element={
          <ProtectedLayout>
            <AdminRoute>
              <AdminUsers />
            </AdminRoute>
          </ProtectedLayout>
        }
      />
      <Route path="*" element={<Navigate to="/calendar" replace />} />
    </Routes>
  )
}

export default App
