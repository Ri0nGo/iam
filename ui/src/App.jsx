import { Navigate, Route, Routes } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import AppLayout from './layouts/AppLayout'
import DashboardPage from './pages/DashboardPage'
import UserPage from './pages/UserPage'
import RolePage from './pages/RolePage'
import AuthManagementPage from './pages/AuthManagementPage'
import ProfilePage from './pages/ProfilePage'
import { useAuth } from './store/auth'

function ProtectedRoute({ children }) {
  const { token } = useAuth()
  if (!token) {
    return <Navigate to="/login" replace />
  }
  return children
}

export default function App() {
  const { token } = useAuth()

  return (
    <Routes>
      <Route path="/login" element={token ? <Navigate to="/dashboard" replace /> : <LoginPage />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <AppLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<DashboardPage />} />
        <Route path="auth-management" element={<AuthManagementPage />} />
        <Route path="users" element={<UserPage />} />
        <Route path="roles" element={<RolePage />} />
        <Route path="profile" element={<ProfilePage />} />
      </Route>
      <Route path="*" element={<Navigate to={token ? '/dashboard' : '/login'} replace />} />
    </Routes>
  )
}
