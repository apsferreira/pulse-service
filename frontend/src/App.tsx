import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
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

function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Rotas públicas — sem layout do dashboard */}
        <Route path="/landing" element={<LandingPage />} />
        <Route path="/pricing" element={<Pricing />} />
        {/* BKL-1068: status page pública dos serviços IIT */}
        <Route path="/status" element={<StatusPage />} />

        {/* Rotas autenticadas — com layout */}
        <Route
          path="/*"
          element={
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
          }
        />
      </Routes>
    </BrowserRouter>
  )
}

export default App
