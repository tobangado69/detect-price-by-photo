import { Routes, Route } from 'react-router-dom';
import Home from './pages/Home';
import Dashboard from './pages/Dashboard';
import Upload from './pages/Upload';
import Pricing from './pages/Pricing';
import RootLayout from './components/RootLayout';

const App = () => (
  <RootLayout>
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/dashboard" element={<Dashboard />} />
      <Route path="/upload" element={<Upload />} />
      <Route path="/pricing" element={<Pricing />} />
    </Routes>
  </RootLayout>
);

export default App;
