import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { isAuthenticated } from './lib/auth'
import Layout from './components/Layout'
import Overview from './pages/Overview'
import Flags from './pages/Flags'
import Events from './pages/Events'
import Retention from './pages/Retention'
import Login from './pages/Login'
import AuthCallback from './pages/AuthCallback'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  if (!isAuthenticated()) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/auth/callback" element={<AuthCallback />} />
        <Route
          path="/*"
          element={
            <ProtectedRoute>
              <Layout>
                <Routes>
                  <Route path="/" element={<Overview />} />
                  <Route path="/flags" element={<Flags />} />
                  <Route path="/events" element={<Events />} />
                  <Route path="/retention" element={<Retention />} />
                  <Route path="*" element={<Navigate to="/" replace />} />
                </Routes>
              </Layout>
            </ProtectedRoute>
          }
        />
      </Routes>
    </BrowserRouter>
  )
}

export default App
