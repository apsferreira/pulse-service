import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Events from './pages/Events'
import Uptime from './pages/Uptime'
import Retention from './pages/Retention'
import ABTesting from './pages/ABTesting'
import Costs from './pages/Costs'
import CSAT from './pages/CSAT'

function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/uptime" element={<Uptime />} />
          <Route path="/events" element={<Events />} />
          <Route path="/retention" element={<Retention />} />
          <Route path="/ab-testing" element={<ABTesting />} />
          <Route path="/costs" element={<Costs />} />
          <Route path="/csat" element={<CSAT />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  )
}

export default App
