import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import Overview from './pages/Overview'
import Flags from './pages/Flags'
import Events from './pages/Events'
import Retention from './pages/Retention'

function App() {
  return (
    <BrowserRouter>
      <Layout>
        <Routes>
          <Route path="/" element={<Overview />} />
          <Route path="/flags" element={<Flags />} />
          <Route path="/events" element={<Events />} />
          <Route path="/retention" element={<Retention />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  )
}

export default App
