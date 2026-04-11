import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import ProtectedRoute from './components/ProtectedRoute'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Events from './pages/Events'
import Uptime from './pages/Uptime'
import Retention from './pages/Retention'
import ABTesting from './pages/ABTesting'
import Costs from './pages/Costs'
import CSAT from './pages/CSAT'
import LandingPage from './pages/LandingPage'
import FlagAudit from './pages/FlagAudit'
import Pricing from './pages/Pricing'
import StatusPage from './pages/StatusPage'
import LoginPage from './pages/LoginPage'
import AuthCallbackPage from './pages/AuthCallbackPage'

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          {/* Rotas públicas */}
          <Route path="/landing" element={<LandingPage />} />
          <Route path="/pricing" element={<Pricing />} />
          <Route path="/status" element={<StatusPage />} />
          <Route path="/login" element={<LoginPage />} />
          {/* BKL-1115: recebe access_token + refresh_token do auth-service após login OTP */}
          <Route path="/auth/callback" element={<AuthCallbackPage />} />

          {/* Rotas protegidas — requerem JWT emitido pelo auth-service */}
          <Route
            path="/*"
            element={
              <ProtectedRoute>
                <Layout>
                  <Routes>
                    <Route path="/" element={<Dashboard />} />
                    <Route path="/uptime" element={<Uptime />} />
                    <Route path="/events" element={<Events />} />
                    <Route path="/retention" element={<Retention />} />
                    <Route path="/ab-testing" element={<ABTesting />} />
                    <Route path="/costs" element={<Costs />} />
                    <Route path="/csat" element={<CSAT />} />
                    <Route path="/flags/audit" element={<FlagAudit />} />
                    <Route path="*" element={<Navigate to="/" replace />} />
                  </Routes>
                </Layout>
              </ProtectedRoute>
            }
          />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}

export default App
