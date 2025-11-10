import { PropsWithChildren } from 'react';
import { Link } from 'react-router-dom';
import './RootLayout.css';

const RootLayout = ({ children }: PropsWithChildren) => (
  <div className="app-shell">
    <header className="app-header">
      <h1>Detect Price by Photo</h1>
      <nav>
        <Link to="/">Home</Link>
        <Link to="/dashboard">Dashboard</Link>
        <Link to="/upload">Upload</Link>
        <Link to="/pricing">Pricing</Link>
      </nav>
    </header>
    <main className="app-main">{children}</main>
  </div>
);

export default RootLayout;
